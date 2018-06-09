Using Terraform
===============

This is a `Terraform <https://www.terraform.io/>`__ configuration that sets up DigitalOcean droplets.

Prerequisites
-------------

-  Install `HashiCorp Terraform <https://www.terraform.io>`__ on a linux machine.
-  Create a `DigitalOcean API token <https://cloud.digitalocean.com/settings/api/tokens>`__ with read and write capability.
-  Create SSH keys

Build
-----

::

    export DO_API_TOKEN="abcdef01234567890abcdef01234567890"
    export TESTNET_NAME="remotenet"
    export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"
    export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"

    terraform init
    terraform apply -var DO_API_TOKEN="$DO_API_TOKEN" -var SSH_PUBLIC_FILE="$SSH_PUBLIC_FILE" -var SSH_PRIVATE_FILE="$SSH_PRIVATE_FILE"

At the end you will get a list of IP addresses that belongs to your new droplets.

Destroy
-------

Run the below:

::

    terraform destroy -var DO_API_TOKEN="$DO_API_TOKEN" -var SSH_PUBLIC_FILE="$SSH_PUBLIC_FILE" -var SSH_PRIVATE_FILE="$SSH_PRIVATE_FILE"

Good to know
------------

The DigitalOcean API was not very reliable for me. If you find that terraform fails to install a specific server (for example cluster[2]), check
the regions variable and remove data center names that you find unreliable. The variable is at cluster/variables.tf

Example:

::

    variable "regions" {
      description = "Regions to launch in"
      type = "list"
      default = ["TOR1", "LON1"]
    }


