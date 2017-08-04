#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}
DUDE=${ACCOUNTS[2]}
ROLE="10CAFE4E"

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

    SIGS=2

    assertFalse "line=${LINENO}, missing min-sigs" "echo qwertyuiop | ${CLIENT_EXE} tx create-role --role=${ROLE} --members=${MEMBERS} --sequence=1 --name=$RICH"
    assertFalse "line=${LINENO}, missing members" "echo qwertyuiop | ${CLIENT_EXE} tx create-role --role=${ROLE} --min-sigs=2 --sequence=1 --name=$RICH"
    assertFalse "line=${LINENO}, missing role" "echo qwertyuiop | ${CLIENT_EXE} tx create-role --min-sigs=2 --members=${MEMBERS} --sequence=1 --name=$RICH"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx create-role --role=${ROLE} --min-sigs=$SIGS  --members=${MEMBERS} --sequence=1 --name=$RICH)
    txSucceeded $? "$TX" "${ROLE}"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    checkRole "${ROLE}" $SIGS 3

    # Make sure tx is indexed
    checkRoleTx $HASH $TX_HEIGHT "${ROLE}" 3
}

test02SendTxToRole() {
    SENDER=$(getAddr $RICH)
    RECV=role:${ROLE}

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --fee=90mycoin --amount=10000mycoin --to=$RECV --sequence=2 --name=$RICH)
    txSucceeded $? "$TX" "${ROLE}"
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
    BANK=role:${ROLE}

    # no money to start mr. poor...
    assertFalse "line=${LINENO}, has no money yet" "${CLIENT_EXE} query account $TWO 2>/dev/null"

    # let's try to send money from the role directly without multisig
    FAIL=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=6000mycoin --from=$BANK --to=$TWO --sequence=1 --name=$POOR 2>/dev/null)
    assertFalse "need to assume role" $?
    FAIL=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=6000mycoin --from=$BANK --to=$TWO --sequence=2 --assume-role=${ROLE} --name=$POOR 2>/dev/null)
    assertFalse "need two signatures" $?

    # okay, begin a multisig transaction mr. poor...
    TX_FILE=$BASE_DIR/tx.json
    echo qwertyuiop | ${CLIENT_EXE} tx send --amount=6000mycoin --from=$BANK --to=$TWO --sequence=1 --assume-role=${ROLE} --name=$POOR --multi --prepare=$TX_FILE
    assertTrue "line=${LINENO}, successfully prepare tx" $?
    # and get some dude to sign it
    # FAIL=$(echo qwertyuiop | ${CLIENT_EXE} tx --in=$TX_FILE --name=$POOR 2>/dev/null)
    # assertFalse "line=${LINENO}, double signing doesn't get bank" $?
    # and get some dude to sign it for the full access
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx --in=$TX_FILE --name=$DUDE)
    txSucceeded $? "$TX" "multi-bank"

    checkAccount $TWO "6000"
    checkAccount $BANK "4000"
}


# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
