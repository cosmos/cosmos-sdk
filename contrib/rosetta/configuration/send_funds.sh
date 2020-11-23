#!/bin/sh

set -e

simcli tx send cosmos1yk9hdjpfeu7c9yfzapa3k4qk7lhksdsu9zt5ea "$1" 100stake --chain-id="testing" --node tcp://cosmos:26657 --yes