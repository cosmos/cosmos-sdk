#!/bin/bash

#!/bin/bash

# These global variables are required for common.sh
SERVER_EXE=basecoin
CLIENT_EXE=basecli
ACCOUNTS=(jae ethan bucky rigel igor)
RICH=${ACCOUNTS[0]}
POOR=${ACCOUNTS[4]}

# For full stack traces in error output, run
# BC_TRACE=1 ./ibc.sh

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
    if [ $? != 0 ]; then exit 1; fi
    PID_SERVER_1=$PID_SERVER

    # Start second basecoin server, giving money to the key in the second client
    BC_HOME=${CLIENT_2} initServer $BASE_DIR_2 $CHAIN_ID_2 $PREFIX_2
    if [ $? != 0 ]; then exit 1; fi
    PID_SERVER_2=$PID_SERVER

    # Connect both clients
    BC_HOME=${CLIENT_1} initClient $CHAIN_ID_1 $PORT_1
    if [ $? != 0 ]; then exit 1; fi
    BC_HOME=${CLIENT_2} initClient $CHAIN_ID_2 $PORT_2
    if [ $? != 0 ]; then exit 1; fi

    printf "...Testing may begin!\n\n\n"
}

oneTimeTearDown() {
    printf "\n\nstopping both $SERVER_EXE test servers... $PID_SERVER_1 $PID_SERVER_2"
    kill -9 $PID_SERVER_1
    kill -9 $PID_SERVER_2
    sleep 1
}

test00GetAccount() {
    export BC_HOME=${CLIENT_1}
    SENDER_1=$(getAddr $RICH)
    RECV_1=$(getAddr $POOR)

    assertFalse "line=${LINENO}, requires arg" "${CLIENT_EXE} query account 2>/dev/null"
    assertFalse "line=${LINENO}, has no genesis account" "${CLIENT_EXE} query account $RECV_1 2>/dev/null"
    checkAccount $SENDER_1 "9007199254740992"

    export BC_HOME=${CLIENT_2}
    SENDER_2=$(getAddr $RICH)
    RECV_2=$(getAddr $POOR)

    assertFalse "line=${LINENO}, requires arg" "${CLIENT_EXE} query account 2>/dev/null"
    assertFalse "line=${LINENO}, has no genesis account" "${CLIENT_EXE} query account $RECV_2 2>/dev/null"
    checkAccount $SENDER_2 "9007199254740992"

    # Make sure that they have different addresses on both chains (they are random keys)
    assertNotEquals "line=${LINENO}, sender keys must be different" "$SENDER_1" "$SENDER_2"
    assertNotEquals "line=${LINENO}, recipient keys must be different" "$RECV_1" "$RECV_2"
}

test01RegisterChains() {
    # let's get the root seeds to cross-register them
    ROOT_1="$BASE_DIR_1/root_seed.json"
    ${CLIENT_EXE} seeds export $ROOT_1 --home=${CLIENT_1}
    assertTrue "line=${LINENO}, export seed failed" $?

    ROOT_2="$BASE_DIR_2/root_seed.json"
    ${CLIENT_EXE} seeds export $ROOT_2 --home=${CLIENT_2}
    assertTrue "line=${LINENO}, export seed failed" $?

    # register chain2 on chain1
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx ibc-register \
        --sequence=1 --seed=${ROOT_2} --name=$POOR --home=${CLIENT_1})
    txSucceeded $? "$TX" "register chain2 on chain 1"
    # an example to quit early if there is no point in more tests
    if [ $? != 0 ]; then echo "aborting!"; return 1; fi
    # this is used later to check data
    REG_HEIGHT=$(echo $TX | jq .height)

    # register chain1 on chain2 (no money needed... yet)
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx ibc-register \
        --sequence=1 --seed=${ROOT_1} --name=$POOR --home=${CLIENT_2})
    txSucceeded $? "$TX" "register chain1 on chain 2"
    # an example to quit early if there is no point in more tests
    if [ $? != 0 ]; then echo "aborting!"; return 1; fi
}

