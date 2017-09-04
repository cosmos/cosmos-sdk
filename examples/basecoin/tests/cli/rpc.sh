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

    # this sets the base for all client queries in the tests
    export BCHOME=${BASE}/client
    ${CLIENT_EXE} init --node=tcp://localhost:46657 --genesis=${GENESIS_FILE} > /dev/null 2>&1
    if ! assertTrue "line=${LINENO}, initialized light-client" "$?"; then
        return 1
    fi
}

oneTimeTearDown() {
    printf "\nstopping ${SERVER_EXE}..."
    kill -9 $PID_SERVER >/dev/null 2>&1
    sleep 1
}

test01GetInsecure() {
    GENESIS=$(${CLIENT_EXE} rpc genesis)
    assertTrue "line=${LINENO}, get genesis" "$?"
    MYCHAIN=$(echo ${GENESIS} | jq .genesis.chain_id | tr -d \")
    assertEquals "line=${LINENO}, genesis chain matches" "${CHAIN_ID}" "${MYCHAIN}"

    STATUS=$(${CLIENT_EXE} rpc status)
    assertTrue "line=${LINENO}, get status" "$?"
    SHEIGHT=$(echo ${STATUS} | jq .latest_block_height)
    assertTrue "line=${LINENO}, parsed status" "$?"
    assertNotNull "line=${LINENO}, has a height" "${SHEIGHT}"

    VALS=$(${CLIENT_EXE} rpc validators)
    assertTrue "line=${LINENO}, get validators" "$?"
    VHEIGHT=$(echo ${VALS} | jq .block_height)
    assertTrue "line=${LINENO}, parsed validators" "$?"
    assertTrue "line=${LINENO}, sensible heights: $SHEIGHT / $VHEIGHT" "test $VHEIGHT -ge $SHEIGHT"
    VCNT=$(echo ${VALS} | jq '.validators | length')
    assertEquals "line=${LINENO}, one validator" "1" "$VCNT"

    INFO=$(${CLIENT_EXE} rpc info)
    assertTrue "line=${LINENO}, get info" "$?"
    DATA=$(echo $INFO | jq .response.data)
    assertEquals "line=${LINENO}, basecoin info" '"Basecoin v0.7.0-alpha"' "$DATA"
}

test02GetSecure() {
    HEIGHT=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
    assertTrue "line=${LINENO}, get status" "$?"

    # check block produces something reasonable
    assertFalse "line=${LINENO}, missing height" "${CLIENT_EXE} rpc block"
    BLOCK=$(${CLIENT_EXE} rpc block --height=$HEIGHT)
    assertTrue "line=${LINENO}, get block" "$?"
    MHEIGHT=$(echo $BLOCK | jq .block_meta.header.height)
    assertEquals "line=${LINENO}, meta height" "${HEIGHT}" "${MHEIGHT}"
    BHEIGHT=$(echo $BLOCK | jq .block.header.height)
    assertEquals "line=${LINENO}, meta height" "${HEIGHT}" "${BHEIGHT}"

    # check commit produces something reasonable
    assertFalse "line=${LINENO}, missing height" "${CLIENT_EXE} rpc commit"
    let "CHEIGHT = $HEIGHT - 1"
    COMMIT=$(${CLIENT_EXE} rpc commit --height=$CHEIGHT)
    assertTrue "line=${LINENO}, get commit" "$?"
    HHEIGHT=$(echo $COMMIT | jq .header.height)
    assertEquals "line=${LINENO}, commit height" "${CHEIGHT}" "${HHEIGHT}"
    assertEquals "line=${LINENO}, canonical" "true" $(echo $COMMIT | jq .canonical)
    BSIG=$(echo $BLOCK | jq .block.last_commit)
    CSIG=$(echo $COMMIT | jq .commit)
    assertEquals "line=${LINENO}, block and commit" "$BSIG" "$CSIG"

    # now let's get some headers
    # assertFalse "missing height" "${CLIENT_EXE} rpc headers"
    HEADERS=$(${CLIENT_EXE} rpc headers --min=$CHEIGHT --max=$HEIGHT)
    assertTrue "line=${LINENO}, get headers" "$?"
    assertEquals "line=${LINENO}, proper height" "$HEIGHT" $(echo $HEADERS | jq '.block_metas[0].header.height')
    assertEquals "line=${LINENO}, two headers" "2" $(echo $HEADERS | jq '.block_metas | length')
    # should we check these headers?
    CHEAD=$(echo $COMMIT | jq .header)
    # most recent first, so the commit header is second....
    HHEAD=$(echo $HEADERS | jq .block_metas[1].header)
    assertEquals "line=${LINENO}, commit and header" "$CHEAD" "$HHEAD"
}

test03Waiting() {
    START=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
    assertTrue "line=${LINENO}, get status" "$?"

    let "NEXT = $START + 5"
    assertFalse "line=${LINENO}, no args" "${CLIENT_EXE} rpc wait"
    assertFalse "line=${LINENO}, too long" "${CLIENT_EXE} rpc wait --height=1234"
    assertTrue "line=${LINENO}, normal wait" "${CLIENT_EXE} rpc wait --height=$NEXT"

    STEP=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
    assertEquals "line=${LINENO}, wait until height" "$NEXT" "$STEP"

    let "NEXT = $STEP + 3"
    assertTrue "line=${LINENO}, ${CLIENT_EXE} rpc wait --delta=3"
    STEP=$(${CLIENT_EXE} rpc status | jq .latest_block_height)
    assertEquals "line=${LINENO}, wait for delta" "$NEXT" "$STEP"
}

# load and run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
CLI_DIR=$GOPATH/src/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/shunit2
