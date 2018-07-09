# Security

As part of our [Coordinated Vulnerability Disclosure
Policy](https://tendermint.com/security), we operate a bug bounty.
See the policy for more details on submissions and rewards.

The following is a list of examples of the kinds of bugs we're most interested in for
the Cosmos SDK. See [here](https://github.com/tendermint/tendermint/blob/master/SECURITY.md) for vulnerabilities we are interested in for Tendermint, and lower-level libraries, e.g. IAVL.

## Modules 
- x/staking
- x/slashing
- x/types
- x/gov

We are interested in bugs in other modules, however the above are most likely to have 
significant vulnerabilities, due to the complexity / nuance involved

## How we process Tx parameters
- Integer operations on tx parameters, especially sdk.Int / sdk.Uint
- Gas calculation & parameter choices 
- Tx signature verification (code in x/auth/ante.go)
- Possible Node DoS vectors. (Perhaps due to Gas weighting / non constant timing)

## Handling private keys
- HD key derivation, local and Ledger, and all key-management functionality
- Side-channel attack vectors with our implementations
  - e.g. key exfiltration based on time or memory-access patterns when decrypting privkey
  
