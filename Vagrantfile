# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/trusty64"

  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end

  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    apt-get install -y --no-install-recommends wget curl jq shellcheck bsdmainutils psmisc zip

    wget -qO- https://get.docker.com/ | sh
    usermod -a -G docker vagrant
    apt-get autoremove -y

    curl -O https://storage.googleapis.com/golang/go1.8.linux-amd64.tar.gz
    tar -xvf go1.8.linux-amd64.tar.gz
    rm -rf /usr/local/go
    mv go /usr/local
    rm -f go1.8.linux-amd64.tar.gz
    mkdir -p /home/vagrant/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin:/home/vagrant/go/bin' >> /home/vagrant/.bash_profile
    echo 'export GOPATH=/home/vagrant/go' >> /home/vagrant/.bash_profile

    echo 'export LC_ALL=en_US.UTF-8' >> /home/vagrant/.bash_profile

    mkdir -p /home/vagrant/go/src/github.com/tendermint
    ln -s /vagrant /home/vagrant/go/src/github.com/tendermint/basecoin

    chown -R vagrant:vagrant /home/vagrant/go

    rm -rf /home/vagrant/go/src/github.com/tendermint/basecoin/vendor
  SHELL
end
