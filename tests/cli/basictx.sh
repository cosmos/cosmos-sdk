#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

oneTimeSetUp() {
    if ! quickSetup .basecoin_test_basictx basictx-chain; then
        exit 1;
    fi
}

oneTimeTearDown() {
    quickTearDown
}

test00GetAccount() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "line=${LINENO}, requires arg" "${CLIENT_EXE} query account"

    checkAccount $SENDER "9007199254740992"

    ACCT2=$(${CLIENT_EXE} query account $RECV 2>/dev/null)
    assertFalse "line=${LINENO}, has no genesis account" $?
}

test01SendTx() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "line=${LINENO}, missing dest" "${CLIENT_EXE} tx send --amount=992mycoin --sequence=1"
    assertFalse "line=${LINENO}, bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    checkAccount $SENDER "9007199254740000"
    # make sure 0x prefix also works
    checkAccount "0x$SENDER" "9007199254740000"
    checkAccount $RECV "992"

    # Make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

test02SendTxWithFee() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    # Test to see if the auto-sequencing works, the sequence here should be calculated to be 2
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=90mycoin --fee=10mycoin --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    # deduct 100 from sender, add 90 to receiver... fees "vanish"
    checkAccount $SENDER "9007199254739900"
    checkAccount $RECV "1082"

    # Make sure tx is indexed
    checkSendFeeTx $HASH $TX_HEIGHT $SENDER "90" "10"

    # assert replay protection
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=90mycoin --fee=10mycoin --sequence=2 --to=$RECV --name=$RICH 2>/dev/null)
    assertFalse "line=${LINENO}, replay: $TX" $?

    # checking normally
    checkAccount $SENDER "9007199254739900"
    checkAccount $RECV "1082"

    # make sure we can query the proper nonce
    NONCE=$(${CLIENT_EXE} query nonce $SENDER)
    if [ -n "$DEBUG" ]; then echo $NONCE; echo; fi
    # TODO: note that cobra returns error code 0 on parse failure,
    # so currently this check passes even if there is no nonce query command
    if assertTrue "line=${LINENO}, no nonce query" $?; then
        assertEquals "line=${LINENO}, proper nonce" "2" $(echo $NONCE | jq .data)
    fi

    # make sure this works without trust also
    OLD_BC_HOME=$BC_HOME
    export BC_HOME=/foo
    export BCTRUST_NODE=1
    export BCNODE=localhost:46657
    checkSendFeeTx $HASH $TX_HEIGHT $SENDER "90" "10"
    checkAccount $SENDER "9007199254739900"
    checkAccount $RECV "1082"
    unset BCTRUST_NODE
    unset BCNODE
    export BC_HOME=$OLD_BC_HOME
}


test03CreditTx() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    # make sure we are controlled by permissions (only rich can issue credit)
    assertFalse "line=${LINENO}, bad password" "echo qwertyuiop | ${CLIENT_EXE} tx credit --amount=1000mycoin --sequence=1 --to=$RECV --name=$POOR"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx credit --amount=1000mycoin --sequence=3 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    # receiver got cash, sender didn't lose any (1000 more than last check)
    checkAccount $RECV "2082"
    checkAccount $SENDER "9007199254739900"
}


# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
