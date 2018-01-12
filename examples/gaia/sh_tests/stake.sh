#!/bin/bash
set -u

# These global variables are required for common.sh
SERVER_EXE="gaia node"
CLIENT_EXE="gaia client"
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
DELEGATOR=${ACCOUNTS[2]}
POOR=${ACCOUNTS[4]}

BASE_DIR=$HOME/stake_test
BASE_DIR2=$HOME/stake_test2
SERVER1=$BASE_DIR/server
SERVER2=$BASE_DIR2/server

oneTimeSetUp() {
    #[ "$2" ] || echo "missing parameters, line=${LINENO}" ; exit 1;


    # These are passed in as args
    CHAIN_ID="stake_test"

    # TODO Make this more robust
    if [ "$BASE_DIR" == "$HOME/" ]; then
        echo "Must be called with argument, or it will wipe your home directory"
        exit 1
    fi

    rm -rf $BASE_DIR 2>/dev/null
    mkdir -p $BASE_DIR

    if [ "$BASE_DIR2" == "$HOME/" ]; then
        echo "Must be called with argument, or it will wipe your home directory"
        exit 1
    fi
    rm -rf $BASE_DIR2 2>/dev/null
    mkdir -p $BASE_DIR2

    # Set up client - make sure you use the proper prefix if you set
    #   a custom CLIENT_EXE
    export BC_HOME=${BASE_DIR}/client
    prepareClient

    # start the node server
    set +u ; initServer $BASE_DIR $CHAIN_ID ; set -u
    if [ $? != 0 ]; then return 1; fi

    set +u ; initClient $CHAIN_ID ; set -u
    if [ $? != 0 ]; then return 1; fi

    printf "...Testing may begin!\n\n\n"

}

oneTimeTearDown() {
    kill -9 $PID_SERVER2 >/dev/null 2>&1
    set +u ; quickTearDown ; set -u
}

# Ex Usage: checkCandidate $PUBKEY $EXPECTED_VOTING_POWER
checkCandidate() {
    CANDIDATE=$(${CLIENT_EXE} query candidate --pubkey=$1)
    if ! assertTrue "line=${LINENO}, bad query" $?; then
        return 1
    fi
    assertEquals "line=${LINENO}, proper voting power" "$2" $(echo $CANDIDATE | jq .data.voting_power)
    return $?
}

# Ex Usage: checkCandidate $PUBKEY
checkCandidateEmpty() { 
    CANDIDATE=$(${CLIENT_EXE} query candidate --pubkey=$1 2>/dev/null)
    if ! assertFalse "line=${LINENO}, expected empty query" $?; then
        return 1
    fi
}

# Ex Usage: checkCandidate $DELEGATOR_ADDR $PUBKEY $EXPECTED_SHARES
checkDelegatorBond() {
    BOND=$(${CLIENT_EXE} query delegator-bond --delegator-address=$1 --pubkey=$2)
    if ! assertTrue "line=${LINENO}, account must exist" $?; then
        return 1
    fi
    assertEquals "line=${LINENO}, proper bond amount" "$3" $(echo $BOND | jq .data.Shares)
    return $?
}

# Ex Usage: checkCandidate $DELEGATOR_ADDR $PUBKEY
checkDelegatorBondEmpty() { 
    BOND=$(${CLIENT_EXE} query delegator-bond --delegator-address=$1 --pubkey=$2 2>/dev/null)
    if ! assertFalse "line=${LINENO}, expected empty query" $?; then
        return 1
    fi
}

#______________________________________________________________________________________

test00GetAccount() {
    SENDER=$(getAddr $RICH)
    RECV=$(getAddr $POOR)

    assertFalse "line=${LINENO}, requires arg" "${CLIENT_EXE} query account"

    set +u ; checkAccount $SENDER "9007199254740992" ; set -u

    ACCT2=$(${CLIENT_EXE} query account $RECV 2>/dev/null)
    assertFalse "line=${LINENO}, has no genesis account" $?
}

test01SendTx() {
    assertFalse "line=${LINENO}, missing dest" "${CLIENT_EXE} tx send --amount=992fermion --sequence=1"
    assertFalse "line=${LINENO}, bad password" "echo foo | ${CLIENT_EXE} tx send --amount=992fermion --sequence=1 --to=$RECV --name=$RICH"
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=992fermion --sequence=1 --to=$RECV --name=$RICH)
    txSucceeded $? "$TX" "$RECV"
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    set +u 
    checkAccount $SENDER "9007199254740000" $TX_HEIGHT 
    # make sure 0x prefix also works
    checkAccount "0x$SENDER" "9007199254740000" $TX_HEIGHT
    checkAccount $RECV "992" $TX_HEIGHT 

    # Make sure tx is indexed
    checkSendTx $HASH $TX_HEIGHT $SENDER "992" 
    set -u
}

