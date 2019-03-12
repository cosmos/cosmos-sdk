# KMS - Key Management System

[Tendermint KMS](https://github.com/tendermint/kms) is a key management service that allows separating key management from Tendermint nodes. In addition it provides other advantages such as:

- Improved security and risk management policies
- Unified API and support for various HSM (hardware security modules)
- Double signing protection (software or hardware based)

It is recommended that the KMS service runs in a separate physical hosts.

## Building

Detailed build instructions can be found [here](https://github.com/tendermint/kms#installation).

::: tip
When compiling the KMS, ensure you have enabled the applicable features:
:::

| Backend               | Recommended Command line              |
|-----------------------|---------------------------------------|
| YubiHSM               | ```cargo build --features yubihsm```  |
| Ledger+Tendermint App | ```cargo build --features ledgertm``` |

## Configuration

A KMS can be configured in various ways:

### Using a YubiHSM
  
  Detailed information on how to setup a KMS with YubiHSM2 can be found [here](https://github.com/tendermint/kms/blob/master/README.yubihsm.md)

### Using a Ledger device running the Tendermint app

  Detailed information on how to setup a KMS with Ledger Tendermint App can be found [here](kms_ledger.md)
