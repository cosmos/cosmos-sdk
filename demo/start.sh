#! /bin/bash
set -e

cd $GOPATH/src/github.com/tendermint/basecoin/demo

LOG_DIR="."

if [[ "$CIRCLECI" == "true" ]]; then
	# set log dir
	LOG_DIR="${CIRCLE_ARTIFACTS}"

	# install tendermint
	set +e
	go get github.com/tendermint/tendermint
	pushd $GOPATH/src/github.com/tendermint/tendermint
	git checkout develop
	glide install
	go install ./cmd/tendermint
	popd
	set -e
fi

set -u

function removeQuotes() {
	temp="${1%\"}"
	temp="${temp#\"}"
	echo "$temp"
}

function waitForNode() {
        addr=$1
	set +e
        curl -s $addr/status > /dev/null
        ERR=$?
	i=0
        while [ "$ERR" != 0 ]; do
		if [[ "$i" == 10 ]]; then
			echo "waited too long for chain to start"
			exit 1
		fi
		echo "...... still waiting on $addr"
                sleep 1 
                curl -s $addr/status > /dev/null
                ERR=$?
		i=$((i+1))
        done
	set -e
        echo "... node $addr is up"
}

function waitForBlock() {
	addr=$1
	b1=`curl -s $addr/status | jq .result[1].latest_block_height`
	b2=$b1
	while [ "$b2" == "$b1" ]; do
                echo "Waiting for node $addr to commit a block ..."
                sleep 1
		b2=`curl -s $addr/status | jq .result[1].latest_block_height`
	done
}

# make basecoin root vars
export BCHOME="."
BCHOME1="./data/chain1"
BCHOME2="./data/chain2"

# grab the chain ids
CHAIN_ID1=$(cat $BCHOME1/genesis.json | jq .chain_id)
CHAIN_ID1=$(removeQuotes $CHAIN_ID1)
CHAIN_ID2=$(cat $BCHOME2/genesis.json | jq .chain_id)
CHAIN_ID2=$(removeQuotes $CHAIN_ID2)
echo "CHAIN_ID1: $CHAIN_ID1"
echo "CHAIN_ID2: $CHAIN_ID2"

# make reusable chain flags
CHAIN_FLAGS1="--chain_id $CHAIN_ID1 --from $BCHOME1/key.json"
CHAIN_FLAGS2="--chain_id $CHAIN_ID2 --from $BCHOME2/key.json --node tcp://localhost:36657"


echo ""
echo "... starting chains"
echo ""
# start the first node
TMROOT=./data/chain1 tendermint node --skip_upnp --log_level=info &> $LOG_DIR/chain1_tendermint.log &
BCHOME=$BCHOME1 basecoin start --without-tendermint &> $LOG_DIR/chain1_basecoin.log &

# start the second node
TMROOT=./data/chain2 tendermint node --skip_upnp --log_level=info --node_laddr tcp://localhost:36656 --rpc_laddr tcp://localhost:36657 --proxy_app tcp://localhost:36658 &> $LOG_DIR/chain2_tendermint.log &
BCHOME=$BCHOME2 basecoin start --address tcp://localhost:36658 --without-tendermint &> $LOG_DIR/chain2_basecoin.log &

echo ""
echo "... waiting for chains to start"
echo ""

waitForNode localhost:46657
waitForNode localhost:36657

# TODO: remove the sleep
# Without it we sometimes get "Account bytes are empty for address: 053BA0F19616AFF975C8756A2CBFF04F408B4D47"
sleep 3 

echo "... registering chain1 on chain2"
echo ""
# register chain1 on chain2
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS2 register --ibc_chain_id $CHAIN_ID1 --genesis ./data/chain1/genesis.json

echo ""
echo "... creating egress packet on chain1"
echo ""
# create a packet on chain1 destined for chain2
PAYLOAD="DEADBEEF" #TODO
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS1 packet create --ibc_from $CHAIN_ID1 --to $CHAIN_ID2 --type coin --payload $PAYLOAD --ibc_sequence 1

echo ""
echo "... querying for packet data"
echo ""
# query for the packet data and proof
QUERY_RESULT=$(basecoin query ibc,egress,$CHAIN_ID1,$CHAIN_ID2,1)
HEIGHT=$(echo $QUERY_RESULT | jq .height)
PACKET=$(echo $QUERY_RESULT | jq .value)
PROOF=$(echo $QUERY_RESULT | jq .proof)
PACKET=$(removeQuotes $PACKET)
PROOF=$(removeQuotes $PROOF)
echo ""
echo "QUERY_RESULT: $QUERY_RESULT"
echo "HEIGHT: $HEIGHT"
echo "PACKET: $PACKET"
echo "PROOF: $PROOF"


# the query returns the height of the next block, which contains the app hash
# but which may not be committed yet, so we have to wait for it to query the commit
echo ""
echo "... waiting for a block to be committed"
echo ""

waitForBlock localhost:46657
waitForBlock localhost:36657

echo ""
echo "... querying for block data"
echo ""
# get the header and commit for the height
HEADER_AND_COMMIT=$(basecoin block $HEIGHT) 
HEADER=$(echo $HEADER_AND_COMMIT | jq .hex.header)
HEADER=$(removeQuotes $HEADER)
COMMIT=$(echo $HEADER_AND_COMMIT | jq .hex.commit)
COMMIT=$(removeQuotes $COMMIT)
echo ""
echo "HEADER_AND_COMMIT: $HEADER_AND_COMMIT"
echo "HEADER: $HEADER"
echo "COMMIT: $COMMIT"

echo ""
echo "... updating state of chain1 on chain2"
echo ""
# update the state of chain1 on chain2
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS2 update --header 0x$HEADER --commit 0x$COMMIT

echo ""
echo "... posting packet from chain1 on chain2"
echo ""
# post the packet from chain1 to chain2
basecoin tx ibc --amount 10mycoin $CHAIN_FLAGS2 packet post --ibc_from $CHAIN_ID1 --height $HEIGHT --packet 0x$PACKET --proof 0x$PROOF

echo ""
echo "... checking if the packet is present on chain2"
echo ""
# query for the packet on chain2
basecoin query --node tcp://localhost:36657 ibc,ingress,test_chain_2,test_chain_1,1

echo ""
echo "DONE!"
