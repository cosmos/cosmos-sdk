# Security

> **IMPORTANT**: If you find a security issue, you can contact our team directly at
security@tendermint.com, or report it to our [bug bounty program](https://hackerone.com/tendermint) on HackerOne. *DO NOT* open a public issue on the repository.

## Bug Bounty

As part of our [Coordinated Vulnerability Disclosure Policy](https://tendermint.com/security), we operate a
[bug bounty program](https://hackerone.com/tendermint) with Hacker One.

See the policy linked above for more details on submissions and rewards and read
this [blog post](https://blog.cosmos.network/bug-bounty-program-for-tendermint-cosmos-833c67693586) for the program scope.

The following is a list of examples of the kinds of bugs we're most interested
in for the Cosmos SDK. See [here](https://github.com/tendermint/tendermint/blob/master/SECURITY.md) for vulnerabilities we are interested
in for Tendermint and other lower-level libraries (eg. [IAVL](https://github.com/tendermint/iavl)).

### Core packages

- [`/baseapp`](https://github.com/cosmos/cosmos-sdk/tree/master/baseapp)
- [`/crypto`](https://github.com/cosmos/cosmos-sdk/tree/master/crypto)
- [`/types`](https://github.com/cosmos/cosmos-sdk/tree/master/types)
- [`/store`](https://github.com/cosmos/cosmos-sdk/tree/master/store)

### Modules

- [`x/auth`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth)
- [`x/bank`](https://github.com/cosmos/cosmos-sdk/tree/master/x/bank)
- [`x/capability`](https://github.com/cosmos/cosmos-sdk/tree/master/x/capability)
- [`x/staking`](https://github.com/cosmos/cosmos-sdk/tree/master/x/staking)
- [`x/slashing`](https://github.com/cosmos/cosmos-sdk/tree/master/x/slashing)
- [`x/evidence`](https://github.com/cosmos/cosmos-sdk/tree/master/x/evidence)
- [`x/distribution`](https://github.com/cosmos/cosmos-sdk/tree/master/x/distribution)
- [`x/ibc`](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc)
- [`x/ibc-transfer`](https://github.com/cosmos/cosmos-sdk/tree/master/x/ibc-transfer)
- [`x/mint`](https://github.com/cosmos/cosmos-sdk/tree/master/x/mint)

We are interested in bugs in other modules, however the above are most likely to
have significant vulnerabilities, due to the complexity / nuance involved. We
also recommend you to read the [specification](https://github.com/cosmos/cosmos-sdk/blob/master/docs/building-modules/README.md) of each module before digging into
the code.

### How we process Tx parameters

- Integer operations on tx parameters, especially `sdk.Int` / `sdk.Dec`
- Gas calculation & parameter choices
- Tx signature verification (see [`x/auth/ante`](https://github.com/cosmos/cosmos-sdk/tree/master/x/auth/ante))
- Possible Node DoS vectors (perhaps due to gas weighting / non constant timing)

### Handling private keys

- HD key derivation, local and Ledger, and all key-management functionality
- Side-channel attack vectors with our implementations
  - e.g. key exfiltration based on time or memory-access patterns when decrypting privkey
