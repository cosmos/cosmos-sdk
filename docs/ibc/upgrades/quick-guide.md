<!--
order: 1
-->

# How to Upgrade IBC Chains and their Clients

Learn how to upgrade your chain and counterparty clients. {synopsis}

The information in this doc for upgrading chains is relevant to Cosmos SDK chains. However, the guide for counterparty clients is relevant to any Tendermint client that enables upgrades.

## IBC Client Breaking Upgrades

IBC-connected chains must perform an IBC upgrade if their upgrade will break counterparty IBC clients. The current IBC protocol supports upgrading tendermint chains for a specific subset of IBC-client-breaking upgrades. Here is the exhaustive list of IBC client-breaking upgrades and whether the IBC protocol currently supports such upgrades.

IBC currently does **NOT** support unplanned upgrades. All of the following upgrades must be planned and committed to in advance by the upgrading chain, in order for counterparty clients to maintain their connections securely.

Note: Since upgrades are only implemented for Tendermint clients, this doc only discusses upgrades on Tendermint chains that would break counterparty IBC Tendermint Clients.

1. Changing the Chain-ID: **Supported**
2. Changing the UnbondingPeriod: **Partially Supported**, chains may increase the unbonding period with no issues. However, decreasing the unbonding period may irreversibly break some counterparty clients. Thus, it is **not recommended** that chains reduce the unbonding period.
3. Changing the height (resetting to 0): **Supported**, so long as chains remember to increment the revision number in their chain-id.
4. Changing the ProofSpecs: **Supported**, this should be changed if the proof structure needed to verify IBC proofs is changed across the upgrade. Ex: Switching from an IAVL store, to a SimpleTree Store
5. Changing the UpgradePath: **Supported**, this might involve changing the key under which upgraded clients and consensus states are stored in the upgrade store, or even migrating the upgrade store itself.
6. Migrating the IBC store: **Unsupported**, as the IBC store location is negotiated by the connection.
7. Upgrading to a backwards compatible version of IBC: Supported
8. Upgrading to a non-backwards compatible version of IBC: **Unsupported**, as IBC version is negotiated on connection handshake.
9. Changing the Tendermint LightClient algorithm: **Partially Supported**. Changes to the light client algorithm that do not change the ClientState or ConsensusState struct may be supported, provided that the counterparty is also upgraded to support the new light client algorithm. Changes that require updating the ClientState and ConsensusState structs themselves are theoretically possible by providing a path to translate an older ClientState struct into the new ClientState struct; however this is not currently implemented.

## Step-by-Step Upgrade Process for Cosmos SDK chains

If the IBC-connected chain is conducting an upgrade that will break counterparty clients, it must ensure that the upgrade is first supported by IBC using the list above and then execute the upgrade process described below in order to prevent counterparty clients from breaking.

1. Create an `UpgradeProposal` with an IBC ClientState in the `UpgradedClientState` field and a `UpgradePlan` in the `Plan` field. Note that the proposal `Plan` must specify an upgrade height **only** (no upgrade time), and the `ClientState` should only include the fields common to all valid clients and zero out any client-customizable fields (such as TrustingPeriod).
2. Vote on and pass the `UpgradeProposal`

Upon the `UpgradeProposal` passing, the upgrade module will commit the UpgradedClient under the key: `upgrade/UpgradedIBCState/{upgradeHeight}/upgradedClient`. On the block right before the upgrade height, the upgrade module will also commit an initial consensus state for the next chain under the key: `upgrade/UpgradedIBCState/{upgradeHeight}/upgradedConsState`.

Once the chain reaches the upgrade height and halts, a relayer can upgrade the counterparty clients to the last block of the old chain. They can then submit the proofs of the `UpgradedClient` and `UpgradedConsensusState` against this last block and upgrade the counterparty client.

## Step-by-Step Upgrade Process for Relayers Upgrading Counterparty Clients

Once the upgrading chain has committed to upgrading, relayers must wait till the chain halts at the upgrade height before upgrading counterparty clients. This is because chains may reschedule or cancel upgrade plans before they occur. Thus, relayers must wait till the chain reaches the upgrade height and halts before they can be sure the upgrade will take place.

Thus, the upgrade process for relayers trying to upgrade the counterparty clients is as follows:

1. Wait for the upgrading chain to reach the upgrade height and halt
2. Query a full node for the proofs of `UpgradedClient` and `UpgradedConsensusState` at the last height of the old chain.
3. Update the counterparty client to the last height of the old chain using the `UpdateClient` msg.
4. Submit an `UpgradeClient` msg to the counterparty chain with the `UpgradedClient`, `UpgradedConsensusState` and their respective proofs.
5. Submit an `UpdateClient` msg to the counterparty chain with a header from the new upgraded chain.

The Tendermint client on the counterparty chain will verify that the upgrading chain did indeed commit to the upgraded client and upgraded consensus state at the upgrade height (since the upgrade height is included in the key). If the proofs are verified against the upgrade height, then the client will upgrade to the new client while retaining all of its client-customized fields. Thus, it will retain its old TrustingPeriod, TrustLevel, MaxClockDrift, etc; while adopting the new chain-specified fields such as UnbondingPeriod, ChainId, UpgradePath, etc. Note, this can lead to an invalid client since the old client-chosen fields may no longer be valid given the new chain-chosen fields. Upgrading chains should try to avoid these situations by not altering parameters that can break old clients. For an example, see the UnbondingPeriod example in the supported upgrades section.

The upgraded consensus state will serve purely as a basis of trust for future `UpdateClientMsgs` and will not contain a consensus root to perform proof verification against. Thus, relayers must submit an `UpdateClientMsg` with a header from the new chain so that the connection can be used for proof verification again.
