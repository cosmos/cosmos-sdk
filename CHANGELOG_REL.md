### Client Breaking

- (x/auth) [\#6054](https://github.com/cosmos/cosmos-sdk/pull/6054) Remove custom JSON marshaling for base accounts as multsigs cannot be bech32 decoded.
- (modules) [\#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) The `/bank/balances/{address}` endpoint now returns all account
  balances or a single balance by denom when the `denom` query parameter is present.
- (client) [\#5640](https://github.com/cosmos/cosmos-sdk/pull/5640) The rest server endpoint `/swagger-ui/` is replaced by ´/´.
- (x/auth) [\#5702](https://github.com/cosmos/cosmos-sdk/pull/5702) The `x/auth` querier route has changed from `"acc"` to `"auth"`.
- (store/types) [\#5730](https://github.com/cosmos/cosmos-sdk/pull/5730) store.types.Cp() is removed in favour of types.CopyBytes().
- (client) [\#5640](https://github.com/cosmos/cosmos-sdk/issues/5783) Unify all coins representations on JSON client requests for governance proposals.
- [\#5785](https://github.com/cosmos/cosmos-sdk/issues/5785) JSON strings coerced to valid UTF-8 bytes at JSON marshalling time
  are now replaced by human-readable expressions. This change can potentially break compatibility with all those client side tools
  that parse log messages.
- (client) [\#5799](https://github.com/cosmos/cosmos-sdk/pull/5799) The `tx encode/decode` commands, due to change on encoding break compatibility with
  older clients.
- (x/auth) [\#5844](https://github.com/cosmos/cosmos-sdk/pull/5844) `tx sign` command now returns an error when signing is attempted with offline/multisig keys.
- (x/auth) [\#6108](https://github.com/cosmos/cosmos-sdk/pull/6108) `tx sign` command's `--validate-signatures` flag is migrated into a `tx validate-signatures` standalone command.
- (client/keys) [\#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) Remove `keys update` command.
- (x/evidence) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove CLI and REST handlers for querying `x/evidence` parameters.
- (server) [\#5982](https://github.com/cosmos/cosmos-sdk/pull/5982) `--pruning` now must be set to `custom` if you want to customise the granular options.

### API Breaking Changes

- [\#6212](https://github.com/cosmos/cosmos-sdk/pull/6212) Remove `Get*` prefixes from key construction functions
- [\#6079](https://github.com/cosmos/cosmos-sdk/pull/6079) Remove `UpgradeOldPrivValFile` (deprecated in Tendermint Core v0.28).
- (modules) [\#5664](https://github.com/cosmos/cosmos-sdk/pull/5664) Remove amino `Codec` from simulation `StoreDecoder`, which now returns a function closure in order to unmarshal the key-value pairs.
- (x/auth) [\#6029](https://github.com/cosmos/cosmos-sdk/pull/6029) Module accounts have been moved from `x/supply` to `x/auth`.
- (x/supply) [\#6010](https://github.com/cosmos/cosmos-sdk/pull/6010) All `x/supply` types and APIs have been moved to `x/bank`.
- (baseapp) [\#5865](https://github.com/cosmos/cosmos-sdk/pull/5865) The `SimulationResponse` returned from tx simulation is now JSON encoded instead of Amino binary.
- [\#5719](https://github.com/cosmos/cosmos-sdk/pull/5719) Bump Go requirement to 1.14+
- (x/params) [\#5619](https://github.com/cosmos/cosmos-sdk/pull/5619) The `x/params` keeper now accepts a `codec.Marshaller` instead of
  a reference to an amino codec. Amino is still used for JSON serialization.
- (types) [\#5579](https://github.com/cosmos/cosmos-sdk/pull/5579) The `keepRecent` field has been removed from the `PruningOptions` type.
  The `PruningOptions` type now only includes fields `KeepEvery` and `SnapshotEvery`, where `KeepEvery`
  determines which committed heights are flushed to disk and `SnapshotEvery` determines which of these
  heights are kept after pruning. The `IsValid` method should be called whenever using these options. Methods
  `SnapshotVersion` and `FlushVersion` accept a version arugment and determine if the version should be
  flushed to disk or kept as a snapshot. Note, `KeepRecent` is automatically inferred from the options
  and provided directly the IAVL store.
- (modules) [\#5555](https://github.com/cosmos/cosmos-sdk/pull/5555) Move `x/auth/client/utils/` types and functions to `x/auth/client/`.
- (modules) [\#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) Move account balance logic and APIs from `x/auth` to `x/bank`.
- (types) [\#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Refactored `AppModuleBasic` and `AppModuleGenesis`
  to now accept a `codec.JSONMarshaler` for modular serialization of genesis state.
- (types/rest) [\#5779](https://github.com/cosmos/cosmos-sdk/pull/5779) Drop unused Parse{Int64OrReturnBadRequest,QueryParamBool}() functions.
- (keys) [\#5820](https://github.com/cosmos/cosmos-sdk/pull/5820/) Removed method CloseDB from Keybase interface.
- (baseapp) [\#5837](https://github.com/cosmos/cosmos-sdk/issues/5837) Transaction simulation now returns a `SimulationResponse` which contains the `GasInfo` and
  `Result` from the execution.
- (client/input) [\#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) Removal of unnecessary `GetCheckPassword`, `PrintPrefixed` functions.
- (client/keys) [\#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) Rename `NewKeyBaseFromDir()` -> `NewLegacyKeyBaseFromDir()`.
- (crypto) [\#5880](https://github.com/cosmos/cosmos-sdk/pull/5880) Merge `crypto/keys/mintkey` into `crypto`.
- (crypto/hd) [\#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) `crypto/keys/hd` moved to `crypto/hd`.
- (crypto/keyring):
  - [\#5866](https://github.com/cosmos/cosmos-sdk/pull/5866) Rename `crypto/keys/` to `crypto/keyring/`.
  - [\#5904](https://github.com/cosmos/cosmos-sdk/pull/5904) `Keybase` -> `Keyring` interfaces migration. `LegacyKeybase` interface is added in order
    to guarantee limited backward compatibility with the old Keybase interface for the sole purpose of migrating keys across the new keyring backends. `NewLegacy`
    constructor is provided [\#5889](https://github.com/cosmos/cosmos-sdk/pull/5889) to allow for smooth migration of keys from the legacy LevelDB based implementation
    to new keyring backends. Plus, the package and the new keyring no longer depends on the sdk.Config singleton. Please consult the package documentation for more
    information on how to implement the new `Keyring` interface.
  - [\#5858](https://github.com/cosmos/cosmos-sdk/pull/5858) Make Keyring store keys by name and address's hexbytes representation.
- (x/evidence) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove APIs for getting and setting `x/evidence` parameters. `BaseApp` now uses a `ParamStore` to manage Tendermint consensus parameters which is managed via the `x/params` `Substore` type.
- (export) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) `AppExporter` now returns ABCI consensus parameters to be included in marshaled exported state. These parameters must be returned from the application via the `BaseApp`.

### Features

- (rest) [\#6167](https://github.com/cosmos/cosmos-sdk/pull/6167) Support `max-body-bytes` CLI flag for the REST service.
- (x/ibc) [\#5588](https://github.com/cosmos/cosmos-sdk/pull/5588) Add [ICS 024 - Host State Machine Requirements](https://github.com/cosmos/ics/tree/master/spec/ics-024-host-requirements) subpackage to `x/ibc` module.
- (x/ibc) [\#5277](https://github.com/cosmos/cosmos-sdk/pull/5277) `x/ibc` changes from IBC alpha. For more details check the the [`x/ibc/spec`](https://github.com/cosmos/tree/master/x/ibc/spec) directory:
  - [ICS 002 - Client Semantics](https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics) subpackage
  - [ICS 003 - Connection Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-003-connection-semantics) subpackage
  - [ICS 004 - Channel and Packet Semantics](https://github.com/cosmos/ics/blob/master/spec/ics-004-channel-and-packet-semantics) subpackage
  - [ICS 005 - Port Allocation](https://github.com/cosmos/ics/blob/master/spec/ics-005-port-allocation) subpackage
  - [ICS 007 - Tendermint Client](https://github.com/cosmos/ics/blob/master/spec/ics-007-tendermint-client) subpackage
  - [ICS 020 - Fungible Token Transfer](https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer) module
  - [ICS 023 - Vector Commitments](https://github.com/cosmos/ics/tree/master/spec/ics-023-vector-commitments) subpackage
  - (ibc/ante) Implement IBC `AnteHandler` as per [ADR 15 - IBC Packet Receiver](https://github.com/cosmos/tree/master/docs/architecture/adr-015-ibc-packet-receiver.md).
  - (x/capability) [\#5828](https://github.com/cosmos/cosmos-sdk/pull/5828) Capability module integration as outlined in [ADR 3 - Dynamic Capability Store](https://github.com/cosmos/tree/master/docs/architecture/adr-003-dynamic-capability-store.md).
  - (x/params) [\#6005](https://github.com/cosmos/cosmos-sdk/pull/6005) Add new CLI command for querying raw x/params parameters by subspace and key.
  - (x/ibc) [\#5769](https://github.com/cosmos/cosmos-sdk/pull/5769) [ICS 009 - Loopback Client](https://github.com/cosmos/ics/tree/master/spec/ics-009-loopback-client) subpackage

### Bug Fixes

- (x/distribution) [\#6210](https://github.com/cosmos/cosmos-sdk/pull/6210) Register `MsgFundCommunityPool` in distribution amino codec.
- (x/staking) [\#6061](https://github.com/cosmos/cosmos-sdk/pull/6061) Allow a validator to immediately unjail when no signing info is present due to
  falling below their minimum self-delegation and never having been bonded. The validator may immediately unjail once they've met their minimum self-delegation.
- (types) [\#5741](https://github.com/cosmos/cosmos-sdk/issues/5741) Prevent ChainAnteDecorators() from panicking when empty AnteDecorator slice is supplied.
- (modules) [\#5569](https://github.com/cosmos/cosmos-sdk/issues/5569) `InitGenesis`, for the relevant modules, now ensures module accounts exist.
- (crypto/keyring) [\#5844](https://github.com/cosmos/cosmos-sdk/pull/5844) `Keyring.Sign()` methods no longer decode amino signatures when method receivers
  are offline/multisig keys.
- (x/auth) [\#5892](https://github.com/cosmos/cosmos-sdk/pull/5892) Add `RegisterKeyTypeCodec` to register new
  types (eg. keys) to the `auth` module internal amino codec.
- (rest) [\#5906](https://github.com/cosmos/cosmos-sdk/pull/5906) Fix an issue that make some REST calls panic when sending
  invalid or incomplete requests.
- (x/genutil) [\#5938](https://github.com/cosmos/cosmos-sdk/pull/5938) Fix `InitializeNodeValidatorFiles` error handling.
- (x/staking) [\#5949](https://github.com/cosmos/cosmos-sdk/pull/5949) Skip staking `HistoricalInfoKey` in simulations as headers are not exported.
- (x/auth) [\#5950](https://github.com/cosmos/cosmos-sdk/pull/5950) Fix `IncrementSequenceDecorator` to use is `IsReCheckTx` instead of `IsCheckTx` to allow account sequence incrementing.
- (client) [\#5964](https://github.com/cosmos/cosmos-sdk/issues/5964) `--trust-node` is now false by default - for real. Users must ensure it is set to true if they don't want to enable the verifier.

### State Machine Breaking

- (x/staking) [\#6061](https://github.com/cosmos/cosmos-sdk/pull/6061) Allow a validator to immediately unjail when no signing info is present due to
  falling below their minimum self-delegation and never having been bonded. The validator may immediately unjail once they've met their minimum self-delegation.
- (x/supply) [\#6010](https://github.com/cosmos/cosmos-sdk/pull/6010) Removed the `x/supply` module by merging the existing types and APIs into the `x/bank` module.
- (modules) [\#5572](https://github.com/cosmos/cosmos-sdk/pull/5572) Separate balance from accounts per ADR 004.
  - Account balances are now persisted and retrieved via the `x/bank` module.
  - Vesting account interface has been modified to account for changes.
  - Callers to `NewBaseVestingAccount` are responsible for verifying account balance in relation to
    the original vesting amount.
  - The `SendKeeper` and `ViewKeeper` interfaces in `x/bank` have been modified to account for changes.
- (x/staking) [\#5600](https://github.com/cosmos/cosmos-sdk/pull/5600) Migrate the `x/staking` module to use Protocol Buffers for state
  serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
  for JSON encoding.
  - `BondStatus` is now of type `int32` instead of `byte`.
  - Types of `int16` in the `Params` type are now of type `int32`.
  - Every reference of `crypto.Pubkey` in context of a `Validator` is now of type string. `GetPubKeyFromBech32` must be used to get the `crypto.Pubkey`.
  - The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
    provided is specified by `ModuleCdc`.
- (x/slashing) [\#5627](https://github.com/cosmos/cosmos-sdk/pull/5627) Migrate the `x/slashing` module to use Protocol Buffers for state
  serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
  for JSON encoding.
  - The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
    provided is specified by `ModuleCdc`.
- (x/distribution) [\#5610](https://github.com/cosmos/cosmos-sdk/pull/5610) Migrate the `x/distribution` module to use Protocol Buffers for state
  serialization instead of Amino. The exact codec used is `codec.HybridCodec` which utilizes Protobuf for binary encoding and Amino
  for JSON encoding.
  - `ValidatorHistoricalRewards.ReferenceCount` is now of types `uint32` instead of `uint16`.
  - `ValidatorSlashEvents` is now a struct with `slashevents`.
  - `ValidatorOutstandingRewards` is now a struct with `rewards`.
  - `ValidatorAccumulatedCommission` is now a struct with `commission`.
  - The `Keeper` constructor now takes a `codec.Marshaler` instead of a concrete Amino codec. This exact type
    provided is specified by `ModuleCdc`.
- (x/auth) [\#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Migrate the `x/auth` module to use Protocol Buffers for state
  serialization instead of Amino.
  - The `BaseAccount.PubKey` field is now represented as a Bech32 string instead of a `crypto.Pubkey`.
  - `NewBaseAccountWithAddress` now returns a reference to a `BaseAccount`.
  - The `x/auth` module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize accounts.
  - The `AccountRetriever` type now accepts a `Codec` in its constructor in order to know how to
    serialize accounts.
- (x/supply) [\#5533](https://github.com/cosmos/cosmos-sdk/pull/5533) Migrate the `x/supply` module to use Protocol Buffers for state
  serialization instead of Amino.
  - The `internal` sub-package has been removed in order to expose the types proto file.
  - The `x/supply` module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize `SupplyI` types.
  - The `SupplyI` interface has been modified to no longer return `SupplyI` on methods. Instead the
    concrete type's receiver should modify the type.
- (x/mint) [\#5634](https://github.com/cosmos/cosmos-sdk/pull/5634) Migrate the `x/mint` module to use Protocol Buffers for state
  serialization instead of Amino.
  - The `internal` sub-package has been removed in order to expose the types proto file.
- (x/evidence) [\#5634](https://github.com/cosmos/cosmos-sdk/pull/5634) Migrate the `x/evidence` module to use Protocol Buffers for state
  serialization instead of Amino.
  - The `internal` sub-package has been removed in order to expose the types proto file.
  - The module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize `Evidence` types.
  - The `MsgSubmitEvidence` message has been removed in favor of `MsgSubmitEvidenceBase`. The application-level
    codec must now define the concrete `MsgSubmitEvidence` type which must implement the module's `MsgSubmitEvidence`
    interface.
- (x/upgrade) [\#5659](https://github.com/cosmos/cosmos-sdk/pull/5659) Migrate the `x/upgrade` module to use Protocol
  Buffers for state serialization instead of Amino.
  - The `internal` sub-package has been removed in order to expose the types proto file.
  - The `x/upgrade` module now accepts a `codec.Marshaler` interface.
- (x/gov) [\#5737](https://github.com/cosmos/cosmos-sdk/pull/5737) Migrate the `x/gov` module to use Protocol
  Buffers for state serialization instead of Amino.
  - `MsgSubmitProposal` will be removed in favor of the application-level proto-defined `MsgSubmitProposal` which
    implements the `MsgSubmitProposalI` interface. Applications should extend the `NewMsgSubmitProposalBase` type
    to define their own concrete `MsgSubmitProposal` types.
  - The module now accepts a `Codec` interface which extends the `codec.Marshaler` interface by
    requiring a concrete codec to know how to serialize `Proposal` types.
- (codec) [\#5799](https://github.com/cosmos/cosmos-sdk/pull/5799) Now we favor the use of `(Un)MarshalBinaryBare` instead of `(Un)MarshalBinaryLengthPrefixed` in all cases that are not needed.
- (x/evidence) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Remove parameters from `x/evidence` genesis and module state. The `x/evidence` module now solely uses Tendermint consensus parameters to determine of evidence is valid or not.

### Improvements

- (baseapp) [\#6186](https://github.com/cosmos/cosmos-sdk/issues/6186) Support emitting events during `AnteHandler` execution.
- (x/auth) [\#5702](https://github.com/cosmos/cosmos-sdk/pull/5702) Add parameter querying support for `x/auth`.
- (types) [\#5581](https://github.com/cosmos/cosmos-sdk/pull/5581) Add convenience functions {,Must}Bech32ifyAddressBytes.
- (staking) [\#5584](https://github.com/cosmos/cosmos-sdk/pull/5584) Add util function `ToTmValidator` that converts a `staking.Validator` type to `*tmtypes.Validator`.
- (client) [\#5585](https://github.com/cosmos/cosmos-sdk/pull/5585) IBC additions:
  - Added `prove` flag for commitment proof verification.
  - Added `queryABCI` function that returns the full `abci.ResponseQuery` with inclusion merkle proofs.
- (types) [\#5585](https://github.com/cosmos/cosmos-sdk/pull/5585) IBC additions:
  - `Coin` denomination max lenght has been increased to 32.
  - Added `CapabilityKey` alias for `StoreKey` to match IBC spec.
- (server) [\#5709](https://github.com/cosmos/cosmos-sdk/pull/5709) There are two new flags for pruning, `--pruning-keep-every`
  and `--pruning-snapshot-every` as an alternative to `--pruning`. They allow to fine tune the strategy for pruning the state.
- (client) [\#5810](https://github.com/cosmos/cosmos-sdk/pull/5810) Added a new `--offline` flag that allows commands to be executed without an
  internet connection. Previously, `--generate-only` served this purpose in addition to only allowing txs to be generated. Now, `--generate-only` solely
  allows txs to be generated without being broadcasted and disallows Keybase use and `--offline` allows the use of Keybase but does not allow any
  functionality that requires an online connection.
- (types/module) [\#5724](https://github.com/cosmos/cosmos-sdk/issues/5724) The `types/module` package does no longer depend on `x/simulation`.
- (client) [\#5856](https://github.com/cosmos/cosmos-sdk/pull/5856) Added the possibility to set `--offline` flag with config command.
- (client) [\#5895](https://github.com/cosmos/cosmos-sdk/issues/5895) show config options in the config command's help screen.
- (types/rest) [\#5900](https://github.com/cosmos/cosmos-sdk/pull/5900) Add Check\*Error function family to spare developers from replicating tons of boilerplate code.
- (x/evidence) [\#5952](https://github.com/cosmos/cosmos-sdk/pull/5952) Tendermint Consensus parameters can now be changed via parameter change proposals through `x/gov`.
- (x/evidence) [\#5961](https://github.com/cosmos/cosmos-sdk/issues/5961) Add `StoreDecoder` simulation for evidence module.
- (x/auth/ante) [\#6040](https://github.com/cosmos/cosmos-sdk/pull/6040) `AccountKeeper` interface used for `NewAnteHandler` and handler's decorators to add support of using custom `AccountKeeper` implementations.
- (simulation) [\#6002](https://github.com/cosmos/cosmos-sdk/pull/6002) Add randomized consensus params into simulation.
- (x/staking) [\#6059](https://github.com/cosmos/cosmos-sdk/pull/6059) Updated `HistoricalEntries` parameter default to 100.
- (x/ibc) [\#5948](https://github.com/cosmos/cosmos-sdk/issues/5948) Add `InitGenesis` and `ExportGenesis` functions for `ibc` module.
- (types) [\#6128](https://github.com/cosmos/cosmos-sdk/pull/6137) Add `String()` method to `GasMeter`.
- (types) [\#6195](https://github.com/cosmos/cosmos-sdk/pull/6195) Add codespace to broadcast(sync/async) response.
