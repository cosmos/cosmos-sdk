# ADR 033: Arbitrary Message Signature

## Changelog

- 28/10/2020 - Initial draft

## Authors
- Antoine Herzog (@antoineherzog)
- Zaki Manian (@zmanian)
- Frojdi Dymylja (@fdymylja)

## Status
Draft

## Abstract

Currently, in the SDK, there is no convention to sign arbitrary message like on Ethereum. We propose with this specification a way to sign arbitrary message for Cosmos SDK chains.

## Context
Having the ability to sign messages off-chain has proven to be a fundamental aspect of nearly any blockchain. The notion of signing messages off-chain has many added benefits such as saving on computational costs and reducing transaction throughput and overhead. Within the context of the Cosmos, some of the major applications of signing such data includes, but is not limited to, providing a cryptographic secure and verifiable means of proving validator identity and possibly associating it with some other framework or organization. In addition, having the ability to sign Cosmos messages with a Ledger or similar HSM device.


## Decision

The aim is being able to sign arbitrary messages, even using Ledger or similar HSM devices.

As a result signed messages should look roughly like Cosmos SDK messages. chain-id, account_number and sequence can all be assigned invalid values.
The CLI should set those to default values.

Cosmos SDK 0.40 also introduces a concept of “auth_info” this can specify SIGN_MODES.

A spec should include an auth info that supports SIGN_MODE_DIRECT and SIGN_MODE_LEGACY_AMINO.

- the memo should be empty
- nonce, sequence number should be equal to 0
- chain-id should be equal to “signature”
- fee gas should be equal to 20000 (0 is not applicable)
- fee amount should be an empty array
- Inside the message with the type MsgSignText, we put inside a  *bytes* message and the address of the signer.


## Consequences

There is a new optional module that offers arbitrary signing and verification capabilities.

### Backwards Compatibility

Backwards compatibility is maintained as this is a new module.

### Positive

- Ability to sign arbitrary messages

### Negative

- Probably this feature should be provided via the keystore 
- Developing it as a module means we need to rely on using sign transactions in offline mode, and implementing it as default for the command means that we need to do a flag override (offline, account number, sequence, chain id)


## References
