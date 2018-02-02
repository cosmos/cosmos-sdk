#!/bin/bash
EXE=basecli


oneTimeSetUp() {
    PASS=qwertyuiop
    export BCHOME=$HOME/.bc_keys_test
    ${EXE} reset_all
    assertTrue "line ${LINENO}" $?
}

newKey(){
  assertNotNull "keyname required" "$1"
  KEYPASS=${2:-qwertyuiop}
  KEY=$(echo $KEYPASS | ${EXE} keys new $1 -o json)
  if ! assertTrue "line ${LINENO}: created $1" $?; then return 1; fi
  assertEquals "$1" $(echo $KEY | jq .key.name | tr -d \")
  return $?
}

# updateKey <name> <oldkey> <newkey>
updateKey() {
  (echo $2; echo $3) | ${EXE} keys update $1 > /dev/null
  return $?
}

test00MakeKeys() {
  USER=demouser
  assertFalse "line ${LINENO}: already user $USER" "${EXE} keys get $USER"
  newKey $USER
  assertTrue "line ${LINENO}: no user $USER" "${EXE} keys get $USER"
  # make sure bad password not accepted
  assertFalse "accepts short password" "echo 123 | ${EXE} keys new badpass"
}

test01ListKeys() {
  # one line plus the number of keys
  assertEquals "2" $(${EXE} keys list | wc -l)
  newKey foobar
  assertEquals "3" $(${EXE} keys list | wc -l)
  # we got the proper name here...
  assertEquals "foobar" $(${EXE} keys list -o json | jq .[1].name | tr -d \" )
  # we get all names in normal output
  EXPECTEDNAMES=$(echo demouser; echo foobar)
  TEXTNAMES=$(${EXE} keys list | tail -n +2 | cut -f1)
  assertEquals "$EXPECTEDNAMES" "$TEXTNAMES"
  # let's make sure the addresses match!
  assertEquals "line ${LINENO}: text and json addresses don't match" $(${EXE} keys list | tail -1 | cut -f3) $(${EXE} keys list -o json | jq .[1].address | tr -d \")
}

test02updateKeys() {
  USER=changer
  PASS1=awsedrftgyhu
  PASS2=S4H.9j.D9S7hso
  PASS3=h8ybO7GY6d2

  newKey $USER $PASS1
  assertFalse "line ${LINENO}: accepts invalid pass" "updateKey $USER $PASS2 $PASS2"
  assertTrue "line ${LINENO}: doesn't update" "updateKey $USER $PASS1 $PASS2"
  assertTrue "line ${LINENO}: takes new key after update" "updateKey $USER $PASS2 $PASS3"
}

test03recoverKeys() {
  USER=sleepy
  PASS1=S4H.9j.D9S7hso

  USER2=easy
  PASS2=1234567890

  # make a user and check they exist
  KEY=$(echo $PASS1 | ${EXE} keys new $USER -o json)
  if ! assertTrue "created $USER" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY; echo; fi

  SEED=$(echo $KEY | jq .seed | tr -d \")
  ADDR=$(echo $KEY | jq .key.address | tr -d \")
  PUBKEY=$(echo $KEY | jq .key.pubkey | tr -d \")
  assertTrue "line ${LINENO}" "${EXE} keys get $USER > /dev/null"

  # let's delete this key
  assertFalse "line ${LINENO}" "echo foo | ${EXE} keys delete $USER > /dev/null"
  assertTrue "line ${LINENO}" "echo $PASS1 | ${EXE} keys delete $USER > /dev/null"
  assertFalse "line ${LINENO}" "${EXE} keys get $USER > /dev/null"

  # fails on short password
  assertFalse "line ${LINENO}" "echo foo; echo $SEED | ${EXE} keys recover $USER2 -o json > /dev/null"
  # fails on bad seed
  assertFalse "line ${LINENO}" "echo $PASS2; echo \"silly white whale tower bongo\" | ${EXE} keys recover $USER2 -o json > /dev/null"
  # now we got it
  KEY2=$((echo $PASS2; echo $SEED) | ${EXE} keys recover $USER2 -o json)
  if ! assertTrue "recovery failed: $KEY2" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY2; echo; fi

  # make sure it looks the same
  NAME2=$(echo $KEY2 | jq .name | tr -d \")
  ADDR2=$(echo $KEY2 | jq .address | tr -d \")
  PUBKEY2=$(echo $KEY2 | jq .pubkey | tr -d \")
  assertEquals "line ${LINENO}: wrong username" "$USER2" "$NAME2"
  assertEquals "line ${LINENO}: address doesn't match" "$ADDR" "$ADDR2"
  assertEquals "line ${LINENO}: pubkey doesn't match" "$PUBKEY" "$PUBKEY2"

  # and we can find the info
  assertTrue "line ${LINENO}" "${EXE} keys get $USER2 > /dev/null"
}

# try recovery with secp256k1 keys
test03recoverSecp() {
  USER=dings
  PASS1=Sbub-U9byS7hso

  USER2=booms
  PASS2=1234567890

  KEY=$(echo $PASS1 | ${EXE} keys new $USER -o json -t secp256k1)
  if ! assertTrue "created $USER" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY; echo; fi

  SEED=$(echo $KEY | jq .seed | tr -d \")
  ADDR=$(echo $KEY | jq .key.address | tr -d \")
  PUBKEY=$(echo $KEY | jq .key.pubkey | tr -d \")
  assertTrue "line ${LINENO}" "${EXE} keys get $USER > /dev/null"

  # now we got it
  KEY2=$((echo $PASS2; echo $SEED) | ${EXE} keys recover $USER2 -o json)
  if ! assertTrue "recovery failed: $KEY2" $?; then return 1; fi
  if [ -n "$DEBUG" ]; then echo $KEY2; echo; fi

  # make sure it looks the same
  NAME2=$(echo $KEY2 | jq .name | tr -d \")
  ADDR2=$(echo $KEY2 | jq .address | tr -d \")
  PUBKEY2=$(echo $KEY2 | jq .pubkey | tr -d \")
  assertEquals "line ${LINENO}: wrong username" "$USER2" "$NAME2"
  assertEquals "line ${LINENO}: address doesn't match" "$ADDR" "$ADDR2"
  assertEquals "line ${LINENO}: pubkey doesn't match" "$PUBKEY" "$PUBKEY2"

  # and we can find the info
  assertTrue "line ${LINENO}" "${EXE} keys get $USER2 > /dev/null"
}

# load and run these tests with shunit2!

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
CLI_DIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/shunit2