test02UpdateChains() {
    # let's get the root seeds to cross-register them
    UPDATE_1="$BASE_DIR_1/seed_1.json"
    ${CLIENT_EXE} seeds update --home=${CLIENT_1}  > /dev/null
    ${CLIENT_EXE} seeds export $UPDATE_1 --home=${CLIENT_1}
    assertTrue "line=${LINENO}, export seed failed" $?
    # make sure it is newer than the other....
    assertNewHeight "line=${LINENO}" $ROOT_1 $UPDATE_1

    UPDATE_2="$BASE_DIR_2/seed_2.json"
    ${CLIENT_EXE} seeds update --home=${CLIENT_2} > /dev/null
    ${CLIENT_EXE} seeds export $UPDATE_2 --home=${CLIENT_2}
    assertTrue "line=${LINENO}, export seed failed" $?
    assertNewHeight "line=${LINENO}" $ROOT_2 $UPDATE_2
    # this is used later to check query data
    REGISTER_2_HEIGHT=$(cat $ROOT_2 | jq .checkpoint.header.height)
    UPDATE_2_HEIGHT=$(cat $UPDATE_2 | jq .checkpoint.header.height)

    # update chain2 on chain1
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx ibc-update \
        --sequence=2 --seed=${UPDATE_2} --name=$POOR --home=${CLIENT_1})
    txSucceeded $? "$TX" "update chain2 on chain 1"
    # an example to quit early if there is no point in more tests
    if [ $? != 0 ]; then echo "aborting!"; return 1; fi

    # update chain1 on chain2 (no money needed... yet)
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx ibc-update \
        --sequence=2 --seed=${UPDATE_1} --name=$POOR --home=${CLIENT_2})
    txSucceeded $? "$TX" "update chain1 on chain 2"
    # an example to quit early if there is no point in more tests
    if [ $? != 0 ]; then echo "aborting!"; return 1; fi
}

# make sure all query commands about ibc work...
test03QueryIBC() {
    # just test on one chain, as they are all symetrical
    export BC_HOME=${CLIENT_1}

    # make sure we can list all chains
    CHAINS=$(${CLIENT_EXE} query ibc chains)
    assertTrue "line=${LINENO}, cannot query chains" $?
    assertEquals "1" $(echo $CHAINS | jq '.data | length')
    assertEquals "line=${LINENO}" "\"$CHAIN_ID_2\"" $(echo $CHAINS | jq '.data[0]')

    # error on unknown chain, data on proper chain
    assertFalse "line=${LINENO}, unknown chain" "${CLIENT_EXE} query ibc chain random 2>/dev/null"
    CHAIN_INFO=$(${CLIENT_EXE} query ibc chain $CHAIN_ID_2)
    assertTrue "line=${LINENO}, cannot query chain $CHAIN_ID_2" $?
    assertEquals "line=${LINENO}, register height" $REG_HEIGHT $(echo $CHAIN_INFO | jq .data.registered_at)
    assertEquals "line=${LINENO}, tracked height" $UPDATE_2_HEIGHT $(echo $CHAIN_INFO | jq .data.remote_block)
}

# Trigger a cross-chain sendTx... from RICH on chain1 to POOR on chain2
#   we make sure the money was reduced, but nothing arrived
test04SendIBCPacket() {
    export BC_HOME=${CLIENT_1}

    # make sure there are no packets yet
    PACKETS=$(${CLIENT_EXE} query ibc packets --to=$CHAIN_ID_2 2>/dev/null)
    assertFalse "line=${LINENO}, packet query" $?

    SENDER=$(getAddr $RICH)
    RECV=$(BC_HOME=${CLIENT_2} getAddr $POOR)

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=20002mycoin \
        --to=${CHAIN_ID_2}::${RECV} --name=$RICH)
    txSucceeded $? "$TX" "${CHAIN_ID_2}::${RECV}"
    # quit early if there is no point in more tests
    if [ $? != 0 ]; then echo "aborting!"; return 1; fi

    HASH=$(echo $TX | jq .hash | tr -d \")
    TX_HEIGHT=$(echo $TX | jq .height)

    # Make sure balance went down and tx is indexed
    checkAccount $SENDER "9007199254720990"
    checkSendTx $HASH $TX_HEIGHT $SENDER "20002"

    # look, we wrote a packet
    PACKETS=$(${CLIENT_EXE} query ibc packets --to=$CHAIN_ID_2)
    assertTrue "line=${LINENO}, packets query" $?
    assertEquals "line=${LINENO}, packet count" 1 $(echo $PACKETS | jq .data)

    # and look at the packet itself
    PACKET=$(${CLIENT_EXE} query ibc packet --to=$CHAIN_ID_2 --sequence=0)
    assertTrue "line=${LINENO}, packet query" $?
    assertEquals "line=${LINENO}, proper src" "\"$CHAIN_ID_1\"" $(echo $PACKET | jq .src_chain)
    assertEquals "line=${LINENO}, proper dest" "\"$CHAIN_ID_2\"" $(echo $PACKET | jq .packet.dest_chain)
    assertEquals "line=${LINENO}, proper sequence" "0" $(echo $PACKET | jq .packet.sequence)

    # nothing arrived
    ARRIVED=$(${CLIENT_EXE} query ibc packets --from=$CHAIN_ID_1 --home=$CLIENT_2 2>/dev/null)
    assertFalse "line=${LINENO}, packet query" $?
    assertFalse "line=${LINENO}, no relay running" "BC_HOME=${CLIENT_2} ${CLIENT_EXE} query account $RECV"
}

