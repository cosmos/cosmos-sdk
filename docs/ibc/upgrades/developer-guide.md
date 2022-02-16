<!--
order: 2
-->

# IBC Client Developer Guide to Upgrades

Learn how to implement upgrade functionality for your custom IBC client. {synopsis}

As mentioned in the [README](./README.md), it is vital that high-value IBC clients can upgrade along with their underlying chains to avoid disruption to the IBC ecosystem. Thus, IBC client developers will want to implement upgrade functionality to enable clients to maintain connections and channels even across chain upgrades.

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
    cdc codec.BinaryCodec,
    store sdk.KVStore,
    newClient ClientState,
    newConsState ConsensusState,
    proofUpgradeClient,
    proofUpgradeConsState []byte,
) (upgradedClient ClientState, upgradedConsensus ConsensusState, err error)
```

Note that the clients should have prior knowledge of the merkle path that the upgraded client and upgraded consensus states will use. The height at which the upgrade has occurred should also be encoded in the proof. The Tendermint client implementation accomplishes this by including an `UpgradePath` in the ClientState itself, which is used along with the upgrade height to construct the merkle path under which the client state and consensus state are committed.

Developers must ensure that the `UpgradeClientMsg` does not pass until the last height of the old chain has been committed, and after the chain upgrades, the `UpgradeClientMsg` should pass once and only once on all counterparty clients.

Developers must ensure that the new client adopts all of the new Client parameters that must be uniform across every valid light client of a chain (chain-chosen parameters), while maintaining the Client parameters that are customizable by each individual client (client-chosen parameters) from the previous version of the client.

Upgrades must adhere to the IBC Security Model. IBC does not rely on the assumption of honest relayers for correctness. Thus users should not have to rely on relayers to maintain client correctness and security (though honest relayers must exist to maintain relayer liveness). While relayers may choose any set of client parameters while creating a new `ClientState`, this still holds under the security model since users can always choose a relayer-created client that suits their security and correctness needs or create a Client with their desired parameters if no such client exists.

However, when upgrading an existing client, one must keep in mind that there are already many users who depend on this client's particular parameters. We cannot give the upgrading relayer free choice over these parameters once they have already been chosen. This would violate the security model since users who rely on the client would have to rely on the upgrading relayer to maintain the same level of security. Thus, developers must make sure that their upgrade mechanism allows clients to upgrade the chain-specified parameters whenever a chain upgrade changes these parameters (examples in the Tendermint client include `UnbondingPeriod`, `ChainID`, `UpgradePath`, etc.), while ensuring that the relayer submitting the `UpgradeClientMsg` cannot alter the client-chosen parameters that the users are relying upon (examples in Tendermint client include `TrustingPeriod`, `TrustLevel`, `MaxClockDrift`, etc).

Developers should maintain the distinction between Client parameters that are uniform across every valid light client of a chain (chain-chosen parameters), and Client parameters that are customizable by each individual client (client-chosen parameters); since this distinction is necessary to implement the `ZeroCustomFields` method in the `ClientState` interface:

```go
// Utility function that zeroes out any client customizable fields in client state
// Ledger enforced fields are maintained while all custom fields are zero values
// Used to verify upgrades
ZeroCustomFields() ClientState
```

Counterparty clients can upgrade securely by using all of the chain-chosen parameters from the chain-committed `UpgradedClient` and preserving all of the old client-chosen parameters. This enables chains to securely upgrade without relying on an honest relayer, however it can in some cases lead to an invalid final `ClientState` if the new chain-chosen parameters clash with the old client-chosen parameter. This can happen in the Tendermint client case if the upgrading chain lowers the `UnbondingPeriod` (chain-chosen) to a duration below that of a counterparty client's `TrustingPeriod` (client-chosen). Such cases should be clearly documented by developers, so that chains know which upgrades should be avoided to prevent this problem. The final upgraded client should also be validated in `VerifyUpgradeAndUpdateState` before returning to ensure that the client does not upgrade to an invalid `ClientState`.
