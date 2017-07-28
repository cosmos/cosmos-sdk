#!/bin/bash

#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

# Uncomment the following line for full stack traces in error output
# CLIENT_EXE="basecli --trace"

oneTimeSetUp() {
    # These are passed in as args
    BASE_DIR_1=$HOME/.basecoin_test_ibc/chain1
    CHAIN_ID_1=test-chain-1
    CLIENT_1=${BASE_DIR_1}/client
    PREFIX_1=1234
    PORT_1=${PREFIX_1}7

    BASE_DIR_2=$HOME/.basecoin_test_ibc/chain2
    CHAIN_ID_2=test-chain-2
    CLIENT_2=${BASE_DIR_2}/client
    PREFIX_2=2345
    PORT_2=${PREFIX_2}7

    # Clean up and create the test dirs
    rm -rf $BASE_DIR_1 $BASE_DIR_2 2>/dev/null
    mkdir -p $BASE_DIR_1 $BASE_DIR_2

    # Set up client for chain 1- make sure you use the proper prefix if you set
    #   a custom CLIENT_EXE
    BC_HOME=${CLIENT_1} prepareClient
    BC_HOME=${CLIENT_2} prepareClient

    # Start basecoin server, giving money to the key in the first client
    BC_HOME=${CLIENT_1} initServer $BASE_DIR_1 $CHAIN_ID_1 $PREFIX_1
    if [ $? != 0 ]; then return 1; fi
    PID_SERVER_1=$PID_SERVER

    # Start second basecoin server, giving money to the key in the second client
    BC_HOME=${CLIENT_2} initServer $BASE_DIR_2 $CHAIN_ID_2 $PREFIX_2
    if [ $? != 0 ]; then return 1; fi
    PID_SERVER_2=$PID_SERVER

    # Connect both clients
    BC_HOME=${CLIENT_1} initClient $CHAIN_ID_1 $PORT_1
    if [ $? != 0 ]; then return 1; fi
    BC_HOME=${CLIENT_2} initClient $CHAIN_ID_2 $PORT_2
    if [ $? != 0 ]; then return 1; fi

    printf "...Testing may begin!\n\n\n"
}

oneTimeTearDown() {
    printf "\n\nstopping both $SERVER_EXE test servers... $PID_SERVER_1 $PID_SERVER_2"
    kill -9 $PID_SERVER_1
    kill -9 $PID_SERVER_2
    sleep 1
}

test00GetAccount() {
    SENDER_1=$(BC_HOME=${CLIENT_1} getAddr $RICH)
    RECV_1=$(BC_HOME=${CLIENT_1} getAddr $POOR)
    export BC_HOME=${CLIENT_1}

    assertFalse "requires arg" "${CLIENT_EXE} query account 2>/dev/null"
    assertFalse "has no genesis account" "${CLIENT_EXE} query account $RECV_1 2>/dev/null"
    checkAccount $SENDER_1 "0" "9007199254740992"

    export BC_HOME=${CLIENT_2}
    SENDER_2=$(getAddr $RICH)
    RECV_2=$(getAddr $POOR)

    assertFalse "requires arg" "${CLIENT_EXE} query account 2>/dev/null"
    assertFalse "has no genesis account" "${CLIENT_EXE} query account $RECV_2 2>/dev/null"
    checkAccount $SENDER_2 "0" "9007199254740992"

    # Make sure that they have different addresses on both chains (they are random keys)
    assertNotEquals "sender keys must be different" "$SENDER_1" "$SENDER_2"
    assertNotEquals "recipient keys must be different" "$RECV_1" "$RECV_2"
}

