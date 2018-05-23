Terraform & Ansible
===================

Automated deployments are done using `Terraform <https://www.terraform.io/>`__ to create servers on Digital Ocean then
`Ansible <http://www.ansible.com/>`__ to create and manage testnets on those servers.

Prerequisites
-------------

-  Install `Terraform <https://www.terraform.io/downloads.html>`__ and `Ansible <http://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html>`__ on a Linux machine.
-  Create a `DigitalOcean API token <https://cloud.digitalocean.com/settings/api/tokens>`__ with read and write capability.
- Install the python dopy package (``pip install dopy``) (This is necessary for the digitalocean.py script for ansible.)
-  Create SSH keys

::

    export DO_API_TOKEN="abcdef01234567890abcdef01234567890"
    export TESTNET_NAME="remotenet"
    export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
    export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"

These will be used by both ``terraform`` and ``ansible``.

Create a remote network
-----------------------

::

    make remotenet-start


Optionally, you can set the number of servers you want to launch and the name of the testnet (which defaults to remotenet):

::

    TESTNET_NAME="mytestnet" SERVERS=7 make remotenet-start


Quickly see the /status endpoint
--------------------------------

::

    make remotenet-status


Delete servers
--------------

::

    make remotenet-stop

