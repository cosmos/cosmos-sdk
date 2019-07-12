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
<appbinary> migrate v0.36 genesis_0_34.json [--time "2019-04-22T17:00:11Z"] [--chain-id test] > ~/.gaiad/genesis.json 
```
