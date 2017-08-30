#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
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
    sleep 0.1 # for startup
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
    assertNotEquals "line=${LINENO}, null key" "null" "$ADDR"
    assertNotEquals "line=${LINENO}, no key" "" "$ADDR"
    echo $ADDR
}

# XXX Ex Usage: restAccount $ADDR $AMOUNT
# Desc: Assumes just one coin, checks the balance of first coin in any case
restAccount() {
    assertNotNull "line=${LINENO}, address required" "$1"
    ACCT=$(curl ${URL}/query/account/sigs:$1 2>/dev/null)
    if [ -n "$DEBUG" ]; then echo $ACCT; echo; fi
    assertEquals "line=${LINENO}, proper money" "$2" $(echo $ACCT | jq .data.coins[0].amount)
    return $?
}

restNoAccount() {
    ERROR=$(curl ${URL}/query/account/sigs:$1 2>/dev/null)
    assertEquals "line=${LINENO}, should error" 400 $(echo $ERROR | jq .code)
}

test00GetAccount() {
    RECV=$(restAddr $POOR)
    SENDER=$(restAddr $RICH)

    restNoAccount $RECV
    restAccount $SENDER "9007199254740992"
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


# XXX Ex Usage: restCreateRole $PAYLOAD $EXPECTED
# Desc: Tests that the first returned signer.addr matches the expected
restCreateRole() {
    assertNotNull "line=${LINENO}, data required" "$1"
    ROLE=$(curl ${URL}/build/create_role --data "$1" 2>/dev/null)
    if [ -n "$DEBUG" ]; then echo -e "$ROLE\n"; fi
    assertEquals "line=${LINENO}, role required" "$2" $(echo $ROLE | jq .data.tx.data.signers[0].addr)
    return $?
}

test03CreateRole() {
    DATA="{\"role\": \"726f6c65\", \"seq\": 1, \"min_sigs\": 1, \"signers\": [{\"addr\": \"4FF759D47C81754D8F553DCCAC8651D0AF74C7F9\", \"app\": \"role\"}]}"
    restCreateRole "$DATA" \""4FF759D47C81754D8F553DCCAC8651D0AF74C7F9"\"
}

test04CreateRoleInvalid() {
    ERROR=$(curl ${URL}/build/create_role --data '{}' 2>/dev/null)
    assertEquals "line=${LINENO}, should report validation failed" 0 $(echo $ERROR | grep "failed" > /dev/null && echo 0 || echo 1)

    ERROR=$(curl ${URL}/build/create_role --data '{"role": "foo"}' 2>/dev/null)
    assertEquals "line=${LINENO}, should report validation failed" 0 $(echo $ERROR | grep "failed" > /dev/null && echo 0 || echo 1)

    ERROR=$(curl ${URL}/build/create_role --data '{"min_sigs": 2, "role": "abcdef"}' 2>/dev/null)
    assertEquals "line=${LINENO}, should report validation failed" 0 $(echo $ERROR | grep "failed" > /dev/null && echo 0 || echo 1)

    ## Non-hex roles should be rejected
    ERROR=$(curl ${URL}/build/create_role --data "{\"role\": \"foobar\", \"seq\": 2, \"signers\": [{\"addr\": \"4FF759D47C81754D8F553DCCAC8651D0AF74C7F9\", \"app\": \"role\"}], \"min_sigs\": 1}" 2>/dev/null)
    assertEquals "line=${LINENO}, should report validation failed" 0 $(echo $ERROR | grep "invalid hex" > /dev/null && echo 0 || echo 1)
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

# TODO: how to handle this if we are not in the same directory
CLI_DIR=${DIR}/../../../../tests/cli

. $CLI_DIR/common.sh
. $CLI_DIR/shunit2
