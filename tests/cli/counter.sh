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

    assertFalse "Line=${LINENO}, requires arg" "${CLIENT_EXE} query account"

    checkAccount $SENDER "9007199254740992"

    ACCT2=$(${CLIENT_EXE} query account $RECV 2>/dev/null)
    assertFalse "Line=${LINENO}, has no genesis account" $?
}

test01SendTx() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "Line=${LINENO}, missing dest" "${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 2>/dev/null"
    assertFalse "Line=${LINENO}, bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH 2>/dev/null"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH)
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
    assertFalse "Line=${LINENO}, no default count" $?
}

# checkCounter $COUNT $BALANCE
# Assumes just one coin, checks the balance of first coin in any case
checkCounter() {
    # make sure sender goes down
    ACCT=$(${CLIENT_EXE} query counter)
    if assertTrue "Line=${LINENO}, count is set" $?; then
        assertEquals "Line=${LINENO}, proper count" "$1" $(echo $ACCT | jq .data.counter)
        assertEquals "Line=${LINENO}, proper money" "$2" $(echo $ACCT | jq .data.total_fees[0].amount)
    fi
}

test03AddCount() {
    SENDER=$(getAddr $RICH)
    assertFalse "Line=${LINENO}, bad password" "echo hi | ${CLIENT_EXE} tx counter --countfee=100mycoin --sequence=2 --name=${RICH} 2>/dev/null"

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
    if assertTrue "Line=${LINENO}, found tx" $?; then
        assertEquals "Line=${LINENO}, proper height" $TX_HEIGHT $(echo $TX | jq .height)
        assertEquals "Line=${LINENO}, type=sigs/one" '"sigs/one"' $(echo $TX | jq .data.type)
        NTX=$(echo $TX | jq .data.data.tx)
        assertEquals "line=${LINENO}, type=nonce" '"nonce"' $(echo $NTX | jq .type)
        CTX=$(echo $NTX | jq .data.tx)
        assertEquals "Line=${LINENO}, type=chain/tx" '"chain/tx"' $(echo $CTX | jq .type)
        CNTX=$(echo $CTX | jq .data.tx)
        assertEquals "Line=${LINENO}, type=cntr/count" '"cntr/count"' $(echo $CNTX | jq .type)
        assertEquals "Line=${LINENO}, proper fee" "10" $(echo $CNTX | jq .data.fee[0].amount)
    fi

    # test again with fees...
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx counter --countfee=7mycoin --fee=4mycoin --sequence=3 --name=${RICH} --valid)
    txSucceeded $? "$TX" "counter"

    # make sure the counter was updated, added 7
    checkCounter "2" "17"

    # make sure the account was debited 11
    checkAccount $SENDER "9007199254739979"

    # make sure we cannot replay the counter, no state change
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx counter --countfee=10mycoin --sequence=2 --name=${RICH} --valid 2>/dev/null)
    assertFalse "replay: $TX" $?
    checkCounter "2" "17"
    checkAccount $SENDER "9007199254739979"
}

# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
