#!/bin/sh

set -e

simcli tx send cosmos17mmasmn306ztay8lx5kfyvqkv55anha7fpf4sq "$1" 100stake --chain-id="testing" --node tcp://cosmos:26657 --yes