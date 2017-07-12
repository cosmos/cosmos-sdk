#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=counter
CLIENT_EXE=countercli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

oneTimeSetUp() {
    if ! quickSetup .basecoin_test_counter counter-chain; then
        exit 1;
    fi
}

oneTimeTearDown() {
    quickTearDown
}

test00GetAccount() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "requires arg" "${CLIENT_EXE} query account"

    checkAccount $SENDER "9007199254740992"

    ACCT2=$(${CLIENT_EXE} query account $RECV 2>/dev/null)
    assertFalse "has no genesis account" $?
}

test01SendTx() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "missing dest" "${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 2>/dev/null"
    assertFalse "bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    checkAccount $SENDER "9007199254740000"
    checkAccount $RECV "992"

    # make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

test02GetCounter() {
    COUNT=$(${CLIENT_EXE} query counter 2>/dev/null)
    assertFalse "no default count" $?
}

# checkCounter $COUNT $BALANCE
# Assumes just one coin, checks the balance of first coin in any case
checkCounter() {
    # make sure sender goes down
    ACCT=$(${CLIENT_EXE} query counter)
    if assertTrue "count is set" $?; then
      assertEquals "proper count" "$1" $(echo $ACCT | jq .data.counter)
      assertEquals "proper money" "$2" $(echo $ACCT | jq .data.total_fees[0].amount)
    fi
}

test03AddCount() {
    SENDER=$(getAddr $RICH)
    assertFalse "bad password" "echo hi | ${CLIENT_EXE} tx counter --countfee=100mycoin --sequence=2 --name=${RICH} 2>/dev/null"

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx counter --countfee=10mycoin --sequence=2 --name=${RICH} --valid)
    txSucceeded $? "$TX" "counter"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    # make sure the counter was updated
    checkCounter "1" "10"

    # make sure the account was debited
    checkAccount $SENDER "9007199254739990"

    # make sure tx is indexed
    TX=$(${CLIENT_EXE} query tx $HASH --trace)
    if assertTrue "found tx" $?; then
      assertEquals "proper height" $TX_HEIGHT $(echo $TX | jq .height)
      assertEquals "type=sigs/one" '"sigs/one"' $(echo $TX | jq .data.type)
      CTX=$(echo $TX | jq .data.data.tx)
      assertEquals "type=chain/tx" '"chain/tx"' $(echo $CTX | jq .type)
      CNTX=$(echo $CTX | jq .data.tx)
      assertEquals "type=cntr/count" '"cntr/count"' $(echo $CNTX | jq .type)
      assertEquals "proper fee" "10" $(echo $CNTX | jq .data.fee[0].amount)
    fi

    # test again with fees...
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx counter --countfee=7mycoin --fee=4mycoin --sequence=3 --name=${RICH} --valid)
    txSucceeded $? "$TX" "counter"

    # make sure the counter was updated, added 7
    checkCounter "2" "17"

    # make sure the account was debited 11
    checkAccount $SENDER "9007199254739979"
}

# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