test01SendIBCTx() {
    # Trigger a cross-chain sendTx... from RICH on chain1 to POOR on chain2
    #   we make sure the money was reduced, but nothing arrived
    SENDER=$(BC_HOME=${CLIENT_1} getAddr $RICH)
    RECV=$(BC_HOME=${CLIENT_2} getAddr $POOR)

    export BC_HOME=${CLIENT_1}
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=20002mycoin \
        --sequence=1 --to=${CHAIN_ID_2}/${RECV} --name=$RICH)
    txSucceeded $? "$TX" "${CHAIN_ID_2}/${RECV}"
    # an example to quit early if there is no point in more tests
    if [ $? != 0 ]; then echo "aborting!"; return 1; fi

    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    # Make sure balance went down and tx is indexed
    checkAccount $SENDER "1" "9007199254720990"
    checkSendTx $HASH $TX_HEIGHT $SENDER "20002"

    # Make sure nothing arrived - yet
    waitForBlock ${PORT_1}
    assertFalse "no relay running" "BC_HOME=${CLIENT_2} ${CLIENT_EXE} query account $RECV"

    # Start the relay and wait a few blocks...
    # (already sent a tx on chain1, so use higher sequence)
    startRelay 2 1
    if [ $? != 0 ]; then echo "can't start relay"; cat ${BASE_DIR_1}/../relay.log; return 1; fi

    # Give it a little time, then make sure the money arrived
    echo "waiting for relay..."
    sleep 1
    waitForBlock ${PORT_1}
    waitForBlock ${PORT_2}

    # Check the new account
    echo "checking ibc recipient..."
    BC_HOME=${CLIENT_2} checkAccount $RECV "0" "20002"

    # Stop relay
    printf "stoping relay\n"
    kill -9 $PID_RELAY
}

# StartRelay $seq1 $seq2
# startRelay hooks up a relay between chain1 and chain2
# it needs the proper sequence number for $RICH on chain1 and chain2 as args
startRelay() {
    # Send some cash to the default key, so it can send messages
    RELAY_KEY=${BASE_DIR_1}/server/key.json
    RELAY_ADDR=$(cat $RELAY_KEY | jq .address | tr -d \")
    echo starting relay $PID_RELAY ...

    # Get paid on chain1
    export BC_HOME=${CLIENT_1}
    SENDER=$(getAddr $RICH)
    RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=100000mycoin \
        --sequence=$1 --to=$RELAY_ADDR --name=$RICH)
    txSucceeded $? "$RES" "$RELAY_ADDR"
    if [ $? != 0 ]; then echo "can't pay chain1!"; return 1; fi

    # Get paid on chain2
    export BC_HOME=${CLIENT_2}
    SENDER=$(getAddr $RICH)
    RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=100000mycoin \
        --sequence=$2 --to=$RELAY_ADDR --name=$RICH)
    txSucceeded $? "$RES" "$RELAY_ADDR"
    if [ $? != 0 ]; then echo "can't pay chain2!"; return 1; fi

    # Initialize the relay (register both chains)
    ${SERVER_EXE} relay init --chain1-id=$CHAIN_ID_1 --chain2-id=$CHAIN_ID_2 \
        --chain1-addr=tcp://localhost:${PORT_1} --chain2-addr=tcp://localhost:${PORT_2} \
        --genesis1=${BASE_DIR_1}/server/genesis.json --genesis2=${BASE_DIR_2}/server/genesis.json \
        --from=$RELAY_KEY > ${BASE_DIR_1}/../relay.log
    if [ $? != 0 ]; then echo "can't initialize relays"; cat ${BASE_DIR_1}/../relay.log; return 1; fi

    # Now start the relay (constantly send packets)
    ${SERVER_EXE} relay start --chain1-id=$CHAIN_ID_1 --chain2-id=$CHAIN_ID_2 \
        --chain1-addr=tcp://localhost:${PORT_1} --chain2-addr=tcp://localhost:${PORT_2} \
        --from=$RELAY_KEY >> ${BASE_DIR_1}/../relay.log &
    sleep 2
    PID_RELAY=$!
    disown

    # Return an error if it dies in the first two seconds to make sure it is running
    ps $PID_RELAY >/dev/null
    return $?
}

# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
