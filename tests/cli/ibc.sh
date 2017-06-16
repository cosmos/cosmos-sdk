#!/bin/bash

#!/bin/bash

# these are two globals to control all scripts (can use eg. counter instead)
SERVER_EXE=basecoin
CLIENT_EXE=basecli
# just uncomment this line for full stack traces in error output
# CLIENT_EXE="basecli --trace"

oneTimeSetUp() {
  # these are passed in as args
  BASE_DIR_1=$HOME/.basecoin_test_ibc/chain1
  CHAIN_ID_1=test-chain-1
  CLIENT_1=${BASE_DIR_1}/client
  PORT_1=1234

  BASE_DIR_2=$HOME/.basecoin_test_ibc/chain2
  CHAIN_ID_2=test-chain-2
  CLIENT_2=${BASE_DIR_2}/client
  PORT_2=2345

  # clean up and create the test dirs
  rm -rf $BASE_DIR_1 $BASE_DIR_2 2>/dev/null
  mkdir -p $BASE_DIR_1 $BASE_DIR_2

  # set up client for chain 1- make sure you use the proper prefix if you set
  # a custom CLIENT_EXE
  BC_HOME=${CLIENT_1} prepareClient
  BC_HOME=${CLIENT_2} prepareClient

  # start basecoin server, giving money to the key in the first client
  BC_HOME=${CLIENT_1} initServer $BASE_DIR_1 $CHAIN_ID_1 $PORT_1
  PID_SERVER_1=$!

  # start second basecoin server, giving money to the key in the second client
  BC_HOME=${CLIENT_2} initServer $BASE_DIR_2 $CHAIN_ID_2 $PORT_2
  PID_SERVER_2=$!

  # connect both clients
  BC_HOME=${CLIENT_1} initClient $CHAIN_ID_1 $PORT_1
  BC_HOME=${CLIENT_2} initClient $CHAIN_ID_2 $PORT_2

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
  SENDER_1=$(BC_HOME=${CLIENT_1} getAddr $RICH)
  RECV_1=$(BC_HOME=${CLIENT_1} getAddr $POOR)
  export BC_HOME=${CLIENT_1}

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

test01SendIBCTx() {
  # trigger a cross-chain sendTx... from RICH on chain1 to POOR on chain2
  # we make sure the money was reduced, but nothing arrived
  SENDER=$(BC_HOME=${CLIENT_1} getAddr $RICH)
  RECV=$(BC_HOME=${CLIENT_2} getAddr $POOR)

  # we have to remove the password request from stdout, to just get the json
  export BC_HOME=${CLIENT_1}
  RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=20002mycoin --sequence=1 --to=${CHAIN_ID_2}/${RECV} --name=$RICH 2>/dev/null)
  txSucceeded $? "$RES"
  # an example to quit early if there is no point in more tests
  if [ $? != 0 ]; then echo "aborting!"; return 1; fi

  TX=`echo $RES | cut -d: -f2-`
  HASH=$(echo $TX | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $TX | jq .height)

  # make sure balance went down and tx is indexed
  checkAccount $SENDER "1" "9007199254720990"
  checkSendTx $HASH $TX_HEIGHT $SENDER "20002"

  # # make sure nothing arrived - yet
  waitForBlock ${PORT_1}7
  assertFalse "no relay running" "BC_HOME=${CLIENT_2} ${CLIENT_EXE} query account $RECV"

  # start the relay and wait a few blocks...

  # then make sure the money arrived
}


# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory

# load common helpers
. $DIR/common.sh

. $DIR/shunit2
