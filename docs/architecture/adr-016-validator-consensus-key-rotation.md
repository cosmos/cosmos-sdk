# ADR 016: Validator consensus key rotation

## Changelog

- 23-10-2019: initial draft

## Context

Validator consensus key rotation feature has been discussed and requested for a long time, for the sake of safer validator 
key management policy (e.g. https://github.com/tendermint/tendermint/issues/1136). So, we suggest one of the simplest form of
validator consensus key rotation implementation mostly onto Cosmos-SDK.

## Decision

### Pseudo procedure for consensus key rotation

- create new random consensus key.
- create and broadcast a transaction(RotateValConsensusKey) that the new consensus key is now coupled with the validator operator with signature from validator wallet key.
- old consensus key becomes unable to participate on consensus immediately after the update of key mapping state on-chain.
- start validating with new consensus key.

### Considerations

- consensus key mapping information management strategy
    - store history of each key mapping changes in the kvstore.
    - the state machine can search corresponding consensus key paired with given validator operator for any arbitrary height in a recent unbonding period.
    - the state machine does not need any historical mapping information which is past more than unbonding period.
- limits
    - a validator cannot rotate its consensus key more than N time for any unbonding period, to prevent spams.
    - a validator should contribute X atoms to community fund to rotate its consensus key, also to prevent spams.
    - parameters can be decided by governance and stored in genesis file.
- slash module
    - slash module can search corresponding consensus key for any height so that it can decide which consensus key is supposed to be used for given height.
- data pruning
    - database does not need to keep the historical key mapping changes after unbonding period past.
    - when pruning is on, a node prunes all historical key mapping changes past unbonding period.
- further implementation
    - introducing "validator control key" which has authority to sign transaction for given validator.
    - validator control key can be replaced to another key by transaction(RotateValControlKey).

### Special note on implementation

- tendermint already has ability to change a consensus key by ABCI communication(`ValidatorUpdate`).
- validator consensus key update can be done via creating new + delete old by change the power to zero.
- therefore, we expect we even do not need to change tendermint codebase at all to implement this feature.

## Status

Proposed

## Consequences

### Positive

- Validators can immediately or periodically rotate their consensus key to have better security policy

### Negative

- Slash module needs more computation because it needs to lookup corresponding consensus key of validators for each height

### Neutral

## References

- on tendermint repo : https://github.com/tendermint/tendermint/issues/1136
- on cosmos-sdk repo : https://github.com/cosmos/cosmos-sdk/issues/5231
- about multiple consensus keys : https://github.com/tendermint/tendermint/issues/1758#issuecomment-545291698
