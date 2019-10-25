# ADR 016: Validator Consensus Key Rotation

## Changelog

- 2019 Oct 23: Initial draft

## Context

Validator consensus key rotation feature has been discussed and requested for a long time, for the sake of safer validator 
key management policy (e.g. https://github.com/tendermint/tendermint/issues/1136). So, we suggest one of the simplest form of
validator consensus key rotation implementation mostly onto Cosmos-SDK.

## Decision

### Pseudo procedure for consensus key rotation

- create new random consensus key.
- create and broadcast a transaction with a `MsgRotateConsPubKey` that states the new consensus key is now coupled with the validator operator with signature from the validator's operator key.
- old consensus key becomes unable to participate on consensus immediately after the update of key mapping state on-chain.
- start validating with new consensus key.
- validators using HSM and KMS should update the consensus key in HSM to use the new rotated key for signing votes.


### Considerations

- consensus key mapping information management strategy
    - store history of each key mapping changes in the kvstore.
    - the state machine can search corresponding consensus key paired with given validator operator for any arbitrary height in a recent unbonding period.
    - the state machine does not need any historical mapping information which is past more than unbonding period.
- limits
    - a validator cannot rotate its consensus key more than N time for any unbonding period, to prevent spam.
    - parameters can be decided by governance and stored in genesis file.
- slash module
    - slash module can search corresponding consensus key for any height so that it can decide which consensus key is supposed to be used for given height.
- abci.ValidatorUpdate
    - tendermint already has ability to change a consensus key by ABCI communication(`ValidatorUpdate`).
    - validator consensus key update can be done via creating new + delete old by change the power to zero.
    - therefore, we expect we even do not need to change tendermint codebase at all to implement this feature.


### Workflow

1. The validator generates a new consensus keypair.
2. The validator generates and signs a `MsgRotateConsPubKey` tx with their operator key and new ConsPubKey

    ```go
    type MsgRotateConsPubKey struct {
        ValidatorAddress  sdk.ValAddress
        NewPubKey         crypto.PubKey
    }
    ```

3. `handleMsgRotateConsPubKey` gets `MsgRotateConsPubKey`, calls `RotateConsPubKey` with emits event
4. `RotateConsPubKey` 
    - checks if `NewPubKey` is not duplicated on `ValidatorsByConsAddr`
    - overwrites `NewPubKey` in `validator.ConsPubKey`
    - deletes old `ValidatorByConsAddr`
    - `SetValidatorByConsAddr` for `NewPubKey`
    - Add `ConsPubKeyRotationHistory` for tracking rotation

    ```go
    type ConsPubKeyRotationHistory struct {
        OperatorAddress         sdk.ValAddress
        OldConsPubKey           crypto.PubKey
        NewConsPubKey           crypto.PubKey
        RotatedHeight           int64
    }
    ```

5. `ApplyAndReturnValidatorSetUpdates` checks if there is `ConsPubKeyRotationHistory` with `ConsPubKeyRotationHistory.RotatedHeight == ctx.BlockHeight()` and if so, generates 2 `ValidatorUpdate` , one for a remove validator and one for create new validator 

    ```go
    abci.ValidatorUpdate{
        PubKey: tmtypes.TM2PB.PubKey(OldConsPubKey),
        Power:  0,
    }

    abci.ValidatorUpdate{
        PubKey: tmtypes.TM2PB.PubKey(NewConsPubKey),
        Power:  v.ConsensusPower(),
    }
    ```

6. at `previousVotes` Iteration logic of `AllocateTokens`,  `previousVote` using `OldConsPubKey` match up with `ConsPubKeyRotationHistory`, and replace validator for token allocation
7. Migrate `ValidatorSigningInfo` and `ValidatorMissedBlockBitArray` from `OldConsPubKey` to `NewConsPubKey`


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
