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


