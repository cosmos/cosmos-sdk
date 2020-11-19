set -e
echo ".simd not found... initting a new chain"
mkdir config
# init config files
simd init simd --chain-id testing
# configure cli
simcli config chain-id testing
simcli config output json
simcli config trust-node true

# use keyring backend
simcli config keyring-backend test

# create accounts
simcli keys add fd

# give the accounts some money
simd add-genesis-account $(simcli keys show fd -a) 1000000000000stake

# save configs for the daemon
simd gentx --name fd --keyring-backend test --amount 10000000000stake

# input genTx to the genesis file
simd collect-gentxs
# verify genesis file is fine
simd validate-genesis
