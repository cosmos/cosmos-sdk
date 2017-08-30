#!/bin/bash

CLIENT_EXE=basecli
SERVER_EXE=basecoin

test01initOption() {
    BASE=~/.bc_init_test
    rm -rf "$BASE"
    mkdir -p "$BASE"

    SERVER="${BASE}/server"
    GENESIS_FILE=${SERVER}/genesis.json
    HEX="deadbeef1234deadbeef1234deadbeef1234aaaa"

    ${SERVER_EXE} init ${HEX} --home="$SERVER" --option=app1/key1/val1 --option=app2/key2/val2 >/dev/null
    if ! assertTrue "line=${LINENO}" $?; then return 1; fi

    OPTION1KEY=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[2]')
    OPTION1VAL=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[3]')
    OPTION2KEY=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[4]')
    OPTION2VAL=$(cat ${GENESIS_FILE} | jq '.app_options.plugin_options[5]')

    assertEquals "line=${LINENO}" '"app1/key1"' $OPTION1KEY
    assertEquals "line=${LINENO}" '"val1"' $OPTION1VAL
    assertEquals "line=${LINENO}" '"app2/key2"' $OPTION2KEY
    assertEquals "line=${LINENO}" '"val2"' $OPTION2VAL
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
