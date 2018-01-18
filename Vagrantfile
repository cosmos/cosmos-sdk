# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/xenial64"

  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end

  config.vm.provision "shell", inline: <<-SHELL

    # add docker repo
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu xenial stable"

    # and golang 1.9 support
    # official repo doesn't have race detection runtime...
    #add-apt-repository ppa:gophers/archive
    add-apt-repository ppa:longsleep/golang-backports

    # install base requirements
    apt-get update
    apt-get upgrade -y
    apt-get install -y --no-install-recommends wget curl jq \
        make shellcheck bsdmainutils psmisc golang-1.9-go docker-ce

    # needed for docker
    usermod -a -G docker ubuntu

    # cleanup
    apt-get autoremove -y

    # use "EOF" not EOF to avoid variable substitution of $PATH
    cat << "EOF" >> /home/ubuntu/.bash_profile
export PATH=$PATH:/usr/lib/go-1.9/bin:/home/ubuntu/go/bin
export GOPATH=/home/ubuntu/go
export LC_ALL=en_US.UTF-8
cd go/src/github.com/cosmos/cosmos-sdk
EOF

    mkdir -p /home/ubuntu/go/bin
    mkdir -p /home/ubuntu/go/src/github.com/cosmos
    ln -s /vagrant /home/ubuntu/go/src/github.com/cosmos/cosmos-sdk

    chown -R ubuntu:ubuntu /home/ubuntu/go
    chown ubuntu:ubuntu /home/ubuntu/.bash_profile

    su - ubuntu -c 'cd /home/ubuntu/go/src/github.com/cosmos/cosmos-sdk && make get_tools'
  SHELL
end
