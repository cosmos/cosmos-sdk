#!/bin/bash

CLIENT_EXE=basecli


oneTimeSetUp() {
    PASS=qwertyuiop
    export BCHOME=$HOME/.bc_keys_test
    ${CLIENT_EXE} reset_all
    assertTrue "line ${LINENO}" $?
}

newKey(){
    assertNotNull "keyname required" "$1"
    KEYPASS=${2:-qwertyuiop}
    echo $KEYPASS | ${CLIENT_EXE} keys new $1 >/dev/null 2>&1
    assertTrue "line ${LINENO}, created $1" $?
}

testMakeKeys() {
    USER=demouser
    assertFalse "line ${LINENO}, already user $USER" "${CLIENT_EXE} keys get $USER"
    newKey $USER
    assertTrue "line ${LINENO}, no user $USER" "${CLIENT_EXE} keys get $USER"
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory

# TODO: how to handle this if we are not in the same directory
CLI_DIR=${DIR}/../../../../tests/cli

. $CLI_DIR/shunit2
