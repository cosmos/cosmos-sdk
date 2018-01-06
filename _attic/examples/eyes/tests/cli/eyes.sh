#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=eyes
CLIENT_EXE=eyescli

oneTimeSetUp() {
    # These are passed in as args
    BASE_DIR=$HOME/.test_eyes
    CHAIN_ID="eyes-cli-test"

    rm -rf $BASE_DIR 2>/dev/null
    mkdir -p $BASE_DIR

    echo "Setting up genesis..."
    SERVE_DIR=${BASE_DIR}/server
    SERVER_LOG=${BASE_DIR}/${SERVER_EXE}.log

    echo "Starting ${SERVER_EXE} server..."
    export EYE_HOME=${SERVE_DIR}
    ${SERVER_EXE} init --chain-id=$CHAIN_ID  >>$SERVER_LOG
    startServer $SERVE_DIR $SERVER_LOG
    [ $? = 0 ] || return 1

    # Set up client - make sure you use the proper prefix if you set
    #   a custom CLIENT_EXE
    export EYE_HOME=${BASE_DIR}/client

    initClient $CHAIN_ID
    [ $? = 0 ] || return 1

    printf "...Testing may begin!\n\n\n"
}

oneTimeTearDown() {
    quickTearDown
}

test00SetGetRemove() {
    KEY="CAFE6000"
    VALUE="F00D4200"

    assertFalse "line=${LINENO} data present" "${CLIENT_EXE} query eyes ${KEY}"

    # set data
    TXRES=$(${CLIENT_EXE} tx set --key=${KEY} --value=${VALUE})
    txSucceeded $? "$TXRES" "set cafe"
    HASH=$(echo $TXRES | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TXRES | jq .height)

    # make sure it is set
    DATA=$(${CLIENT_EXE} query eyes ${KEY} --height=$TX_HEIGHT)
    assertTrue "line=${LINENO} data not set" $?
    assertEquals "line=${LINENO}" "\"${VALUE}\"" $(echo $DATA | jq .data.value)

    # query the tx
    TX=$(${CLIENT_EXE} query tx $HASH)
    assertTrue "line=${LINENO}, found tx" $?
    if [ -n "$DEBUG" ]; then echo $TX; echo; fi

    assertEquals "line=${LINENO}, proper type" "\"eyes/set\"" $(echo $TX | jq .data.type)
    assertEquals "line=${LINENO}, proper key" "\"${KEY}\"" $(echo $TX | jq .data.data.key)
    assertEquals "line=${LINENO}, proper value" "\"${VALUE}\"" $(echo $TX | jq .data.data.value)
}


# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
CLI_DIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/common.sh
. $CLI_DIR/shunit2

