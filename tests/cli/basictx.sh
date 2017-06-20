#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

oneTimeSetUp() {
  # These are passed in as args
  BASE_DIR=$HOME/.basecoin_test_basictx
  CHAIN_ID=my-test-chain

  rm -rf $BASE_DIR 2>/dev/null
  mkdir -p $BASE_DIR

  # Set up client - make sure you use the proper prefix if you set
  #   a custom CLIENT_EXE
  export BC_HOME=${BASE_DIR}/client
  prepareClient

  # start basecoin server (with counter)
  initServer $BASE_DIR $CHAIN_ID 3456
  if [ $? != 0 ]; then return 1; fi

  initClient $CHAIN_ID 34567

  printf "...Testing may begin!\n\n\n"
}

oneTimeTearDown() {
  printf "\n\nstopping $SERVER_EXE test server..."
  kill -9 $PID_SERVER >/dev/null 2>&1
  sleep 1
}

test00GetAccount() {
  SENDER=$(getAddr $RICH)
  RECV=$(getAddr $POOR)

  assertFalse "requires arg" "${CLIENT_EXE} query account"

  checkAccount $SENDER "0" "9007199254740992"

  ACCT2=$(${CLIENT_EXE} query account $RECV)
  assertFalse "has no genesis account" $?
}

test01SendTx() {
  SENDER=$(getAddr $RICH)
  RECV=$(getAddr $POOR)

  assertFalse "missing dest" "${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 2>/dev/null"
  assertFalse "bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null"
  RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null)
  txSucceeded $? "$RES"
  TX=`echo $RES | cut -d: -f2-`
  HASH=$(echo $TX | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $TX | jq .height)

  checkAccount $SENDER "1" "9007199254740000"
  checkAccount $RECV "0" "992"

  # Make sure tx is indexed
  checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
