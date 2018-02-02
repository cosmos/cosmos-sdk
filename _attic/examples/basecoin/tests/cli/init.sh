#!/bin/bash

CLIENT_EXE=basecli
SERVER_EXE=basecoin

oneTimeSetUp() {
    BASE=~/.bc_init_test
    rm -rf "$BASE"
    mkdir -p "$BASE"

    SERVER="${BASE}/server"
    SERVER_LOG="${BASE}/${SERVER_EXE}.log"

    HEX="deadbeef1234deadbeef1234deadbeef1234aaaa"
    ${SERVER_EXE} init ${HEX} --home="$SERVER" >> "$SERVER_LOG"
    if ! assertTrue "line=${LINENO}" $?; then return 1; fi

    GENESIS_FILE=${SERVER}/genesis.json
    CHAIN_ID=$(cat ${GENESIS_FILE} | jq .chain_id | tr -d \")

    printf "starting ${SERVER_EXE}...\n"
    ${SERVER_EXE} start --home="$SERVER" >> "$SERVER_LOG" 2>&1 &
    sleep 5
    PID_SERVER=$!
    disown
    if ! ps $PID_SERVER >/dev/null; then
        echo "**STARTUP FAILED**"
        cat $SERVER_LOG
        return 1
    fi
}

oneTimeTearDown() {
    printf "\nstopping ${SERVER_EXE}..."
    kill -9 $PID_SERVER >/dev/null 2>&1
    sleep 1
}

test01goodInit() {
    export BCHOME=${BASE}/client-01
    assertFalse "line=${LINENO}" "ls ${BCHOME} 2>/dev/null >&2"

    echo y | ${CLIENT_EXE} init --node=tcp://localhost:46657 --chain-id="${CHAIN_ID}" > /dev/null
    assertTrue "line=${LINENO}, initialized light-client" $?
    checkDir $BCHOME 3
}

test02badInit() {
    export BCHOME=${BASE}/client-02
    assertFalse "line=${LINENO}" "ls ${BCHOME} 2>/dev/null >&2"

    # no node where we go
    echo y | ${CLIENT_EXE} init --node=tcp://localhost:9999 --chain-id="${CHAIN_ID}" > /dev/null 2>&1
    assertFalse "line=${LINENO}, invalid init" $?
    # dir there, but empty...
    checkDir $BCHOME 0

    # try with invalid chain id
    echo y | ${CLIENT_EXE} init --node=tcp://localhost:46657 --chain-id="bad-chain-id" > /dev/null 2>&1
    assertFalse "line=${LINENO}, invalid init" $?
    checkDir $BCHOME 0

    # reject the response
    echo n | ${CLIENT_EXE} init --node=tcp://localhost:46657 --chain-id="${CHAIN_ID}" > /dev/null 2>&1
    assertFalse "line=${LINENO}, invalid init" $?
    checkDir $BCHOME 0
}

test03noDoubleInit() {
    export BCHOME=${BASE}/client-03
    assertFalse "line=${LINENO}" "ls ${BCHOME} 2>/dev/null >&2"

    # init properly
    echo y | ${CLIENT_EXE} init --node=tcp://localhost:46657 --chain-id="${CHAIN_ID}" > /dev/null 2>&1
    assertTrue "line=${LINENO}, initialized light-client" $?
    checkDir $BCHOME 3

    # try again, and we get an error
    echo y | ${CLIENT_EXE} init --node=tcp://localhost:46657 --chain-id="${CHAIN_ID}" > /dev/null 2>&1
    assertFalse "line=${LINENO}, warning on re-init" $?
    checkDir $BCHOME 3

    # unless we --force-reset
    echo y | ${CLIENT_EXE} init --force-reset --node=tcp://localhost:46657 --chain-id="${CHAIN_ID}" > /dev/null 2>&1
    assertTrue "line=${LINENO}, re-initialized light-client" $?
    checkDir $BCHOME 3
}

test04acceptGenesisFile() {
    export BCHOME=${BASE}/client-04
    assertFalse "line=${LINENO}" "ls ${BCHOME} 2>/dev/null >&2"

    # init properly
    ${CLIENT_EXE} init --node=tcp://localhost:46657 --genesis=${GENESIS_FILE} > /dev/null 2>&1
    assertTrue "line=${LINENO}, initialized light-client" $?
    checkDir $BCHOME 3
}

# XXX Ex: checkDir $DIR $FILES
# Makes sure directory exists and has the given number of files
checkDir() {
    assertTrue "line=${LINENO}" "ls ${1} 2>/dev/null >&2"
    assertEquals "line=${LINENO}, no files created" "$2" $(ls $1 | wc -l)
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
CLI_DIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/shunit2
