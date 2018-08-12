# Terraform & Ansible

WARNING: The Digital Ocean scripts are obsolete. They are here because they might still be useful for developers.

Automated deployments are done using [Terraform](https://www.terraform.io/) to create servers on Digital Ocean then
[Ansible](http://www.ansible.com/) to create and manage testnets on those servers.

## Prerequisites

- Install [Terraform](https://www.terraform.io/downloads.html) and [Ansible](http://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html) on a Linux machine.
- Create a [DigitalOcean API token](https://cloud.digitalocean.com/settings/api/tokens) with read and write capability.
- Install the python dopy package (`pip install dopy`) (This is necessary for the digitalocean.py script for ansible.)
- Create SSH keys

```
export DO_API_TOKEN="abcdef01234567890abcdef01234567890"
export TESTNET_NAME="remotenet"
export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"
```

These will be used by both `terraform` and `ansible`.

## Create a remote network

```
make remotenet-start
```

Optionally, you can set the number of servers you want to launch and the name of the testnet (which defaults to remotenet):

```
TESTNET_NAME="mytestnet" SERVERS=7 make remotenet-start
```

## Quickly see the /status endpoint

```
make remotenet-status
```

## Delete servers

```
make remotenet-stop
```

## Logging

You can ship logs to Logz.io, an Elastic stack (Elastic search, Logstash and Kibana) service provider. You can set up your nodes to log there automatically. Create an account and get your API key from the notes on [this page](https://app.logz.io/#/dashboard/data-sources/Filebeat), then:

```
yum install systemd-devel || echo "This will only work on RHEL-based systems."
apt-get install libsystemd-dev || echo "This will only work on Debian-based systems."

go get github.com/mheese/journalbeat
ansible-playbook -i inventory/digital_ocean.py -l remotenet logzio.yml -e LOGZIO_TOKEN=ABCDEFGHIJKLMNOPQRSTUVWXYZ012345
```
