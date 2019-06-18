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
go run contrib/export/main.go v0.36 genesis_0_34.json [-source v0.34] > ~/.gaiad/genesis.json 
```

To build and run the binary:
```bash
make build-genesis-migrate
./build/migrate v0.36 genesis_0_34.json > genesis.json
```

The resulting genesis will be importable into the targeted version of the SDK.
