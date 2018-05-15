#!/usr/bin/env bash

# XXX: this script is intended to be run from a fresh Digital Ocean droplet

# NOTE: you must set this manually now
echo "export DO_API_TOKEN=\"yourToken\"" >> ~/.profile

sudo apt-get update -y
sudo apt-get upgrade -y
sudo apt-get install -y jq unzip python-pip software-properties-common make

# get and unpack golang
curl -O https://storage.googleapis.com/golang/go1.10.linux-amd64.tar.gz
tar -xvf go1.10.linux-amd64.tar.gz

## move binary and add to path
mv go /usr/local
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile

## create the goApps directory, set GOPATH, and put it on PATH
mkdir goApps
echo "export GOPATH=/root/goApps" >> ~/.profile
echo "export PATH=\$PATH:\$GOPATH/bin" >> ~/.profile

source ~/.profile

## get the code and move into repo
REPO=github.com/cosmos/cosmos-sdk
go get $REPO
cd $GOPATH/src/$REPO

## build
git checkout zach/ansible2
make get_tools
make get_vendor_deps
make build

# generate an ssh key
ssh-keygen -f $HOME/.ssh/id_rsa -t rsa -N ''
echo "export SSH_KEY_FILE=\"\$HOME/.ssh/id_rsa.pub\"" >> ~/.profile
source ~/.profile

# install terraform
wget https://releases.hashicorp.com/terraform/0.11.7/terraform_0.11.7_linux_amd64.zip
unzip terraform_0.11.7_linux_amd64.zip -d /usr/bin/

# install ansible
sudo apt-get update -y
sudo apt-add-repository ppa:ansible/ansible -y
sudo apt-get update -y
sudo apt-get install ansible -y

# required by ansible
pip install dopy

# the next two commands are directory sensitive
cd $GOPATH/src/github.com/cosmos/cosmos-sdk/networks/remote/terraform

terraform init
terraform apply -var DO_API_TOKEN="$DO_API_TOKEN" -var SSH_KEY_FILE="$SSH_KEY_FILE" -auto-approve

# let the droplets boot
sleep 60

# get the IPs
ip0=`terraform output -json public_ips | jq '.value[0]'`
ip1=`terraform output -json public_ips | jq '.value[1]'`
ip2=`terraform output -json public_ips | jq '.value[2]'`
ip3=`terraform output -json public_ips | jq '.value[3]'`

# to remove quotes
strip() {
  opt=$1
  temp="${opt%\"}"
  temp="${temp#\"}"
  echo $temp
}

ip0=$(strip $ip0)
ip1=$(strip $ip1)
ip2=$(strip $ip2)
ip3=$(strip $ip3)

# all the ansible commands are also directory specific
cd $GOPATH/src/github.com/cosmos/cosmos-sdk/networks/remote/ansible

ansible-playbook -i inventory/digital_ocean.py -l sentrynet install.yml
ansible-playbook -i inventory/digital_ocean.py -l sentrynet config.yml -e BINARY=$GOPATH/src/github.com/cosmos/cosmos-sdk/build/gaiad -e CONFIGDIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/examples/testnet

sleep 10

# get each nodes ID then populate the ansible file
id0=`curl $ip0:46657/status | jq .result.node_info.id`
id1=`curl $ip1:46657/status | jq .result.node_info.id`
id2=`curl $ip2:46657/status | jq .result.node_info.id`
id3=`curl $ip3:46657/status | jq .result.node_info.id`

id0=$(strip $id0)
id1=$(strip $id1)
id2=$(strip $id2)
id3=$(strip $id3)

# remove file we'll re-write to with new info
old_ansible_file=$GOPATH/src/github.com/cosmos/cosmos-sdk/networks/remote/ansible/roles/install/templates/systemd.service.j2
rm $old_ansible_file

# need to populate the `--p2p.persistent_peers` flag
echo "[Unit]
Description={{service}}
Requires=network-online.target
After=network-online.target

[Service]
Restart=on-failure
User={{service}}
Group={{service}}
PermissionsStartOnly=true
ExecStart=/usr/bin/gaiad start --p2p.persistent_peers=$id0@$ip0:46656,$id1@$ip1:46656,$id2@$ip2:46656,$id3@$ip3:46656
ExecReload=/bin/kill -HUP \$MAINPID
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
" >> $old_ansible_file

# now, we can re-run the install command
ansible-playbook -i inventory/digital_ocean.py -l sentrynet install.yml

# and finally restart it all
ansible-playbook -i inventory/digital_ocean.py -l sentrynet restart.yml

echo "congratulations, your testnet is now running :)"