test02DeclareCandidacy() {

    # the premise of this test is to run a second validator (from rich) and then bond and unbond some tokens
    # first create a second node to run and connect to the system

    # init the second node
    SERVER_LOG2=$BASE_DIR2/node2.log
    GENKEY=$(${CLIENT_EXE} keys get ${RICH} | awk '{print $2}')
    ${SERVER_EXE} init $GENKEY --chain-id $CHAIN_ID --home=$SERVER2 >>$SERVER_LOG2
    if [ $? != 0 ]; then return 1; fi

    # copy in the genesis from the first initialization to the new server
    cp $SERVER1/genesis.json $SERVER2/genesis.json

    # point the new config to the old server location
    rm $SERVER2/config.toml
    echo 'proxy_app = "tcp://127.0.0.1:46668"
    moniker = "anonymous"
    fast_sync = true
    db_backend = "leveldb"
    log_level = "state:info,*:error"

    [rpc]
    laddr = "tcp://0.0.0.0:46667"

    [p2p]
    laddr = "tcp://0.0.0.0:46666"
    seeds = "0.0.0.0:46656"' >$SERVER2/config.toml

    # start the second node
    ${SERVER_EXE} start --home=$SERVER2 >>$SERVER_LOG2 2>&1 &
    sleep 1
    PID_SERVER2=$!
    disown
    if ! ps $PID_SERVER2 >/dev/null; then
        echo "**FAILED**"
        cat $SERVER_LOG2
        return 1
    fi

    # get the pubkey of the second validator
    PK2=$(cat $SERVER2/priv_validator.json | jq -r .pub_key.data)

    CAND_ADDR=$(getAddr $POOR)
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx declare-candidacy --sequence=1 --amount=10fermion --name=$POOR --pubkey=$PK2 --moniker=rigey)
    if [ $? != 0 ]; then return 1; fi
    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $CAND_ADDR "982" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "10"
    checkDelegatorBond $CAND_ADDR $PK2 "10"
}

test03Delegate() {
    # send some coins to a delegator
    DELA_ADDR=$(getAddr $DELEGATOR)
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --sequence=2 --amount=15fermion --to=$DELA_ADDR --name=$RICH)
    txSucceeded $? "$TX" "$DELA_ADDR"
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "15" $TX_HEIGHT ; set -u

    # delegate some coins to the new 
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx delegate --sequence=1 --amount=10fermion --name=$DELEGATOR --pubkey=$PK2)
    if [ $? != 0 ]; then return 1; fi
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "5" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "20"
    checkDelegatorBond $DELA_ADDR $PK2 "10"

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx delegate --sequence=2 --amount=3fermion --name=$DELEGATOR --pubkey=$PK2)
    if [ $? != 0 ]; then return 1; fi
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "2" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "23"
    checkDelegatorBond $DELA_ADDR $PK2 "13"

    # attempt a delegation without enough funds
    # NOTE the sequence number still increments here because it will fail 
    #   only during DeliverTx - however this should be updated (TODO) in new
    #   SDK when we can fail in CheckTx
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx delegate --sequence=3 --amount=3fermion --name=$DELEGATOR --pubkey=$PK2 2>/dev/null)
    if [ $? == 0 ]; then return 1; fi
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "2" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "23"
    checkDelegatorBond $DELA_ADDR $PK2 "13"

    # perform the final delegation which should empty the delegators account
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx delegate --sequence=4 --amount=2fermion --name=$DELEGATOR --pubkey=$PK2)
    if [ $? != 0 ]; then return 1; fi
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "null" $TX_HEIGHT ; set -u #empty account is null 
    checkCandidate $PK2 "25"
}

test04Unbond() {
    # unbond from the delegator a bit
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx unbond --sequence=5 --shares=10 --name=$DELEGATOR --pubkey=$PK2)
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "10" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "15"
    checkDelegatorBond $DELA_ADDR $PK2 "5"

    # attempt to unbond more shares than exist
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx unbond --sequence=6 --shares=10 --name=$DELEGATOR --pubkey=$PK2 2>/dev/null)
    if [ $? == 0 ]; then return 1; fi
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "10" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "15"
    checkDelegatorBond $DELA_ADDR $PK2 "5"

    # unbond entirely from the delegator
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx unbond --sequence=6 --shares=5 --name=$DELEGATOR --pubkey=$PK2)
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $DELA_ADDR "15" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "10"
    checkDelegatorBondEmpty $DELA_ADDR $PK2

    # unbond a bit from the owner
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx unbond --sequence=2 --shares=5 --name=$POOR --pubkey=$PK2)
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $CAND_ADDR "987" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "5"
    checkDelegatorBond $CAND_ADDR $PK2 "5"

    # attempt to unbond more shares than exist
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx unbond --sequence=3 --shares=10 --name=$POOR --pubkey=$PK2 2>/dev/null)
    if [ $? == 0 ]; then return 1; fi
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $CAND_ADDR "987" $TX_HEIGHT ; set -u
    checkCandidate $PK2 "5"
    checkDelegatorBond $CAND_ADDR $PK2 "5"

    # unbond entirely from the validator
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx unbond --sequence=3 --shares=5 --name=$POOR --pubkey=$PK2)
    TX_HEIGHT=$(echo $TX | jq .height)
    set +u ; checkAccount $CAND_ADDR "992" $TX_HEIGHT ; set -u
    checkCandidateEmpty $PK2
    checkDelegatorBondEmpty $CAND_ADDR $PK2
}

# Load common then run these tests with shunit2!
CLI_DIR=$GOPATH/src/github.com/cosmos/gaia/vendor/github.com/cosmos/cosmos-sdk/tests/cli

. $CLI_DIR/common.sh
. $CLI_DIR/shunit2
