#!/bin/bash

# these are two globals to control all scripts (can use eg. counter instead)
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

oneTimeSetUp() {
    if ! quickSetup .basecoin_test_restart restart-chain; then
        exit 1;
    fi
}

oneTimeTearDown() {
    quickTearDown
}

test00PreRestart() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992mycoin --sequence=1 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    checkAccount $SENDER "9007199254740000"  "$TX_HEIGHT"
    checkAccount $RECV "992"  "$TX_HEIGHT"

    # make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992"

}

test01OnRestart() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=10000mycoin --sequence=2 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    if [ $? != 0 ]; then echo "can't make tx!"; return 1; fi

    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    # wait til we have quite a few blocks... like at least 20,
    # so the query command won't just wait for the next eg. 7 blocks to verify the result
    echo "waiting to generate lots of blocks..."
    sleep 5
    echo "done waiting!"

    # last minute tx just at the block cut-off...
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=20000mycoin --sequence=3 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    if [ $? != 0 ]; then echo "can't make second tx!"; return 1; fi

    # now we do a restart...
    quickTearDown
    startServer $BASE_DIR/server $BASE_DIR/${SERVER_EXE}.log
    if [ $? != 0 ]; then echo "can't restart server!"; return 1; fi

    # make sure queries still work properly, with all 3 tx now executed
    echo "Checking state after restart..."
    checkAccount $SENDER "9007199254710000"
    checkAccount $RECV "30992"

    # make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "10000"

    # for double-check of logs
    if [ -n "$DEBUG" ]; then
        cat $BASE_DIR/${SERVER_EXE}.log;
    fi
}


# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
CLI_DIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/common.sh
. $CLI_DIR/shunit2

