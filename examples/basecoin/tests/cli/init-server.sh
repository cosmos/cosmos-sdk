#!/bin/bash

CLIENT_EXE=basecli
SERVER_EXE=basecoin

test01initOption() {
    BASE=~/.bc_init_test
    rm -rf "$BASE"
    mkdir -p "$BASE"

    SERVE_DIR="${BASE}/server"
    GENESIS_FILE=${SERVE_DIR}/genesis.json
    HEX="deadbeef1234deadbeef1234deadbeef1234aaaa"

    ${SERVER_EXE} init ${HEX} --home="$SERVE_DIR" -p=app1/key1/val1 -p='"app2/key2/{""name"": ""joe"", ""age"": ""100""}"' >/dev/null
    if ! assertTrue "line=${LINENO}" $?; then return 1; fi

    OPTION1KEY=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[2]')
    OPTION1VAL=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[3]')
    OPTION2KEY=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[4]')
    OPTION2VAL=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[5]')
    OPTION2VALEXPECTED=$(echo '{"name": "joe", "age": "100"}' | jq '.')

    assertEquals "line=${LINENO}" '"app1/key1"' $OPTION1KEY
    assertEquals "line=${LINENO}" '"val1"' $OPTION1VAL
    assertEquals "line=${LINENO}" '"app2/key2"' $OPTION2KEY
    assertEquals "line=${LINENO}" "$OPTION2VALEXPECTED" "$OPTION2VAL"
}

test02runServer() {
    # Attempt to begin the server with the custom genesis
    SERVER_LOG=$BASE/${SERVER_EXE}.log
    startServer $SERVE_DIR $SERVER_LOG
}

oneTimeTearDown() {
    quickTearDown
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
CLI_DIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/common.sh
. $CLI_DIR/shunit2
