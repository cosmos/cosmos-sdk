#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jaekwon ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

BPORT=7000
URL="localhost:${BPORT}"

oneTimeSetUp() {
    if ! quickSetup .basecoin_test_rest rest-chain; then
        exit 1;
    fi
    baseserver serve --port $BPORT >/dev/null &
    PID_PROXY=$!
    disown
}

oneTimeTearDown() {
    quickTearDown
    kill -9 $PID_PROXY
}

# XXX Ex Usage: restAddr $NAME
# Desc: Gets the address for a key name via rest
restAddr() {
    assertNotNull "line=${LINENO}, keyname required" "$1"
    ADDR=$(curl ${URL}/keys/${1} 2>/dev/null | jq .address | tr -d \")
    assertNotEquals "line=${LINENO}, no key" "null" $ADDR
    echo $ADDR
}

# XXX Ex Usage: restAccount $ADDR $AMOUNT
# Desc: Assumes just one coin, checks the balance of first coin in any case
restAccount() {
    assertNotNull "line=${LINENO}, address required" "$1"
    ACCT=$(curl ${URL}/query/account/sigs:$1 2>/dev/null)
    if [ -n "$DEBUG" ]; then echo $ACCT; echo; fi
    assertEquals "line=${LINENO}, proper money" "$2" $(echo $ACCT | jq .coins[0].amount)
    # assertEquals "line=${LINENO}, proper money" "$2" $(echo $ACCT | jq .data.coins[0].amount)
    return $?
}

restNoAccount() {
    ERROR=$(curl ${URL}/query/account/sigs:$1 2>/dev/null)
    assertEquals "line=${LINENO}, should error" 406 $(echo $ERROR | jq .code)
}

test00GetAccount() {
    SENDER=$(restAddr $RICH)
    RECV=$(restAddr $POOR)

    restAccount $SENDER "9007199254740992"
    restNoAccount $RECV
}

test01SendTx() {
    SENDER=$(restAddr $RICH)
    RECV=$(restAddr $POOR)

    CMD="{\"from\": {\"app\": \"sigs\", \"addr\": \"$SENDER\"}, \"to\": {\"app\": \"sigs\", \"addr\": \"$RECV\"}, \"amount\": [{\"denom\": \"mycoin\", \"amount\": 992}], \"sequence\": 1}"

    UNSIGNED=$(curl -XPOST ${URL}/build/send -d "$CMD" 2>/dev/null)
    if [ -n "$DEBUG" ]; then echo $UNSIGNED; echo; fi

    TOSIGN="{\"name\": \"$RICH\", \"password\": \"qwertyuiop\", \"tx\": $UNSIGNED}"
    SIGNED=$(curl -XPOST ${URL}/sign -d "$TOSIGN" 2>/dev/null)
    TX=$(curl -XPOST ${URL}/tx -d "$SIGNED" 2>/dev/null)
    if [ -n "$DEBUG" ]; then echo $TX; echo; fi

    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    restAccount $SENDER "9007199254740000"
    restAccount $RECV "992"

    # Make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992"
}

# test02SendTxWithFee() {
#     SENDER=$(getAddr $RICH)
#     RECV=$(getAddr $POOR)

#     # Test to see if the auto-sequencing works, the sequence here should be calculated to be 2
#     TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=90mycoin --fee=10mycoin --to=$RECV --name=$RICH)
#     txSucceeded $? "$TX" "$RECV"
#     HASH=$(echo $TX | jq .hash | tr -d \")
#     TX_HEIGHT=$(echo $TX | jq .height)

#     # deduct 100 from sender, add 90 to receiver... fees "vanish"
#     checkAccount $SENDER "9007199254739900"
#     checkAccount $RECV "1082"

#     # Make sure tx is indexed
#     checkSendFeeTx $HASH $TX_HEIGHT $SENDER "90" "10"

#     # assert replay protection
#     TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=90mycoin --fee=10mycoin --sequence=2 --to=$RECV --name=$RICH 2>/dev/null)
#     assertFalse "line=${LINENO}, replay: $TX" $?
#     checkAccount $SENDER "9007199254739900"
#     checkAccount $RECV "1082"

#     # make sure we can query the proper nonce
#     NONCE=$(${CLIENT_EXE} query nonce $SENDER)
#     if [ -n "$DEBUG" ]; then echo $NONCE; echo; fi
#     # TODO: note that cobra returns error code 0 on parse failure,
#     # so currently this check passes even if there is no nonce query command
#     if assertTrue "line=${LINENO}, no nonce query" $?; then
#         assertEquals "line=${LINENO}, proper nonce" "2" $(echo $NONCE | jq .data)
#     fi
# }


# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
