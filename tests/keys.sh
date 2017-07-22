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
  KEY=$(echo $KEYPASS | ${EXE} new $1 -o json)
  if ! assertTrue "created $1" $?; then return 1; fi
  assertEquals "$1" $(echo $KEY | jq .key.name | tr -d \")
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

test03recoverKeys() {
  USER=sleepy
  PASS1=S4H.9j.D9S7hso

  USER2=easy
  PASS2=1234567890

  # make a user and check they exist
  KEY=$(echo $PASS1 | ${EXE} new $USER -o json)
  if ! assertTrue "created $USER" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY; echo; fi

  SEED=$(echo $KEY | jq .seed | tr -d \")
  ADDR=$(echo $KEY | jq .key.address | tr -d \")
  PUBKEY=$(echo $KEY | jq .key.pubkey | tr -d \")
  assertTrue "${EXE} get $USER > /dev/null"

  # let's delete this key
  assertFalse "echo foo | ${EXE} delete $USER > /dev/null"
  assertTrue "echo $PASS1 | ${EXE} delete $USER > /dev/null"
  assertFalse "${EXE} get $USER > /dev/null"

  # fails on short password
  assertFalse "echo foo; echo $SEED | ${EXE} recover $USER2 -o json > /dev/null"
  # fails on bad seed
  assertFalse "echo $PASS2; echo \"silly white whale tower bongo\" | ${EXE} recover $USER2 -o json > /dev/null"
  # now we got it
  KEY2=$((echo $PASS2; echo $SEED) | ${EXE} recover $USER2 -o json)
  if ! assertTrue "recovery failed: $KEY2" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY2; echo; fi

  # make sure it looks the same
  NAME2=$(echo $KEY2 | jq .name | tr -d \")
  ADDR2=$(echo $KEY2 | jq .address | tr -d \")
  PUBKEY2=$(echo $KEY2 | jq .pubkey | tr -d \")
  assertEquals "wrong username" "$USER2" "$NAME2"
  assertEquals "address doesn't match" "$ADDR" "$ADDR2"
  assertEquals "pubkey doesn't match" "$PUBKEY" "$PUBKEY2"

  # and we can find the info
  assertTrue "${EXE} get $USER2 > /dev/null"
}

# try recovery with secp256k1 keys
test03recoverSecp() {
  USER=dings
  PASS1=Sbub-U9byS7hso

  USER2=booms
  PASS2=1234567890

  KEY=$(echo $PASS1 | ${EXE} new $USER -o json -t secp256k1)
  if ! assertTrue "created $USER" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY; echo; fi

  SEED=$(echo $KEY | jq .seed | tr -d \")
  ADDR=$(echo $KEY | jq .key.address | tr -d \")
  PUBKEY=$(echo $KEY | jq .key.pubkey | tr -d \")
  assertTrue "${EXE} get $USER > /dev/null"

  # now we got it
  KEY2=$((echo $PASS2; echo $SEED) | ${EXE} recover $USER2 -o json)
  if ! assertTrue "recovery failed: $KEY2" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY2; echo; fi

  # make sure it looks the same
  NAME2=$(echo $KEY2 | jq .name | tr -d \")
  ADDR2=$(echo $KEY2 | jq .address | tr -d \")
  PUBKEY2=$(echo $KEY2 | jq .pubkey | tr -d \")
  assertEquals "wrong username" "$USER2" "$NAME2"
  assertEquals "address doesn't match" "$ADDR" "$ADDR2"
  assertEquals "pubkey doesn't match" "$PUBKEY" "$PUBKEY2"

  # and we can find the info
  assertTrue "${EXE} get $USER2 > /dev/null"
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