test05ReceiveIBCPacket() {
    export BC_HOME=${CLIENT_2}

    # make some credit, so we can accept the packet
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx credit --amount=60006mycoin --to=$CHAIN_ID_1:: --name=$RICH)
    txSucceeded $? "$TX" "${CHAIN_ID_1}::"
    checkAccount $CHAIN_ID_1:: "60006"

    # now, we try to post it.... (this is PACKET from last test)

    # get the seed and post it
    SRC_HEIGHT=$(echo $PACKET | jq .src_height)
    # FIXME: this should auto-update on proofs...
    ${CLIENT_EXE} seeds update --height=$SRC_HEIGHT --home=${CLIENT_1}  > /dev/null
    assertTrue "line=${LINENO}, update seed failed" $?

    PACKET_SEED="$BASE_DIR_1/packet_seed.json"
    ${CLIENT_EXE} seeds export $PACKET_SEED --home=${CLIENT_1} #--height=$SRC_HEIGHT
    assertTrue "line=${LINENO}, export seed failed" $?
    # echo "**** SEED ****"
    # cat $PACKET_SEED | jq .

    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx ibc-update \
        --seed=${PACKET_SEED} --name=$POOR)
    txSucceeded $? "$TX" "prepare packet chain1 on chain 2"
    # an example to quit early if there is no point in more tests
    if [ $? != 0 ]; then echo "aborting!"; return 1; fi

    # write the packet to the file
    POST_PACKET="$BASE_DIR_1/post_packet.json"
    echo $PACKET > $POST_PACKET
    # echo "**** POST ****"
    # cat $POST_PACKET | jq .

    # post it as a tx (cross-fingers)
    TX=$(echo qwertyuiop | ${CLIENT_EXE} tx ibc-post \
        --packet=${POST_PACKET} --name=$POOR)
    txSucceeded $? "$TX" "post packet from chain1 on chain 2"

    # TODO: more queries on stuff...

    # look, we wrote a packet
    PACKETS=$(${CLIENT_EXE} query ibc packets --from=$CHAIN_ID_1)
    assertTrue "line=${LINENO}, packets query" $?
    assertEquals "line=${LINENO}, packet count" 1 $(echo $PACKETS | jq .data)
}


# XXX Ex Usage: assertNewHeight $MSG $SEED_1 $SEED_2
# Desc: Asserts that seed2 has a higher block height than seed 1
assertNewHeight() {
    H1=$(cat $2 | jq .checkpoint.header.height)
    H2=$(cat $3 | jq .checkpoint.header.height)
    assertTrue "$MSG" "test $H2 -gt $H1"
    return $?
}

# test01SendIBCTx() {
#     # Trigger a cross-chain sendTx... from RICH on chain1 to POOR on chain2
#     #   we make sure the money was reduced, but nothing arrived
#     SENDER=$(BC_HOME=${CLIENT_1} getAddr $RICH)
#     RECV=$(BC_HOME=${CLIENT_2} getAddr $POOR)

