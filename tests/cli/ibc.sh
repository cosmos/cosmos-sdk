#!/bin/bash

#!/bin/bash

# these are two globals to control all scripts (can use eg. counter instead)
SERVER_EXE=basecoin
CLIENT_EXE=basecli

oneTimeSetUp() {
  # these are passed in as args
  BASE_DIR_1=$HOME/.basecoin_test_ibc/chain1
  CHAIN_ID_1=test-chain-1
  CLIENT_1=${BASE_DIR_1}/client

  BASE_DIR_2=$HOME/.basecoin_test_ibc/chain2
  CHAIN_ID_2=test-chain-2
  CLIENT_2=${BASE_DIR_2}/client

  # clean up and create the test dirs
  rm -rf $BASE_DIR_1 $BASE_DIR_2 2>/dev/null
  mkdir -p $BASE_DIR_1 $BASE_DIR_2

  # set up client for chain 1- make sure you use the proper prefix if you set
  # a custom CLIENT_EXE
  BC_HOME=${CLIENT_1} prepareClient
  BC_HOME=${CLIENT_2} prepareClient

  # start basecoin server, giving money to the key in the first client
  BC_HOME=${CLIENT_1} initServer $BASE_DIR_1 $CHAIN_ID_1 2345
  PID_SERVER_1=$!

  # start second basecoin server, giving money to the key in the second client
  BC_HOME=${CLIENT_2} initServer $BASE_DIR_2 $CHAIN_ID_2 3456
  PID_SERVER_2=$!

  # connect both clients
  BC_HOME=${CLIENT_1} initClient $CHAIN_ID_1 2345
  BC_HOME=${CLIENT_2} initClient $CHAIN_ID_2 3456

  echo "...Testing may begin!"
  echo
  echo
  echo
}

oneTimeTearDown() {
  echo
  echo
  echo "stopping both $SERVER_EXE test servers... $PID_SERVER_1 $PID_SERVER_2"
  kill -9 $PID_SERVER_1
  kill -9 $PID_SERVER_2
  sleep 1
}

test00GetAccount() {
  export BC_HOME=${CLIENT_1}
  SENDER_1=$(getAddr $RICH)
  RECV_1=$(getAddr $POOR)

  assertFalse "requires arg" "${CLIENT_EXE} query account"
  assertFalse "has no genesis account" "${CLIENT_EXE} query account $RECV_1"
  checkAccount $SENDER_1 "0" "9007199254740992"

  export BC_HOME=${CLIENT_2}
  SENDER_2=$(getAddr $RICH)
  RECV_2=$(getAddr $POOR)

  assertFalse "requires arg" "${CLIENT_EXE} query account"
  assertFalse "has no genesis account" "${CLIENT_EXE} query account $RECV_2"
  checkAccount $SENDER_2 "0" "9007199254740992"

  # make sure that they have different addresses on both chains (they are random keys)
  assertNotEquals "sender keys must be different" "$SENDER_1" "$SENDER_2"
  assertNotEquals "recipient keys must be different" "$RECV_1" "$RECV_2"
}


# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory

# load common helpers
. $DIR/common.sh

. $DIR/shunit2
