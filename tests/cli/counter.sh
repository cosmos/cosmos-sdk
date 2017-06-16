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

# blatently copied to make sure it works with counter as well
test00GetAccount() {
  SENDER=$(getAddr $RICH)
  RECV=$(getAddr $POOR)

  assertFalse "requires arg" "${CLIENT_EXE} query account"
  ACCT=$(${CLIENT_EXE} query account $SENDER)
  assertTrue "must have proper genesis account" $?
  assertEquals "no tx" "0" $(echo $ACCT | jq .data.sequence)
  assertEquals "has money" "9007199254740992" $(echo $ACCT | jq .data.coins[0].amount)

  ACCT2=$(${CLIENT_EXE} query account $RECV)
  assertFalse "has no genesis account" $?
}

# blatently copied to make sure it works with counter as well
test01SendTx() {
  SENDER=$(getAddr $RICH)
  RECV=$(getAddr $POOR)

  assertFalse "missing dest" "${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 2>/dev/null"
  assertFalse "bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null"
  # we have to remove the password request from stdout, to just get the json
  RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null | tail -n +2)
  assertTrue "sent tx" $?
  HASH=$(echo $RES | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $RES | jq .height)
  assertEquals "good check" "0" $(echo $RES | jq .check_tx.code)
  assertEquals "good deliver" "0" $(echo $RES | jq .deliver_tx.code)

  # make sure sender goes down
  ACCT=$(${CLIENT_EXE} query account $SENDER)
  assertTrue "must have genesis account" $?
  assertEquals "one tx" "1" $(echo $ACCT | jq .data.sequence)
  assertEquals "has money" "9007199254740000" $(echo $ACCT | jq .data.coins[0].amount)

  # make sure recipient goes up
  ACCT2=$(${CLIENT_EXE} query account $RECV)
  assertTrue "must have new account" $?
  assertEquals "no tx" "0" $(echo $ACCT2 | jq .data.sequence)
  assertEquals "has money" "992" $(echo $ACCT2 | jq .data.coins[0].amount)

  # make sure tx is indexed
  TX=$(${CLIENT_EXE} query tx $HASH)
  assertTrue "found tx" $?
  assertEquals "proper height" $TX_HEIGHT $(echo $TX | jq .height)
  assertEquals "type=send" '"send"' $(echo $TX | jq .data.type)
  assertEquals "proper sender" "\"$SENDER\"" $(echo $TX | jq .data.data.inputs[0].address)
  assertEquals "proper out amount" "992" $(echo $TX | jq .data.data.outputs[0].coins[0].amount)
}

test02GetCounter() {
  COUNT=$(${CLIENT_EXE} query counter)
  assertFalse "no default count" $?
}

test02AddCount() {
  SENDER=$(getAddr $RICH)
  assertFalse "bad password" "echo hi | ${CLIENT_EXE} tx counter --amount=1000mycoin --sequence=2 --name=${RICH} 2>/dev/null"

  # we have to remove the password request from stdout, to just get the json
  RES=$(echo qwertyuiop | ${CLIENT_EXE} tx counter --amount=10mycoin --sequence=2 --name=${RICH} --valid --countfee=5mycoin 2>/dev/null | tail -n +2)
  assertTrue "sent tx" $?
  HASH=$(echo $RES | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $RES | jq .height)
  assertEquals "good check" "0" $(echo $RES | jq .check_tx.code)
  assertEquals "good deliver" "0" $(echo $RES | jq .deliver_tx.code)

  # check new state
  COUNT=$(${CLIENT_EXE} query counter)
  assertTrue "count now set" $?
  assertEquals "one tx" "1" $(echo $COUNT | jq .data.Counter)
  assertEquals "has money" "5" $(echo $COUNT | jq .data.TotalFees[0].amount)

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
