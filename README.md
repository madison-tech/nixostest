# nixostest
This is a proof of concept for using Go, Pulumi, and NixOS to spin up a VM on DigitalOcean. The VM will have podman installed and will pull the hello-world container which can then be run.

You need to do some setup first:
1. Create and account and install Pulumi as per the [instructions](https://www.pulumi.com/docs/install/) on their site.
2. Create an account on DigitalOcean, set it up with a payment method so you can provision droplets and then generate a [Personal Access Token](https://cloud.digitalocean.com/account/api/tokens)
3. [Create and add or add an existing SSH key](https://cloud.digitalocean.com/account/security) to your DO account so that new VMs automagically get this key added. This way you can ssh to your droplet as root without password.
4. Clone this repo
5. Login to Pulumi: `pulumi login`
6. Provision the stack.

## Provisioning
To provision the stack, you have to run:

```
$  DIGITALOCEAN_TOKEN=dop_v1_abcdefghijklmnopqrstuvwxyz123456 SSH_HOST=Leonov go run main.go
```

To destroy the stack:
```
$  DIGITALOCEAN_TOKEN=dop_v1_abcdefghijklmnopqrstuvwxyz123456 go run main.go destroy
```

## After provisioning
After the stack is provisioned, you will have to wait for [NixOS](https://nixos.org/) to be installed. This will take about 2 to 3 min. What I do is ping the IP of the droplet and wait until I see the pings drop briefly. That is when the droplet is provisioned and restarts back into NixOS. Then I can ssh to the droplet:

```
$ ssh root@{droplet.public.ip.address}
```

Then we can check to see if the hello-world image has been installed.
```
# podman images
REPOSITORY                     TAG         IMAGE ID      CREATED       SIZE
docker.io/library/hello-world  latest      9c7a54a9a43c  2 months ago  19.9 kB
```

And we can run it:
```
# podman run hello-world

Hello from Docker!
This message shows that your installation appears to be working correctly.

To generate this message, Docker took the following steps:
 1. The Docker client contacted the Docker daemon.
 2. The Docker daemon pulled the "hello-world" image from the Docker Hub.
    (amd64)
 3. The Docker daemon created a new container from that image which runs the
    executable that produces the output you are currently reading.
 4. The Docker daemon streamed that output to the Docker client, which sent it
    to your terminal.

To try something more ambitious, you can run an Ubuntu container with:
 $ docker run -it ubuntu bash

Share images, automate workflows, and more with a free Docker ID:
 https://hub.docker.com/

For more examples and ideas, visit:
 https://docs.docker.com/get-started/

#
```
