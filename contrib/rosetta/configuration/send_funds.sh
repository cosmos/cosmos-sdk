#!/bin/sh

set -e

simcli tx send cosmos16qhazsqdrzw96d03h37rp0p4m06hu8tyam3fjs "$1" 100stake --chain-id="testing" --node tcp://cosmos:26657 --yes