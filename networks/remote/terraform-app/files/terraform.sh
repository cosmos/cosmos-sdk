#!/bin/bash
# Script to initialize a testnet settings on a server

#Usage: terraform.sh <testnet_name> <testnet_node_number>

#Add gaiad node number for remote identification
echo "$2" > /etc/gaiad-nodeid

#Create gaiad user
useradd -m -s /bin/bash gaiad

#Reload services to enable the gaiad service (note that the gaiad binary is not available yet)
systemctl daemon-reload
systemctl enable gaiad


