# Security

As part of our [Coordinated Vulnerability Disclosure
Policy](https://tendermint.com/security), we operate a bug bounty.
See the policy for more details on submissions and rewards.

The following is a list of examples of the kinds of bugs we're most interested in for
the cosmos-sdk. See [here](https://github.com/tendermint/tendermint/blob/master/SECURITY.md) for vulnerabilities we are interested in for tendermint / lower level libs.

## Specification
- Conceptual flaws
- Ambiguities, inconsistencies, or incorrect statements
- Mis-match between specification and implementation of any component

## Modules 
- x/staking
- x/slashing
- SDK standard datatype library

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

## Least capabilities system
- Attack vectors in our least capabilities system
- Scenarios where a chain runs a "Malicious module"
  - One example is a malicious module getting priviledge escalation to read
   a store which it doesn't have the key for

