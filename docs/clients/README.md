# Clients

This section explains contains information on clients for SDK based blockchain. 

>*NOTE*: This section is a WIP. 

## Light-client

Light-clients enable users to interact with your application without having to download the entire state history but with a good level of security. 

- [Overview of light clients](./lite/README.md)
- [Starting a light-client server](./lite/getting_started.md)
- [Light-client specification](./lite/specification.md)

## Other clients

- [Command-Line interface for SDK-based blockchain](./cli.md)
- [Service provider doc](./service-providers.md)

## Genesis upgrade

If you need to upgrade your node you could export the genesis and migrate it to the new version through this script:

```bash
go run contrib/export/main.go genesis v0.36 genesis_0_34.json [--time "2019-04-22T17:00:11Z"] [--chain-id test] > ~/.gaiad/genesis.json 
```

Example from old version:
```bash
# if using a 0.34.7 chain
git checkout v0.34.7 && make install

# export your current state into a genesis file
gaiad export > v0_34_7_genesis.json

# reset all data
gaiad unsafe-reset-all

# checkout the current release
git checkout v0.36.0 && make build-genesis-migrate

# modify the genesis to be compatible with 0.34
./build/migrate genesis v0.36 v0_34_7_genesis.json --time <genesis-start-time-rfc3339> --chain-id=<new-chain-id> > [path_to_genesis.json]
```

The resulting genesis will be importable into the targeted version of the dapp using the SDK.




### Practical testnet upgrade example

Given a fresh machine, with a working go 12.5 environment we can experiment on migrating by doing:

```bash
ACCOUNT=myexample
NETWORK=gaia-13k
mkdir -p ${GOPATH}/src/github.com/cosmos/cosmos-sdk
cd ${GOPATH}/src/github.com/cosmos
git clone https://github.com/cosmos/cosmos-sdk.git
cd cosmos-sdk
git checkout v0.33.2
make install
gaiad init ${ACCOUNT} --chain-id ${NETWORK}
curl https://raw.githubusercontent.com/cosmos/testnets/master/${NETWORK}/genesis.json -o ~/.gaiad/config/genesis.json
sed -i '/persistent_peers = ""/c\persistent_peers = "c24f496b951148697f8a24fd749786075c128f00@35.203.176.214:26656"' .gaiad/config/config.toml
gaiad start
[kill gaiad]

# First we need to upgrade from 0.33 to 0.34
gaiad export > ~/v0_33_2_genesis.json
git checkout v0.34.7
make install
python3 contrib/export/v0.33.x-to-v0.34.0.py --chain-id=gaia-13k-034 ~/v0_33_2_genesis.json > ~/.gaiad/config/genesis.json
gaiad start
[kill gaiad]

# Now let's upgrade from 0.34 to 0.36
gaiad export > ~/v0_34_7_genesis.json
git checkout master # TODO: replace with v0.36.0 once the tag is released
make build-genesis-migrate
./build/migrate genesis v0.36 ~/v0_34_7_genesis.json --time <genesis-start-time-rfc3339> --chain-id <new-chain-id> > ~/.gaiad/config/genesis.json
cd ${GOPATH}/src/github.com/cosmos

# Needed since v0.36.0 as gaia has been removed from the SDK repo
git clone https://github.com/cosmos/gaia.git
git checkout master # TODO: replace with proper tag when released
cd gaia
make install
gaiad start

```