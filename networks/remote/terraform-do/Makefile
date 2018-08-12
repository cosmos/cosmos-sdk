########################################
### WARNING: The DigitalOcean scripts are deprecated. They are still here because
### they might be useful for developers.

# Name of the testnet. Used in chain-id.
TESTNET_NAME?=remotenet

# Name of the servers grouped together for management purposes. Used in tagging the servers in the cloud.
CLUSTER_NAME?=$(TESTNET_NAME)

# Number of servers deployed in Digital Ocean.
# Number of servers to put in one availability zone in AWS.
SERVERS?=1

# Path to gaiad for deployment. Must be a Linux binary.
BINARY?=$(CURDIR)/../build/gaiad

# Path to the genesis.json and config.toml files to deploy on full nodes.
GENESISFILE?=$(CURDIR)/../build/genesis.json
CONFIGFILE?=$(CURDIR)/../build/config.toml

all:
	@echo "There is no all. Only sum of the ones."


########################################
### Extract genesis.json and config.toml from a node in a cluster

extract-config:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -e TESTNET_NAME="$(TESTNET_NAME)" -e GENESISFILE="$(GENESISFILE)" -e CONFIGFILE="$(CONFIGFILE)" extract-config.yml


########################################
### Remote validator nodes using terraform and ansible in Digital Ocean

validators-start:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/terraform-do && terraform init && (terraform workspace new "$(CLUSTER_NAME)" || terraform workspace select "$(CLUSTER_NAME)") && terraform apply -auto-approve -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" -var TESTNET_NAME="$(CLUSTER_NAME)" -var SERVERS="$(SERVERS)"
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root -e BINARY=$(BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" setup-validators.yml
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root start.yml

validators-stop:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	cd remote/terraform-do && terraform workspace select "$(CLUSTER_NAME)" && terraform destroy -force -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" && terraform workspace select default && terraform workspace delete "$(CLUSTER_NAME)"
	rm -rf remote/ansible/keys/

validators-status:
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" status.yml


########################################
### Remote full nodes using terraform and ansible in Digital Ocean

fullnodes-start:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/terraform-do && terraform init && (terraform workspace new "$(CLUSTER_NAME)" || terraform workspace select "$(CLUSTER_NAME)") && terraform apply -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" -var TESTNET_NAME="$(CLUSTER_NAME)" -var SERVERS="$(SERVERS)"
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -e BINARY=$(BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" -e GENESISFILE="$(GENESISFILE)" -e CONFIGFILE="$(CONFIGFILE)" setup-fullnodes.yml
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root start.yml

fullnodes-stop:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	cd remote/terraform-do && terraform workspace select "$(CLUSTER_NAME)" && terraform destroy -force -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" && terraform workspace select default && terraform workspace delete "$(CLUSTER_NAME)"

fullnodes-status:
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" status.yml


########################################
### Other calls

upgrade-gaiad:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -e BINARY=$(BINARY) upgrade-gaiad.yml

list:
	remote/ansible/inventory/digital_ocean.py | python -c 'import json,sys ; print "\n".join(json.loads("".join(sys.stdin.readlines()))["$(CLUSTER_NAME)"]["hosts"])'

install-datadog:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if [ -z "$(DD_API_KEY)" ]; then echo "DD_API_KEY environment variable not set." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root -e DD_API_KEY="$(DD_API_KEY)" -e TESTNET_NAME=$(TESTNET_NAME) -e CLUSTER_NAME=$(CLUSTER_NAME) install-datadog-agent.yml

remove-datadog:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root remove-datadog-agent.yml


# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: all extract-config validators-start validators-stop validators-status fullnodes-start fullnodes-stop fullnodes-status upgrade-gaiad list-do install-datadog remove-datadog

