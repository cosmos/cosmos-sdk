#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}
DUDE=${ACCOUNTS[2]}

oneTimeSetUp() {
    if ! quickSetup .basecoin_test_roles roles-chain; then
        exit 1;
    fi
}

oneTimeTearDown() {
    quickTearDown
}

test01SetupRole() {
    ONE=$(getAddr $RICH)
    TWO=$(getAddr $POOR)
    THREE=$(getAddr $DUDE)
    MEMBERS=${ONE},${TWO},${THREE}

    SIGS=1

    assertFalse "line=${LINENO}, missing min-sigs" "echo qwertyuiop | ${CLIENT_EXE} tx create-role --role=bank --members=${MEMBERS} --sequence=1 --name=$RICH"
    assertFalse "line=${LINENO}, missing members" "echo qwertyuiop | ${CLIENT_EXE} tx create-role --role=bank --min-sigs=2 --sequence=1 --name=$RICH"
    assertFalse "line=${LINENO}, missing role" "echo qwertyuiop | ${CLIENT_EXE} tx create-role --min-sigs=2 --members=${MEMBERS} --sequence=1 --name=$RICH"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx create-role --role=bank --min-sigs=$SIGS  --members=${MEMBERS} --sequence=1 --name=$RICH)
    txSucceeded $? "$TX" "bank"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    checkRole bank $SIGS 3

    # Make sure tx is indexed
    checkRoleTx $HASH $TX_HEIGHT "bank" 3
}

test02SendTxToRole() {
    SENDER=$(getAddr $RICH)
    RECV=role:$(toHex bank)

    # HEXROLE=$(toHex bank)
    # RECV="role:$HEXROLE"

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --fee=90mycoin --amount=10000mycoin --to=$RECV --sequence=2 --name=$RICH)
    txSucceeded $? "$TX" "bank"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    # reduce by 10090
    checkAccount $SENDER "9007199254730902"
    checkAccount $RECV "10000"

    checkSendFeeTx $HASH $TX_HEIGHT $SENDER "10000" "90"
}

test03SendMultiFromRole() {
    ONE=$(getAddr $RICH)
    TWO=$(getAddr $POOR)
    THREE=$(getAddr $DUDE)
    BANK=role:$(toHex bank)

    # no money to start mr. poor...
    assertFalse "line=${LINENO}, has no money yet" "${CLIENT_EXE} query account $TWO 2>/dev/null"

    # let's try to send money from the role directly without multisig
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=6000mycoin --from=$BANK --to=$TWO --sequence=1 --name=$POOR 2>/dev/null)
    assertFalse "need to assume role" $?
    # echo qwertyuiop | ${CLIENT_EXE} tx send --amount=6000mycoin --from=$BANK --to=$TWO --sequence=1 --assume-role=bank --name=$POOR --prepare=-
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=6000mycoin --from=$BANK --to=$TWO --sequence=1 --assume-role=bank --name=$POOR)
    txSucceeded $? "$TX" "from-bank"

    checkAccount $TWO "6000"
    checkAccount $BANK "4000"
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
