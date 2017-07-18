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
  if ! assertTrue $?; then return 1; fi

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

  # this sets the base for all client queries in the tests
  export BCHOME=${BASE}/client
  ${CLIENT_EXE} init --node=tcp://localhost:46657 --genesis=${GENESIS_FILE} > /dev/null 2>&1
  if ! assertTrue "initialized light-client" "$?"; then
    return 1
  fi
}

oneTimeTearDown() {
  printf "\nstopping ${SERVER_EXE}..."
  kill -9 $PID_SERVER >/dev/null 2>&1
  sleep 1
}

test01getInsecure() {
  GENESIS=$(${CLIENT_EXE} rpc genesis)
  assertTrue "get genesis" "$?"
  MYCHAIN=$(echo ${GENESIS} | jq .genesis.chain_id | tr -d \")
  assertEquals "genesis chain matches" "${CHAIN_ID}" "${MYCHAIN}"

  STATUS=$(${CLIENT_EXE} rpc status)
  assertTrue "get status" "$?"
  SHEIGHT=$(echo ${STATUS} | jq .latest_block_height)
  assertTrue "parsed status" "$?"
  assertNotNull "has a height" "${SHEIGHT}"

  VALS=$(${CLIENT_EXE} rpc validators)
  assertTrue "get validators" "$?"
  VHEIGHT=$(echo ${VALS} | jq .block_height)
  assertTrue "parsed validators" "$?"
  assertTrue "sensible heights: $SHEIGHT / $VHEIGHT" "test $VHEIGHT -ge $SHEIGHT"
  VCNT=$(echo ${VALS} | jq '.validators | length')
  assertEquals "one validator" "1" "$VCNT"

  INFO=$(${CLIENT_EXE} rpc info)
  assertTrue "get info" "$?"
  DATA=$(echo $INFO | jq .response.data)
  assertEquals "basecoin info" '"Basecoin v0.6.1"' "$DATA"
}

test02getSecure() {
  HEIGHT=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
  assertTrue "get status" "$?"

  # check block produces something reasonable
  assertFalse "missing height" "${CLIENT_EXE} rpc block"
  BLOCK=$(${CLIENT_EXE} rpc block --height=$HEIGHT)
  assertTrue "get block" "$?"
  MHEIGHT=$(echo $BLOCK | jq .block_meta.header.height)
  assertEquals "meta height" "${HEIGHT}" "${MHEIGHT}"
  BHEIGHT=$(echo $BLOCK | jq .block.header.height)
  assertEquals "meta height" "${HEIGHT}" "${BHEIGHT}"

  # check commit produces something reasonable
  assertFalse "missing height" "${CLIENT_EXE} rpc commit"
  let "CHEIGHT = $HEIGHT - 1"
  COMMIT=$(${CLIENT_EXE} rpc commit --height=$CHEIGHT)
  assertTrue "get commit" "$?"
  HHEIGHT=$(echo $COMMIT | jq .header.height)
  assertEquals "commit height" "${CHEIGHT}" "${HHEIGHT}"
  assertEquals "canonical" "true" $(echo $COMMIT | jq .canonical)
  BSIG=$(echo $BLOCK | jq .block.last_commit)
  CSIG=$(echo $COMMIT | jq .commit)
  assertEquals "block and commit" "$BSIG" "$CSIG"

  # now let's get some headers
  # assertFalse "missing height" "${CLIENT_EXE} rpc headers"
  HEADERS=$(${CLIENT_EXE} rpc headers --min=$CHEIGHT --max=$HEIGHT)
  assertTrue "get headers" "$?"
  assertEquals "proper height" "$HEIGHT" $(echo $HEADERS | jq '.last_height')
  assertEquals "two headers" "2" $(echo $HEADERS | jq '.block_metas | length')
  # should we check these headers?
  CHEAD=$(echo $COMMIT | jq .header)
  # most recent first, so the commit header is second....
  HHEAD=$(echo $HEADERS | jq .block_metas[1].header)
  assertEquals "commit and header" "$CHEAD" "$HHEAD"
}

test03waiting() {
  START=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
  assertTrue "get status" "$?"

  let "NEXT = $START + 5"
  assertFalse "no args" "${CLIENT_EXE} rpc wait"
  assertFalse "too long" "${CLIENT_EXE} rpc wait --height=1234"
  assertTrue "normal wait" "${CLIENT_EXE} rpc wait --height=$NEXT"

  STEP=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
  assertEquals "wait until height" "$NEXT" "$STEP"

  let "NEXT = $STEP + 3"
  assertTrue "${CLIENT_EXE} rpc wait --delta=3"
  STEP=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
  assertEquals "wait for delta" "$NEXT" "$STEP"
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/shunit2
