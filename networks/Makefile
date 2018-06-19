########################################
### These targets were broken out of the main Makefile to enable easy setup of testnets.
### They use a form of terraform + ansible to build full nodes in Digital Ocean or AWS.
### The shell scripts in this folder are example uses of the targets.

# Name of the testnet. Used in chain-id.
TESTNET_NAME?=remotenet

# Name of the servers grouped together for management purposes. Used in tagging the servers in the cloud.
CLUSTER_NAME?=$(TESTNET_NAME)

# Number of servers deployed in Digital Ocean.
# Number of servers to put in one availability zone in AWS.
SERVERS?=4

# Number of regions to use in AWS. One region usually contains 2-3 availability zones.
REGION_LIMIT?=1

# Path to gaiad for deployment. Must be a Linux binary.
BINARY?=$(CURDIR)/../build/gaiad

# Path to the genesis.json and config.toml files to deploy on full nodes.
GENESISFILE?=$(CURDIR)/../build/genesis.json
CONFIGFILE?=$(CURDIR)/../build/config.toml

all:
	@echo "There is no all. Only sum of the ones."


########################################
### Extract genesis.json and config.toml from a node in a cluster

extract-config-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -e TESTNET_NAME="$(TESTNET_NAME)" -e GENESISFILE="$(GENESISFILE)" -e CONFIGFILE="$(CONFIGFILE)" extract-config.yml

extract-config-aws:
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b -e TESTNET_NAME="$(TESTNET_NAME)" -e GENESISFILE="$(GENESISFILE)" -e CONFIGFILE="$(CONFIGFILE)" extract-config.yml



########################################
### Remote validator nodes using terraform and ansible in Digital Ocean

