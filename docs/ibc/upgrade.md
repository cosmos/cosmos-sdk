# Upgrading IBC-connected Chains

IBC-connnected chains must be able to upgrade without breaking connections to other chains. Otherwise there would be a massive disincentive towards upgrading and disrupting high-value IBC connections, thus preventing chains in the IBC ecosystem from evolving and improving. Many chain upgrades may be irrelevant to IBC, however some upgrades could potentially break counterparty clients if not handled correctly. Thus, any IBC chain that wishes to perform a IBC-client-breaking upgrade must perform an IBC upgrade in order to allow counterparty clients to securely upgrade to the new light client.

### IBC Client Breaking Upgrades

The current IBC protocol supports upgrading tendermint chains for a specific subset of IBC-client-breaking upgrades. Here is the exhaustive list of IBC client-breaking upgrades and whether the IBC protocol currently supports such upgrades.

Note: Since upgrades are only implemented for Tendermint clients, this doc only discusses upgrades on Tendermint chains that would break counterparty IBC Tendermint Clients.

1. Changing the Chain-ID: **Supported**
2. Changing the UnbondingPeriod: **Partially Supported**, chains may increase the unbonding period with no issues. However, decreasing the unbonding period may irreversibly break some counterparty clients. Thus, it is **not recommended** that chains reduce the unbonding period.
3. Changing the height (resetting to 0): **Supported**, so long as chains remember to increment the revision number in their chain-id.
4. Changing the ProofSpecs: **Supported**, this should be changed if the proof structure needed to verify IBC proofs is changed across the upgrade. Ex: Switching from an IAVL store, to a SimpleTree Store
5. Changing the UpgradePath: **Supported**, this might involve changing the key under which upgraded clients and consensus states are stored in the upgrade store, or even migrating the upgrade store itself.
6. Migrating the IBC store: **Unsupported**, as the IBC store location is negotiated by the connection.
7. Upgrading to a IBC path-breaking version of IBC: **Unsupported**, as IBC version is negotiated on connection handshake.
8. Changing the Tendermint LightClient algorithm: **Partially Supported**. Changes to the light client algorithm that do not change the ClientState or ConsensusState struct may be supported, provided that the counterparty is also upgraded to support the new light client algorithm. Changes that require updating the ClientState and ConsensusState structs themselves are theoretically possible by providing a path to translate an older ClientState struct into the new ClientState struct; however this is not currently implemented.

### Upgrading Chains

The IBC protocol allows client implementations to provide a path to upgrading clients given the upgraded client state, upgraded consensus state and proofs for each.

```go
// Upgrade functions
// NOTE: proof heights are not included as upgrade to a new revision is expected to pass only on the last
// height committed by the current revision. Clients are responsible for ensuring that the planned last
// height of the current revision is somehow encoded in the proof verification process.
// This is to ensure that no premature upgrades occur, since upgrade plans committed to by the counterparty
// may be cancelled or modified before the last planned height.
VerifyUpgradeAndUpdateState(
    ctx sdk.Context,
    cdc codec.BinaryMarshaler,
    store sdk.KVStore,
    newClient ClientState,
    newConsState ConsensusState,
    proofUpgradeClient,
    proofUpgradeConsState []byte,
) (upgradedClient ClientState, upgradedConsensus ConsensusState, err error)
```

The Tendermint client is the only implemented client that implements a non-trivial upgrade process. The Tendermint client implementations allow clients to specify an upgrade path which chains can use to commit to the upgraded client and consensus state before the upgrade occurs. IBC clients will then accept the upgraded client and consensus state if their associated proofs successfully verify against the upgrade path specified in the client.

While the upgrade path may vary from chain to chain, the Tendermint client expects the upgraded client to committed under key: `{upgradePath}/upgradedClient` and the upgraded consensus state is committed under key: `{upgradePath}/upgradedConsState`

The SDK supports the Tendermint IBC client process through the `x/upgrade` module. The upgrade module gives the chain the ability to vote on an `UpgradePlan` through a `SoftwareUpgradeProposal`. If the upgrade will break counterparty IBC clients, then the `UpgradePlan` **MUST** specify an `UpgradedClient` that counterparties can upgrade to. IBC chains planning an upgrade must also specify an upgrade height in the `UpgradePlan` rather than an upgrade time, since this is simpler for counterparty clients to verify.

The upgrade module in the SDK will then commit to the `UpgradedClient` at the key: `upgrade/UpgradedIBCState/upgradedClient`. 


