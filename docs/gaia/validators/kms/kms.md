# KMS - Key Management System

....

## What is a KMS?

...

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

The KMS provides different alternatives

- [Using a CPU-based signer](kms_cpu.md)
- [Using a YubiHSM](kms_ledger.md)
- [Using a Ledger device running the Tendermint Validator app](kms_ledger.md)
  