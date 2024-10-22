#!/bin/bash

# Function to get the number of transactions in a block
get_tx_count() {
    local block_height=$1
    curl -s "http://localhost:26657/block?height=$block_height" | jq '.result.block.data.txs | length'
}

# Function to get the timestamp of a block
get_block_time() {
    local block_height=$1
    curl -s "http://localhost:26657/block?height=$block_height" | jq -r '.result.block.header.time'
}

strip_milliseconds() {
    echo $1 | sed 's/\.[0-9]*Z/Z/'
}

# Function to calculate the difference in seconds between two timestamps
time_diff_in_seconds() {
    local start_time=$(strip_milliseconds $1)
    local end_time=$(strip_milliseconds $2)
    echo $(( $(date -j -u -f "%Y-%m-%dT%H:%M:%SZ" "$end_time" +%s) - $(date -j -u -f "%Y-%m-%dT%H:%M:%SZ" "$start_time" +%s) ))
}

# Main function to calculate TPS
calculate_tps() {
    local start_block=$1
    local end_block=$2

    local total_txs=0
    local start_time=$(get_block_time $start_block)
    local end_time=$(get_block_time $end_block)

    for (( block=$start_block; block<=$end_block; block++ ))
    do
        tx_count=$(get_tx_count $block)
        total_txs=$((total_txs + tx_count))
    done

    local time_diff=$(time_diff_in_seconds "$start_time" "$end_time")
    local tps=$(echo "scale=2; $total_txs / $time_diff" | bc)

    echo "Total transactions: $total_txs"
    echo "Time window: $time_diff seconds"
    echo "Mean TPS: $tps"
}

# Check if jq is installed
if ! command -v jq &> /dev/null
then
    echo "jq could not be found, please install it to proceed."
    exit 1
fi

# Check if the correct number of arguments is provided
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <start_block> <end_block>"
    exit 1
fi

# Call the main function with the provided arguments
calculate_tps $1 $2
