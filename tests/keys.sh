#!/bin/bash

EXE=keys

oneTimeSetUp() {
  PASS=qwertyuiop
  export TM_HOME=$HOME/.keys_test
  rm -rf $TM_HOME
  assertTrue $?
}

newKey(){
  assertNotNull "keyname required" "$1"
  KEYPASS=${2:-qwertyuiop}
  (echo $KEYPASS; echo $KEYPASS) | ${EXE} new $1 >/dev/null 2>&1
  assertTrue "created $1" $?
}

testMakeKeys() {
  USER=demouser
  assertFalse "already user $USER" "${EXE} list | grep $USER"
  assertEquals "1" `${EXE} list | wc -l`
  newKey $USER
  assertTrue "no user $USER" "${EXE} list | grep $USER"
  assertEquals "2" `${EXE} list | wc -l`
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