validators-start-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/terraform && terraform init && (terraform workspace new "$(CLUSTER_NAME)" || terraform workspace select "$(CLUSTER_NAME)") && terraform apply -auto-approve -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" -var TESTNET_NAME="$(CLUSTER_NAME)" -var SERVERS="$(SERVERS)"
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root -e BINARY=$(BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" setup-validators.yml
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root start.yml

validators-stop-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	cd remote/terraform && terraform workspace select "$(CLUSTER_NAME)" && terraform destroy -force -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" && terraform workspace select default && terraform workspace delete "$(CLUSTER_NAME)"
	rm -rf remote/ansible/keys/

validators-status-do:
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" status.yml


########################################
### Remote validator nodes using terraform and ansible in AWS

validators-start-aws:
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/terraform-aws && terraform init && (terraform workspace new "$(CLUSTER_NAME)" || terraform workspace select "$(CLUSTER_NAME)") && terraform apply -auto-approve -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" -var TESTNET_NAME="$(CLUSTER_NAME)" -var SERVERS="$(SERVERS)" -var REGION_LIMIT="$(REGION_LIMIT)"
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b -e BINARY=$(BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" setup-validators.yml
	cd remote/ansible && ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b start.yml

validators-stop-aws:
	cd remote/terraform-aws && terraform workspace select "$(CLUSTER_NAME)" && terraform destroy -force -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" && terraform workspace select default && terraform workspace delete "$(CLUSTER_NAME)"
	rm -rf remote/ansible/keys/ remote/ansible/files/

validators-status-aws:
	cd remote/ansible && ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" status.yml

#validators-clear-aws:
#	cd remote/ansible && ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b clear-config.yml


########################################
### Remote full nodes using terraform and ansible in Digital Ocean

fullnodes-start-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/terraform && terraform init && (terraform workspace new "$(CLUSTER_NAME)" || terraform workspace select "$(CLUSTER_NAME)") && terraform apply -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" -var TESTNET_NAME="$(CLUSTER_NAME)" -var SERVERS="$(SERVERS)"
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -e BINARY=$(BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" -e GENESISFILE="$(GENESISFILE)" -e CONFIGFILE="$(CONFIGFILE)" setup-fullnodes.yml
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root start.yml

fullnodes-stop-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	cd remote/terraform && terraform workspace select "$(CLUSTER_NAME)" && terraform destroy -force -var DO_API_TOKEN="$(DO_API_TOKEN)" -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" && terraform workspace select default && terraform workspace delete "$(CLUSTER_NAME)"

fullnodes-status-do:
	cd remote/ansible && ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" status.yml


########################################
### Remote full nodes using terraform and ansible in Amazon AWS

fullnodes-start-aws:
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/terraform-aws && terraform init && (terraform workspace new "$(CLUSTER_NAME)" || terraform workspace select "$(CLUSTER_NAME)") && terraform apply -auto-approve -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" -var TESTNET_NAME="$(CLUSTER_NAME)" -var SERVERS="$(SERVERS)" -var REGION_LIMIT="$(REGION_LIMIT)"
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b -e BINARY=$(BINARY) -e TESTNET_NAME="$(TESTNET_NAME)" -e GENESISFILE="$(GENESISFILE)" -e CONFIGFILE="$(CONFIGFILE)" setup-fullnodes.yml
	cd remote/ansible && ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b start.yml

fullnodes-stop-aws:
	cd remote/terraform-aws && terraform workspace select "$(CLUSTER_NAME)" && terraform destroy -force -var SSH_PUBLIC_FILE="$(HOME)/.ssh/id_rsa.pub" -var SSH_PRIVATE_FILE="$(HOME)/.ssh/id_rsa" && terraform workspace select default && terraform workspace delete "$(CLUSTER_NAME)"
	rm -rf remote/ansible/keys/ remote/ansible/files/

fullnodes-status-aws:
	cd remote/ansible && ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" status.yml

########################################
### Other calls

upgrade-gaiad-aws:
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b -e BINARY=$(BINARY) upgrade-gaiad.yml

upgrade-gaiad-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if ! [ -f $(HOME)/.ssh/id_rsa.pub ]; then ssh-keygen ; fi
	@if [ -z "`file $(BINARY) | grep 'ELF 64-bit'`" ]; then echo "Please build a linux binary using 'make build-linux'." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -e BINARY=$(BINARY) upgrade-gaiad.yml

list-aws:
	remote/ansible/inventory/ec2.py | python -c 'import json,sys ; print "\n".join(json.loads("".join(sys.stdin.readlines()))["tag_Environment_$(CLUSTER_NAME)"])'

list-do:
	remote/ansible/inventory/digital_ocean.py | python -c 'import json,sys ; print "\n".join(json.loads("".join(sys.stdin.readlines()))["$(CLUSTER_NAME)"]["hosts"])'

install-datadog-aws:
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	@if [ -z "$(DD_API_KEY)" ]; then echo "DD_API_KEY environment variable not set." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b -e DD_API_KEY="$(DD_API_KEY)" -e TESTNET_NAME="$(TESTNET_NAME)" -e CLUSTER_NAME="$(CLUSTER_NAME)" install-datadog-agent.yml

install-datadog-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	@if [ -z "$(DD_API_KEY)" ]; then echo "DD_API_KEY environment variable not set." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root -e DD_API_KEY="$(DD_API_KEY)" -e TESTNET_NAME=$(TESTNET_NAME) -e CLUSTER_NAME=$(CLUSTER_NAME) install-datadog-agent.yml

remove-datadog-aws:
	#Make sure you have AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY or your IAM roles set for AWS API access.
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/ec2.py -l "tag_Environment_$(CLUSTER_NAME)" -u centos -b remove-datadog-agent.yml

remove-datadog-do:
	@if [ -z "$(DO_API_TOKEN)" ]; then echo "DO_API_TOKEN environment variable not set." ; false ; fi
	cd remote/ansible && ANSIBLE_HOST_KEY_CHECKING=False ansible-playbook -i inventory/digital_ocean.py -l "$(CLUSTER_NAME)" -u root remove-datadog-agent.yml

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: all extract-config-do extract-config-aws validators-start-do validators-stop-do validators-status-do validators-start-aws validators-stop-aws validators-status-aws fullnodes-start-do fullnodes-stop-do fullnodes-status-do fullnodes-start-aws fullnodes-stop-aws fullnodes-status-aws upgrade-gaiad-aws upgrade-gaiad-do list-aws list-do install-datadog-aws install-datadog-do remove-datadog-aws remove-datadog-do

