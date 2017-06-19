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
  KEY=$(echo $KEYPASS | ${EXE} new $1)
  assertTrue "created $1" $?
  return $?
}

# updateKey <name> <oldkey> <newkey>
updateKey() {
  (echo $2; echo $3) | keys update $1 > /dev/null
  return $?
}

test00MakeKeys() {
  USER=demouser
  assertFalse "already user $USER" "${EXE} get $USER"
  newKey $USER
  assertTrue "no user $USER" "${EXE} get $USER"
  # make sure bad password not accepted
  assertFalse "accepts short password" "echo 123 | keys new badpass"
}

test01ListKeys() {
  # one line plus the number of keys
  assertEquals "2" $(keys list | wc -l)
  newKey foobar
  assertEquals "3" $(keys list | wc -l)
  # we got the proper name here...
  assertEquals "foobar" $(keys list -o json | jq .[1].name | tr -d \" )
  # we get all names in normal output
  EXPECTEDNAMES=$(echo demouser; echo foobar)
  TEXTNAMES=$(keys list | tail -n +2 | cut -f1)
  assertEquals "$EXPECTEDNAMES" "$TEXTNAMES"
  # let's make sure the addresses match!
  assertEquals "text and json addresses don't match" $(keys list | tail -1 | cut -f3) $(keys list -o json | jq .[1].address | tr -d \")
}

test02updateKeys() {
  USER=changer
  PASS1=awsedrftgyhu
  PASS2=S4H.9j.D9S7hso
  PASS3=h8ybO7GY6d2

  newKey $USER $PASS1
  assertFalse "accepts invalid pass" "updateKey $USER $PASS2 $PASS2"
  assertTrue "doesn't update" "updateKey $USER $PASS1 $PASS2"
  assertTrue "takes new key after update" "updateKey $USER $PASS2 $PASS3"
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
