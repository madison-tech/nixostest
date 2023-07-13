package main

import (
	"context"
	"log"
	"os"

	"github.com/pulumi/pulumi-digitalocean/sdk/v4/go/digitalocean"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func getUserData() pulumi.StringPtrInput {
	// In this function, we only read a file and return it as a Pulumi specific string.
	// We can do so much more here, by using the golang template module (https://pkg.go.dev/text/template)
	configFile := "nixos.yml"
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("cannot read config file: %s\n", err.Error())
	}
	return pulumi.String(data)
}

func main() {
	// We check if there is a command line arg that says "destroy"
	// If so, we make note of this and are aware that this specific execution means
	// that we have to destroy our stack and not create it. See below.
	destroy := false
	if len(os.Args[1:]) > 0 {
		if os.Args[1] == "destroy" {
			destroy = true
		}
	}

	// This function will deploy the stack on Digital Ocean
	deployFunc := func(ctx *pulumi.Context) error {
		// In my personal DO setup, I have added 3 SSH keys which
		// will be automatically added to the root account of my DO
		// droplet. You will have to add an ssh key on the DO dashboard
		// and remember its name. For me, I use the key for my laptop.
		leonovSshKey, err := digitalocean.LookupSshKey(ctx, &digitalocean.LookupSshKeyArgs{
			Name: os.Getenv("SSH_HOST"),
		}, nil)
		if err != nil {
			log.Println(err.Error())
			return err
		}

		// Here, we create the Droplet. We give it the following details:
		// -- Name of the droplet is "dockerhost-{id}" (the "id" is added by Pulumi)
		// -- The OS image to use "ubuntu-22-10-x64"
		// -- The size of the droplet "s-2vcpu-2gb-amd" (this costs $21 per month)
		// -- The region we want to run the droplet in. I chose Singapore "sgp1"
		// -- We specify which ssh key to use
		// -- Then we specify the Userdata. This is a set of commands that gets run
		//    when the droplet gets provisioned for the first time. More info here:
		//    https://www.digitalocean.com/community/tutorials/how-to-use-cloud-config-for-your-initial-server-setup
		droplet, err := digitalocean.NewDroplet(ctx, "temp-drop", &digitalocean.DropletArgs{
			Image:    pulumi.String("ubuntu-22-10-x64"),
			Size:     pulumi.String("s-2vcpu-2gb-amd"),
			Region:   pulumi.String("sgp1"),
			SshKeys:  pulumi.StringArray{pulumi.String(leonovSshKey.Fingerprint)},
			UserData: getUserData(),
		})
		if err != nil {
			return err
		}
		// Pulumi allows you to send back information to the caller after the stack is provisioned.
		// So here, we send back the IP address and Name of the DO Droplet that was just created.
		ctx.Export("dropletIP", droplet.Ipv4Address)
		ctx.Export("dropletName", droplet.Name)

		return nil
	}

	// Now we do a set of tasks that is usually handled by Pulumi when you do "pulumi up". We do it
	// this way because we want to only run "go run main.go" instead of doing "pulumi up"

	// Create a new context
	ctx := context.Background()
	// We use the same project name that we used when we ran "pulumi new --dir nixostest"
	projectName := "nixostest"
	// We also specified the "dev" stack.
	stackName := "dev"
	// Now we insert this project and stack into the Pulumi database. You will be able to see it on your Pulumi
	// dashboard: https://app.pulumi.com/sheran/projects
	s, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, deployFunc)
	if err != nil {
		log.Fatal(err)
	}

	// We get the workspace from our stack so that we can install the DigitalOcean plugin.
	// We need the digital ocean plugin so that we can set our DO personal access token which allows us to
	// programatically provision resources on DO.
	// Get your DO token by visiting logging in to DO and visiting: https://cloud.digitalocean.com/account/api/tokens
	w := s.Workspace()

	err = w.InstallPlugin(ctx, "digitalocean", "v4")
	if err != nil {
		log.Fatalf("failed to install DO plugin: %s\n", err.Error())
	}

	// Here we set the DO Personal Access Token. We use Env vars to do that, so make sure when you
	// run this program you run it like this: "DIGITALOCEAN_TOKEN={my token} go run main.go"
	s.SetConfig(ctx, "digitalocean:token", auto.ConfigValue{Value: os.Getenv("DIGITALOCEAN_TOKEN")})

	// After we prep the config like this we need to refresh the stack
	_, err = s.Refresh(ctx)
	if err != nil {
		log.Fatalf("failed to refresh the stack: %s\n", err.Error())
	}

	// We have to decide if we are creating or destroying our stack at this point.
	// We use a standard command line arg to check for this above. Here we check and if
	// we find the arg "destroy", then we destroy the stack
	if destroy {
		log.Println("destroying stack...")
		_, err := s.Destroy(ctx)
		if err != nil {
			log.Fatalf("failed to destroy stack: %s\n", err.Error())
		}
		log.Println("stack destroyed")
		// After destroying the stack, we don't have to keep executing so we quit.
		os.Exit(0)
	}

	// Finally, we bring up our stack
	log.Println("bringing up the stack...")
	res, err := s.Up(ctx)
	if err != nil {
		log.Fatalf("failed to bring up the stack: %s\n", err.Error())
	}
	// Once we create the stack, we check for our exports. In this case, we get the droplet name
	// and its IP address. We can either use this further in our program to provision more things
	// for example if we were setting up a DB server, then we can give this DB server IP to another stack
	// which connects a PHP website to it.
	// In our case, we just write it to the console.
	log.Printf("Done, droplet IP %s\n", res.Outputs["dropletIP"].Value.(string))
	log.Printf("droplet Name %s\n", res.Outputs["dropletName"].Value.(string))
}
