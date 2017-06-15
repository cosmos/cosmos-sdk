#!/bin/bash

oneTimeSetUp() {
  BASE_DIR=$HOME/.basecoin_test_basictx
  LOG=$BASE_DIR/test.log
  SERVER_LOG=$BASE_DIR/basecoin.log

  rm -rf $BASE_DIR
  mkdir -p $BASE_DIR

  ACCOUNTS=(jae ethan bucky rigel igor)
  RICH=${ACCOUNTS[0]}
  POOR=${ACCOUNTS[1]}

  # set up client
  prepareClient

  # start basecoin server (with counter)
  initServer
  sleep 5
  PID_SERVER=$!
  echo pid $PID_SERVER

  initClient

  echo "...Testing may begin!"
  echo
  echo
  echo
}

oneTimeTearDown() {
  echo "stopping basecoin test server"
  kill -9 $PID_SERVER
  sleep 1
}

prepareClient() {
  echo "Preparing client keys..."
  export BC_HOME=$BASE_DIR/client
  basecli reset_all
  assertTrue $?

  for i in "${!ACCOUNTS[@]}"; do
      newKey ${ACCOUNTS[$i]}
  done
}

initServer() {
  echo "Setting up genesis..."
  SERVE_DIR=$BASE_DIR/server
  rm -rf $SERVE_DIR 2>/dev/null
  basecoin init --home=$SERVE_DIR >>$SERVER_LOG

  #change the genesis to the first account
  GENKEY=$(basecli keys get ${RICH} -o json | jq .pubkey.data)
  GENJSON=$(cat $SERVE_DIR/genesis.json)
  echo $GENJSON | jq '.app_options.accounts[0].pub_key.data='$GENKEY > $SERVE_DIR/genesis.json

  echo "Starting server..."
  basecoin start --home=$SERVE_DIR >>$SERVER_LOG 2>&1 &
}

initClient() {
  echo "Attaching client..."
  # hard-code the expected validator hash
  basecli init --chainid=test_chain_id --node=tcp://localhost:46657 --valhash=EB168E17E45BAEB194D4C79067FFECF345C64DE6
  assertTrue "initialized light-client" $?
}

# newKeys makes a key for a given username, second arg optional password
newKey(){
  assertNotNull "keyname required" "$1"
  KEYPASS=${2:-qwertyuiop}
  (echo $KEYPASS; echo $KEYPASS) | basecli keys new $1 >>$LOG 2>/dev/null
  assertTrue "created $1" $?
  assertTrue "$1 doesn't exist" "basecli keys get $1"
}

# getAddr gets the address for a key name
getAddr() {
  assertNotNull "keyname required" "$1"
  RAW=$(basecli keys get $1)
  assertTrue "no key for $1" $?
  # print the addr
  echo $RAW | cut -d' ' -f2
}

testGetAccount() {
  SENDER=$(getAddr $RICH)
  RECV=$(getAddr $POOR)

  echo sender $RICH
  echo $SENDER

  echo recipient $POOR
  echo $RECV

  assertFalse "requires arg" "basecli query account"
  ACCT=$(basecli query account $SENDER)
  assertTrue "must have proper genesis account" $?
  echo $ACCT
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
