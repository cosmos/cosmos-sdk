#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

oneTimeSetUp() {
    quickSetup .basecoin_test_basictx basictx-chain
}

oneTimeTearDown() {
    quickTearDown
}

test00GetAccount() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "requires arg" "${CLIENT_EXE} query account"

    checkAccount $SENDER "0" "9007199254740992"

    ACCT2=$(${CLIENT_EXE} query account $RECV 2>/dev/null)
    assertFalse "has no genesis account" $?
}

test01SendTx() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "missing dest" "${CLIENT_EXE} tx send --amount=992mycoin --sequence=1"
    assertFalse "bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    checkAccount $SENDER "1" "9007199254740000"
    # make sure 0x prefix also works
    checkAccount "0x$SENDER" "1" "9007199254740000"
    checkAccount $RECV "0" "992"

    # Make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