#     export BC_HOME=${CLIENT_1}
#     TX=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=20002mycoin \
#         --sequence=1 --to=${CHAIN_ID_2}/${RECV} --name=$RICH)
#     txSucceeded $? "$TX" "${CHAIN_ID_2}/${RECV}"
#     # an example to quit early if there is no point in more tests
#     if [ $? != 0 ]; then echo "aborting!"; return 1; fi

#     HASH=$(echo $TX | jq .hash | tr -d \")
#     TX_HEIGHT=$(echo $TX | jq .height)

#     # Make sure balance went down and tx is indexed
#     checkAccount $SENDER "1" "9007199254720990"
#     checkSendTx $HASH $TX_HEIGHT $SENDER "20002"

#     # Make sure nothing arrived - yet
#     waitForBlock ${PORT_1}
#     assertFalse "line=${LINENO}, no relay running" "BC_HOME=${CLIENT_2} ${CLIENT_EXE} query account $RECV"

#     # Start the relay and wait a few blocks...
#     # (already sent a tx on chain1, so use higher sequence)
#     startRelay 2 1
#     if [ $? != 0 ]; then echo "can't start relay"; cat ${BASE_DIR_1}/../relay.log; return 1; fi

#     # Give it a little time, then make sure the money arrived
#     echo "waiting for relay..."
#     sleep 1
#     waitForBlock ${PORT_1}
#     waitForBlock ${PORT_2}

#     # Check the new account
#     echo "checking ibc recipient..."
#     BC_HOME=${CLIENT_2} checkAccount $RECV "0" "20002"

#     # Stop relay
#     printf "stoping relay\n"
#     kill -9 $PID_RELAY
# }

# # StartRelay $seq1 $seq2
# # startRelay hooks up a relay between chain1 and chain2
# # it needs the proper sequence number for $RICH on chain1 and chain2 as args
# startRelay() {
#     # Send some cash to the default key, so it can send messages
#     RELAY_KEY=${BASE_DIR_1}/server/key.json
#     RELAY_ADDR=$(cat $RELAY_KEY | jq .address | tr -d \")
#     echo starting relay $PID_RELAY ...

#     # Get paid on chain1
#     export BC_HOME=${CLIENT_1}
#     SENDER=$(getAddr $RICH)
#     RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=100000mycoin \
#         --sequence=$1 --to=$RELAY_ADDR --name=$RICH)
#     txSucceeded $? "$RES" "$RELAY_ADDR"
#     if [ $? != 0 ]; then echo "can't pay chain1!"; return 1; fi

#     # Get paid on chain2
#     export BC_HOME=${CLIENT_2}
#     SENDER=$(getAddr $RICH)
#     RES=$(echo qwertyuiop | ${CLIENT_EXE} tx send --amount=100000mycoin \
#         --sequence=$2 --to=$RELAY_ADDR --name=$RICH)
#     txSucceeded $? "$RES" "$RELAY_ADDR"
#     if [ $? != 0 ]; then echo "can't pay chain2!"; return 1; fi

#     # Initialize the relay (register both chains)
#     ${SERVER_EXE} relay init --chain1-id=$CHAIN_ID_1 --chain2-id=$CHAIN_ID_2 \
#         --chain1-addr=tcp://localhost:${PORT_1} --chain2-addr=tcp://localhost:${PORT_2} \
#         --genesis1=${BASE_DIR_1}/server/genesis.json --genesis2=${BASE_DIR_2}/server/genesis.json \
#         --from=$RELAY_KEY > ${BASE_DIR_1}/../relay.log
#     if [ $? != 0 ]; then echo "can't initialize relays"; cat ${BASE_DIR_1}/../relay.log; return 1; fi

#     # Now start the relay (constantly send packets)
#     ${SERVER_EXE} relay start --chain1-id=$CHAIN_ID_1 --chain2-id=$CHAIN_ID_2 \
#         --chain1-addr=tcp://localhost:${PORT_1} --chain2-addr=tcp://localhost:${PORT_2} \
#         --from=$RELAY_KEY >> ${BASE_DIR_1}/../relay.log &
#     sleep 2
#     PID_RELAY=$!
#     disown

#     # Return an error if it dies in the first two seconds to make sure it is running
#     ps $PID_RELAY >/dev/null
#     return $?
# }

# Load common then run these tests with shunit2!
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" #get this files directory
. $DIR/common.sh
. $DIR/shunit2
