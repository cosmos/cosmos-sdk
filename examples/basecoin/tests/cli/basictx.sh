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

    checkAccount $SENDER "9007199254740000" "$TX_HEIGHT"
    # make sure 0x prefix also works
    checkAccount "0x$SENDER" "9007199254740000" "$TX_HEIGHT"
    checkAccount $RECV "992" "$TX_HEIGHT"

    # Make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992"

    # For demoing output
    # CMD="${CLIENT_EXE} search sent ${SENDER}"
    # echo $CMD
    # $CMD

    SENT_TX=$(${CLIENT_EXE} search sent ${SENDER})
    assertEquals "line=${LINENO}" 1 $(echo ${SENT_TX} | jq '. | length')
    assertEquals "line=${LINENO}" $TX_HEIGHT $(echo ${SENT_TX} | jq '.[0].height')

    IN=$(echo ${SENT_TX} | jq '.[0].tx.inputs')
    assertEquals "line=${LINENO}" 1 $(echo ${IN} | jq '. | length')
    assertEquals "line=${LINENO}" 992 $(echo ${IN} | jq '.[0].coins[0].amount')
    assertEquals "line=${LINENO}" "\"$SENDER\"" $(echo ${IN} | jq '.[0].sender')

    OUT=$(echo ${SENT_TX} | jq '.[0].tx.outputs')
    assertEquals "line=${LINENO}" 1 $(echo ${OUT} | jq '. | length')
    assertEquals "line=${LINENO}" 992 $(echo ${OUT} | jq '.[0].coins[0].amount')
    assertEquals "line=${LINENO}" "\"$RECV\"" $(echo ${OUT} | jq '.[0].receiver')
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
    checkAccount $SENDER "9007199254739900" "$TX_HEIGHT"
    checkAccount $RECV "1082" "$TX_HEIGHT"

    # Make sure tx is indexed
    checkSendFeeTx $HASH $TX_HEIGHT $SENDER "90" "10"

    # assert replay protection
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=90mycoin --fee=10mycoin --sequence=2 --to=$RECV --name=$RICH 2>/dev/null)
    assertFalse "line=${LINENO}, replay: $TX" $?

    # checking normally
    checkAccount $SENDER "9007199254739900" "$TX_HEIGHT"
    checkAccount $RECV "1082" "$TX_HEIGHT"

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
    export BC_TRUST_NODE=1
    export BC_NODE=localhost:46657
    checkSendFeeTx $HASH $TX_HEIGHT $SENDER "90" "10"
    checkAccount $SENDER "9007199254739900" "$TX_HEIGHT"
    checkAccount $RECV "1082" "$TX_HEIGHT"
    unset BC_TRUST_NODE
    unset BC_NODE
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
    checkAccount $RECV "2082" "$TX_HEIGHT"
    checkAccount $SENDER "9007199254739900" "$TX_HEIGHT"
}


# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
CLI_DIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/common.sh
. $CLI_DIR/shunit2
