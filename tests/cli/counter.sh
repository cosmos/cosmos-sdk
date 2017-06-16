#!/bin/bash

# these are two globals to control all scripts (can use eg. counter instead)
SERVER_EXE=counter
CLIENT_EXE=countercli

oneTimeSetUp() {
  # these are passed in as args
  BASE_DIR=$HOME/.basecoin_test_counter
  CHAIN_ID="counter-chain"

  rm -rf $BASE_DIR 2>/dev/null
  mkdir -p $BASE_DIR

  # set up client - make sure you use the proper prefix if you set
  # a custom CLIENT_EXE
  export BC_HOME=${BASE_DIR}/client
  prepareClient

  # start basecoin server (with counter)
  initServer $BASE_DIR $CHAIN_ID 1234
  echo pid $PID_SERVER

  initClient $CHAIN_ID 1234

  echo "...Testing may begin!"
  echo
  echo
  echo
}

oneTimeTearDown() {
  echo
  echo
  echo "stopping $SERVER_EXE test server..."
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
  # we have to remove the password request from stdout, to just get the json
  RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null | tail -n +2)
  txSucceeded "$RES"
  HASH=$(echo $RES | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $RES | jq .height)

  checkAccount $SENDER "1" "9007199254740000"
  checkAccount $RECV "0" "992"

  # make sure tx is indexed
  checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

test02GetCounter() {
  COUNT=$(${CLIENT_EXE} query counter)
  assertFalse "no default count" $?
}

# checkAccount $COUNT $BALANCE
# assumes just one coin, checks the balance of first coin in any case
checkCounter() {
  # make sure sender goes down
  ACCT=$(${CLIENT_EXE} query counter)
  assertTrue "count is set" $?
  assertEquals "proper count" "$1" $(echo $ACCT | jq .data.Counter)
  assertEquals "proper money" "$2" $(echo $ACCT | jq .data.TotalFees[0].amount)
}

test02AddCount() {
  SENDER=$(getAddr $RICH)
  assertFalse "bad password" "echo hi | ${CLIENT_EXE} tx counter --amount=1000mycoin --sequence=2 --name=${RICH} 2>/dev/null"

  # we have to remove the password request from stdout, to just get the json
  RES=$(echo qwertyuiop | ${CLIENT_EXE} tx counter --amount=10mycoin --sequence=2 --name=${RICH} --valid --countfee=5mycoin 2>/dev/null | tail -n +2)
  txSucceeded "$RES"
  HASH=$(echo $RES | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $RES | jq .height)

  checkCounter "1" "5"

  # FIXME: cannot load apptx properly.
  # Look at the stack trace
  # This cannot be fixed with the current ugly apptx structure...
  # Leave for refactoring

  # make sure tx is indexed
  # echo hash $HASH
  # TX=$(${CLIENT_EXE} query tx $HASH --trace)
  # echo tx $TX
  # if [-z assertTrue "found tx" $?]; then
  #   assertEquals "proper height" $TX_HEIGHT $(echo $TX | jq .height)
  #   assertEquals "type=app" '"app"' $(echo $TX | jq .data.type)
  #   assertEquals "proper sender" "\"$SENDER\"" $(echo $TX | jq .data.data.input.address)
  # fi
  # echo $TX
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory

# load common helpers
. $DIR/common.sh

. $DIR/shunit2
