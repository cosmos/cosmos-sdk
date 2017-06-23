#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=counter
CLIENT_EXE=countercli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

oneTimeSetUp() {
  quickSetup .basecoin_test_counter counter-chain
}

oneTimeTearDown() {
  quickTearDown
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
  TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null)
  txSucceeded $? "$TX"
  HASH=$(echo $TX | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $TX | jq .height)

  checkAccount $SENDER "1" "9007199254740000"
  checkAccount $RECV "0" "992"

  # make sure tx is indexed
  checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

test02GetCounter() {
  COUNT=$(${CLIENT_EXE} query counter)
  assertFalse "no default count" $?
}

# checkCounter $COUNT $BALANCE
# Assumes just one coin, checks the balance of first coin in any case
checkCounter() {
  # make sure sender goes down
  ACCT=$(${CLIENT_EXE} query counter)
  assertTrue "count is set" $?
  assertEquals "proper count" "$1" $(echo $ACCT | jq .data.Counter)
  assertEquals "proper money" "$2" $(echo $ACCT | jq .data.TotalFees[0].amount)
}

test03AddCount() {
  SENDER=$(getAddr $RICH)
  assertFalse "bad password" "echo hi | ${CLIENT_EXE} tx counter --amount=1000mycoin --sequence=2 --name=${RICH} 2>/dev/null"

  TX=$(echo qwertyuiop | ${CLIENT_EXE} tx counter --amount=10mycoin --sequence=2 --name=${RICH} --valid --countfee=5mycoin)
  txSucceeded $? "$TX"
  HASH=$(echo $TX | jq .hash | tr -d \")
  TX_HEIGHT=$(echo $TX | jq .height)

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

# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
