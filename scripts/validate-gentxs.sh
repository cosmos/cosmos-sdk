#!/usr/bin/env bash

DAEMON_HOME="/tmp/regen$(date +%s)"
RANDOM_KEY="randomvalidatorkey"

#############################################
# Ensure to set the below ENV settings  #
#############################################

# DAEMON=regen
# CHAIN_ID=testnet-1
# DENOM=ustake
# GH_URL=github.com/cosmos/cosmos-sdk
# BINARY_VERSION=v0.43.0-beta1
# GO_VERSION=1.15.2
# PRELAUNCH_GENESIS_URL=https://raw.githubusercontent.com/cosmos/cosmos-sdk/master/$CHAIN_ID/genesis-prelaunch.json
# GENTXS_DIR=./$CHAIN_ID/gentxs

command_exists () {
    type "$1" &> /dev/null ;
}

if command_exists go ; then
    echo "Golang is already installed"
else
  echo "Install dependencies"
  sudo apt update
  sudo apt install build-essential -y

  wget https://dl.google.com/go/go$GO_VERSION.linux-amd64.tar.gz
  tar -xvf go$GO_VERSION.linux-amd64.tar.gz
  sudo mv go /usr/local

  echo "" >> ~/.profile
  echo 'export GOPATH=$HOME/go' >> ~/.profile
  echo 'export GOROOT=/usr/local/go' >> ~/.profile
  echo 'export GOBIN=$GOPATH/bin' >> ~/.profile
  echo 'export PATH=$PATH:/usr/local/go/bin:$GOBIN' >> ~/.profile

  . ~/.profile

  go version
fi

if [ "$(ls -A $GENTXS_DIR)" ]; then
    for GENTX_FILE in *; do 
        if [ -f "$GENTX_FILE" ]; then
            set -e

            echo "GentxFile::::"
            echo $GENTX_FILE

            echo "...........Init a testnet.............."
            go get $GH_URL
            cd ~/go/src/$GH_URL
            git fetch && git checkout $BINARY_VERSION
            make build
            chmod +x ./build/$DAEMON

            ./build/$DAEMON keys add $RANDOM_KEY --keyring-backend test --home $DAEMON_HOME

            ./build/$DAEMON init --chain-id $CHAIN_ID validator --home $DAEMON_HOME

            echo "..........Fetching genesis......."
            rm -rf $DAEMON_HOME/config/genesis.json
            curl -s $PRELAUNCH_GENESIS_URL >$DAEMON_HOME/config/genesis.json

            # this genesis time is different from original genesis time, just for validating gentx.
            sed -i '/genesis_time/c\   \"genesis_time\" : \"2021-01-01T00:00:00Z\",' $DAEMON_HOME/config/genesis.json

            GENACC=$(cat ../$GENTX_FILE | sed -n 's|.*"delegator_address":"\([^"]*\)".*|\1|p')
            denomquery=$(jq -r '.body.messages[0].value.denom' ../$GENTX_FILE)
            amountquery=$(jq -r '.body.messages[0].value.amount' ../$GENTX_FILE)

            # only allow $DENOM tokens to be bonded
            if [ $denomquery != $DENOM ]; then
                echo "invalid denomination"
                exit 1
            fi

            ./build/$DAEMON add-genesis-account $RANDOM_KEY 1000000000000000$DENOM --home $DAEMON_HOME \
                --keyring-backend test

            ./build/$DAEMON gentx $RANDOM_KEY 900000000000000$DENOM --home $DAEMON_HOME \
                --keyring-backend test --chain-id $CHAIN_ID

            cp ../$GENTX_FILE $DAEMON_HOME/config/gentx/

            echo "..........Collecting gentxs......."
            ./build/$DAEMON collect-gentxs --home $DAEMON_HOME

            ./build/$DAEMON validate-genesis --home $DAEMON_HOME

            echo "..........Starting node......."
            ./build/$DAEMON start --home $DAEMON_HOME &

            sleep 10s

            echo "...checking network status.."
            echo "if this fails, most probably the gentx with address $GENACC is invalid"
            ./build/$DAEMON status --node http://localhost:26657

            echo "...Cleaning the stuff..."
            killall $DAEMON >/dev/null 2>&1
            rm -rf $DAEMON_HOME >/dev/null 2>&1
        fi
    done
else
    echo "$GENTXS_DIR is empty, nothing to validate"
fi
