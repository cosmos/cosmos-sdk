#! /bin/bash
set -eu

cd $GOPATH/src/github.com/tendermint/basecoin/demo

function removeQuotes() {
	temp="${1%\"}"
	temp="${temp#\"}"
	echo "$temp"
}

# grab the chain ids
CHAIN_ID1=$(cat ./data/chain1/basecoin/genesis.json | jq .[1])
CHAIN_ID1=$(removeQuotes $CHAIN_ID1)
CHAIN_ID2=$(cat ./data/chain2/basecoin/genesis.json | jq .[1])
CHAIN_ID2=$(removeQuotes $CHAIN_ID2)
echo "CHAIN_ID1: $CHAIN_ID1"
echo "CHAIN_ID2: $CHAIN_ID2"

# make reusable chain flags
CHAIN_FLAGS1="--chain_id $CHAIN_ID1 --from ./data/chain1/basecoin/priv_validator.json"
CHAIN_FLAGS2="--chain_id $CHAIN_ID2 --from ./data/chain2/basecoin/priv_validator.json --node tcp://localhost:36657"

echo ""
echo "... starting chains"
echo ""
# start the first node
TMROOT=./data/chain1/tendermint tendermint node &> chain1_tendermint.log &
basecoin start --ibc-plugin --dir ./data/chain1/basecoin &> chain1_basecoin.log &

# start the second node
TMROOT=./data/chain2/tendermint tendermint node --node_laddr tcp://localhost:36656 --rpc_laddr tcp://localhost:36657 --proxy_app tcp://localhost:36658 &> chain2_tendermint.log &
basecoin start --address tcp://localhost:36658 --ibc-plugin --dir ./data/chain2/basecoin &> chain2_basecoin.log &

echo ""
echo "... waiting for chains to start"
echo ""
sleep 5

echo "... registering chain1 on chain2"
echo ""
# register chain1 on chain2
basecoin ibc --amount 10 $CHAIN_FLAGS2 register --chain_id $CHAIN_ID1 --genesis ./data/chain1/tendermint/genesis.json

echo ""
echo "... creating egress packet on chain1"
echo ""
# create a packet on chain1 destined for chain2
PAYLOAD="DEADBEEF" #TODO
basecoin ibc --amount 10 $CHAIN_FLAGS1 packet create --from $CHAIN_ID1 --to $CHAIN_ID2 --type coin --payload $PAYLOAD --sequence 1

echo ""
echo "... querying for packet data"
echo ""
# query for the packet data and proof
QUERY_RESULT=$(basecoin query ibc,egress,$CHAIN_ID1,$CHAIN_ID2,1)
HEIGHT=$(echo $QUERY_RESULT | jq .height)
PACKET=$(echo $QUERY_RESULT | jq .value)
PROOF=$(echo $QUERY_RESULT | jq .proof)
echo "QUERY_RESULT: $QUERY_RESULT"
echo "HEIGHT: $HEIGHT"
echo "PACKET: $PACKET"
echo "PROOF: $PROOF"

echo ""
echo "... querying for block data"
echo ""
# get the header and commit for the height
HEADER_AND_COMMIT=$(basecoin block $HEIGHT) 
HEADER=$(echo $HEADER_AND_COMMIT | jq.hex.header)
COMMIT=$(echo $HEADER_AND_COMMIT | jq.hex.commit)
echo "HEADER_AND_COMMIT: $HEADER_AND_COMMIT"
echo "HEADER: $HEADER"
echo "COMMIT: $COMMIT"

echo ""
echo "... updating state of chain1 on chain2"
echo ""
# update the state of chain1 on chain2
basecoin ibc --amount 10 $CHAIN_FLAGS2 update --header 0x$HEADER --commit 0x$COMMIT

echo ""
echo "... posting packet from chain1 on chain2"
echo ""
# post the packet from chain1 to chain2
basecoin ibc --amount 10 $CHAIN_FLAGS2 packet post --from $CHAIN_ID1 --height $HEIGHT --packet $PACKET --proof $PROOF
