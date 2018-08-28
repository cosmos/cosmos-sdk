#!/bin/bash
# Script to initialize a testnet settings on a server

#Usage: terraform.sh <testnet_name> <testnet_region_number> <testnet_node_number>

#Add gaiad node number for remote identification
REGION="$(($2 + 1))"
RNODE="$(($3 + 1))"
ID="$((${REGION} * 100 + ${RNODE}))"
echo "$ID" > /etc/nodeid

