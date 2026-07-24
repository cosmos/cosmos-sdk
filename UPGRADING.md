# Upgrade Reference

This document provides a reference for upgrading from `v0.54.x` to `v0.55.x` of Cosmos SDK.

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.55.x/CHANGELOG.md).

The headline changes in this release are the removal of three legacy surfaces (`x/params`, `x/protocolpool`, and `SIGN_MODE_TEXTUAL`), a reworked app-side mempool interface, and validator consensus key rotation in `x/staking`. Key rotation ships enabled for every chain that upgrades — see [Validator Consensus Key Rotation](#validator-consensus-key-rotation) — and requires one line of wiring in `app.go`. Everything else in the new-features list (ML-DSA-65 keys, secp256k1eth keys, Block-STM, the application-side mempool ABCI) is opt-in.

## Table of Contents

* [Breaking Changes](#breaking-changes)
    * [CometBFT Upgrade](#cometbft-upgrade)
    * [Removed: x/params](#removed-xparams)
    * [Removed: x/protocolpool](#removed-xprotocolpool)
    * [Removed: SIGN_MODE_TEXTUAL](#removed-sign_mode_textual)
    * [Mempool Interface Changes](#mempool-interface-changes)
    * [Staking: Key Rotation Fee Pool Module Account](#staking-key-rotation-fee-pool-module-account)
    * [genutil: ExportGenesisFileWithTime Signature](#genutil-exportgenesisfilewithtime-signature)
    * [Upgrade Handler and Store Migrations](#upgrade-handler-and-store-migrations)
* [New Features and Non-Breaking Changes](#new-features-and-non-breaking-changes)
    * [Validator Consensus Key Rotation](#validator-consensus-key-rotation)
    * [ML-DSA-65 Validator Consensus Keys](#ml-dsa-65-validator-consensus-keys)
    * [ML-DSA-65 Account Keys](#ml-dsa-65-account-keys)
    * [secp256k1eth Validator Consensus Keys](#secp256k1eth-validator-consensus-keys)
    * [Block-STM Parallel Transaction Execution](#block-stm-parallel-transaction-execution)
    * [Application-Side Mempool ABCI Methods](#application-side-mempool-abci-methods)
* [Behavior Changes Affecting Dapps and Indexers](#behavior-changes-affecting-dapps-and-indexers)

## Breaking Changes

### CometBFT Upgrade

<!-- TODO(release): pin the final CometBFT tag here and in the prose below once it is cut. -->

Cosmos SDK v0.55 requires a new CometBFT release (v0.54.x shipped with CometBFT `v0.39.0`). The v0.55.0 release will pin the final tag in `go.mod`; bump your app's `go.mod` to match. Relevant changes in the new CometBFT version:

* Expanded `MaxSignatureSize` and per-validator `MaxCommitSigBytes` to accommodate post-quantum (ML-DSA-65) signatures.
* New ABCI mempool methods `InsertTx` and `ReapTxs`, including socket-transport support ([cometbft#5958](https://github.com/cometbft/cometbft/pull/5958)), used by the SDK's [application-side mempool](#application-side-mempool-abci-methods).
* Updated `DefaultBlockParams` ([cometbft#5987](https://github.com/cometbft/cometbft/pull/5987)). This changes defaults for new chains only; existing chains keep their on-chain consensus params.

See the [CometBFT changelog](https://github.com/cometbft/cometbft/blob/main/CHANGELOG.md) for the full list.

### Removed: x/params

[#25546](https://github.com/cosmos/cosmos-sdk/pull/25546) removes the `x/params` module entirely (only a tombstone README remains). Module parameters have been managed by each module since v0.47; v0.55 removes the leftover machinery:

1. If your app still imports `x/params` (a `paramskeeper.Keeper`, per-module `Subspace`s, or the legacy gov proposal handler), remove that wiring. If the `params` store is still mounted, delete it in your store upgrades (see [Upgrade Handler and Store Migrations](#upgrade-handler-and-store-migrations)). Chains that have not yet migrated legacy subspace params to module-managed params must complete that migration **before** upgrading to v0.55 — the migration code is gone.

2. Drop the trailing `exported.Subspace` argument (typically passed as `nil`) from the module constructors that carried it for legacy migrations:

```go
// Before                                                          // After
auth.NewAppModule(cdc, accountKeeper, randGenAccountsFn, nil)      auth.NewAppModule(cdc, accountKeeper, randGenAccountsFn)
bank.NewAppModule(cdc, bankKeeper, accountKeeper, nil)             bank.NewAppModule(cdc, bankKeeper, accountKeeper)
gov.NewAppModule(cdc, &govKeeper, accountKeeper, bankKeeper, nil)  gov.NewAppModule(cdc, &govKeeper, accountKeeper, bankKeeper)
mint.NewAppModule(cdc, mintKeeper, accountKeeper, nil, nil)        mint.NewAppModule(cdc, mintKeeper, accountKeeper, nil)
slashing.NewAppModule(cdc, keeper, ak, bk, sk, nil, registry)      slashing.NewAppModule(cdc, keeper, ak, bk, sk, registry)
distr.NewAppModule(cdc, keeper, ak, bk, stakingKeeper, nil)        distr.NewAppModule(cdc, keeper, ak, bk, stakingKeeper)
staking.NewAppModule(cdc, keeper, ak, bk, nil)                     staking.NewAppModule(cdc, keeper, ak, bk)
```

(`mint.NewAppModule` retains its deprecated `InflationCalculationFn` parameter; only the subspace argument is removed.)

### Removed: x/protocolpool

[#26421](https://github.com/cosmos/cosmos-sdk/pull/26421) removes the `x/protocolpool` module and its proto/API surface from the SDK. The `distrkeeper.WithExternalCommunityPool` extension point is removed with it — `x/distribution` always uses its internal `FeePool` community pool again, and `MsgFundCommunityPool` / `MsgCommunityPoolSpend` operate on it directly.

**Required action** if your app wired `x/protocolpool` (the v0.54 SimApp default):

1. Remove all `protocolpool` wiring from `app.go`: the imports, the `ProtocolPoolKeeper` field and its `NewKeeper` call, the `protocolpooltypes.ModuleName` and `protocolpooltypes.ProtocolPoolEscrowAccount` entries in `maccPerms`, the module manager entry, and its entries in the begin-block, end-block, init-genesis, and export orders.
2. Remove `distrkeeper.WithExternalCommunityPool(app.ProtocolPoolKeeper)` from your `distrkeeper.NewKeeper` call.
3. Delete the `protocolpool` store in your store upgrades (see [Upgrade Handler and Store Migrations](#upgrade-handler-and-store-migrations)).
4. Balances held by the protocolpool module accounts are bank state and are **not** migrated automatically. Decide where those funds go and move them in your upgrade handler — e.g. transfer them to the `x/distribution` community pool so community-pool spend proposals keep working.

If your app never wired `x/protocolpool`, no action is needed beyond not being able to import it.

### Removed: SIGN_MODE_TEXTUAL

`SIGN_MODE_TEXTUAL` (proto enum value `2`) and its entire implementation have been removed ([#26456](https://github.com/cosmos/cosmos-sdk/pull/26456)):

* `x/tx/signing/textual/` — all renderers, the CBOR encoder, test data, and internal protos
* `x/auth/tx/textual.go` and `ConfigOptions.TextualCoinMetadataQueryFn`
* Ledger + SIGN_MODE_TEXTUAL integration in `client/` flags and tx factory

The proto enum value `2` and string `"SIGN_MODE_TEXTUAL"` are **reserved** to prevent future reuse. ADR-050 is archived.

**Required action** if your app enabled SIGN_MODE_TEXTUAL:

1. Remove `TextualCoinMetadataQueryFn` from your `tx.ConfigOptions`:

    ```go
    // Before
    txConfig, err := tx.NewTxConfigWithOptions(cdc, tx.ConfigOptions{
        TextualCoinMetadataQueryFn: ...,
    })

    // After — field removed, omit it
    txConfig, err := tx.NewTxConfigWithOptions(cdc, tx.ConfigOptions{...})
    ```

2. Remove any `SIGN_MODE_TEXTUAL` cases from signing mode handler switch statements.

3. Remove Ledger wiring that depended on `SIGN_MODE_TEXTUAL`. Client-side root command wiring that constructed a textual-enabled tx config for online mode (as v0.54 SimApp did in `simd/cmd/root.go`) should be deleted as well.

### Mempool Interface Changes

[#25338](https://github.com/cosmos/cosmos-sdk/pull/25338) changes the `types/mempool` interfaces so the mempool stores the gas wanted reported by the ante handler at `CheckTx` time, and block selection uses that value instead of the tx-declared gas limit.

**Required action** if you implement a custom mempool (chains using the SDK's built-in mempools or no app-side mempool just recompile):

* `Insert` gains an `InsertOption` parameter carrying the ante-reported gas: `Insert(context.Context, sdk.Tx, InsertOption) error`.
* `Iterator.Tx()` now returns a `PooledTx` (`{Tx sdk.Tx; GasWanted uint64}`) instead of `sdk.Tx`.
* `ExtMempool.SelectBy`'s callback now receives a `PooledTx`: `SelectBy(context.Context, [][]byte, func(PooledTx) bool)`.
* `ExtMempool` gains `RemoveWithReason(context.Context, sdk.Tx, RemoveReason) error`, where `RemoveReason` identifies the caller (`run_tx.recheck`, `run_tx.finalize`, `prepare_proposal.remove_invalid`) and an optional error.

Custom `PrepareProposal` handlers that iterate the mempool should read gas from `PooledTx.GasWanted` rather than re-deriving it from the tx.

### Staking: Key Rotation Fee Pool Module Account

`x/staking` now requires a `key_rotation_fee_pool` module account with burn permissions — the staking keeper panics at construction if it is missing (`x/staking/keeper/keeper.go`). Add it to your `maccPerms`:

```go
maccPerms = map[string][]string{
    // ...existing entries...
    stakingtypes.KeyRotationFeePoolName: {authtypes.Burner},
}
```

This is required for **all** chains upgrading to v0.55, whether or not validators are expected to use [key rotation](#validator-consensus-key-rotation).

Two module consensus-version bumps ship in this release and run automatically via `RunMigrations` in your upgrade handler:

* `x/staking` 5 → 6: adds the `key_rotation_fee` param, defaulting to `1000000` of the bond denom ([#26485](https://github.com/cosmos/cosmos-sdk/pull/26485)). `Params.Validate` requires the fee denom to equal `bond_denom` ([#26613](https://github.com/cosmos/cosmos-sdk/pull/26613)).
* `x/auth` 6 → 7: adds the `SigVerifyCostMlDsa65` param with its default value ([#26472](https://github.com/cosmos/cosmos-sdk/pull/26472)).

### genutil: ExportGenesisFileWithTime Signature

[#26468](https://github.com/cosmos/cosmos-sdk/pull/26468) consolidates `ExportGenesisFileWithTime`'s arguments so the exported file preserves consensus params (previously they were rebuilt from defaults, dropping the caller's values):

```go
// Before
func ExportGenesisFileWithTime(genFile, chainID string, validators []cmttypes.GenesisValidator,
    appState json.RawMessage, genTime time.Time) error

// After — build the AppGenesis yourself; everything you set on it is preserved
func ExportGenesisFileWithTime(genFile string, appGenesis *types.AppGenesis, genTime time.Time) error
```

### Upgrade Handler and Store Migrations

A reference upgrade handler for this release (see `simapp/upgrades.go`):

```go
const UpgradeName = "v054-to-v055"

func (app SimApp) RegisterUpgradeHandlers() {
    app.UpgradeKeeper.SetUpgradeHandler(
        UpgradeName,
        func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
            return app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
        },
    )

    upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
    if err != nil {
        panic(err)
    }

    if upgradeInfo.Name == UpgradeName && !app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
        storeUpgrades := storetypes.StoreUpgrades{
            Added:   []string{},
            Deleted: []string{"protocolpool"},
        }
        app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &storeUpgrades))
    }
}
```

Add `"params"` to `Deleted` as well if your app still had the `x/params` store mounted.

## New Features and Non-Breaking Changes

These changes are optional to adopt during the upgrade; they are not required for a successful migration. The exception is key rotation, which is active on every v0.55 chain once the required wiring above is in place.

### Validator Consensus Key Rotation

v0.55 adds consensus key rotation to `x/staking` ([#26440](https://github.com/cosmos/cosmos-sdk/pull/26440)): a validator operator can submit `MsgRotateConsPubKey` (wired into the CLI, [#26461](https://github.com/cosmos/cosmos-sdk/pull/26461)) to replace their consensus key without unbonding. Key properties:

* **Fee.** Each rotation charges the `key_rotation_fee` staking param (default `1000000` of the bond denom) from the operator account; the fee is burned via the `key_rotation_fee_pool` module account.
* **Rate limit.** One rotation per validator per unbonding period.
* **Applied in the end blocker.** The rotation is scheduled by the msg server and applied at the end of the block; CometBFT is informed through a validator-set update.
* **Evidence and slashing.** Equivocation evidence against a rotated-away (historical) consensus address remains attributable to the validator until the evidence is no longer admissible — i.e. until both `evidence.max_age_num_blocks` and `evidence.max_age_duration` have elapsed since the rotation, which can be later than the unbonding time ([#26481](https://github.com/cosmos/cosmos-sdk/pull/26481), [#26616](https://github.com/cosmos/cosmos-sdk/pull/26616)). Slashing signing info is migrated to the active consensus key. Governance changes that extend the evidence-age params after a rotation's expiry has been computed are not retroactively applied; chains should account for this when tuning evidence params.
* **Genesis.** Rotation history and pending-rotation state are included in staking genesis import/export ([#26471](https://github.com/cosmos/cosmos-sdk/pull/26471)); genesis export tooling that parses staking genesis JSON should expect the new fields.
* **Events.** `rotate_cons_pubkey` is emitted when a rotation is scheduled (including apply height, maturity time, evidence-expiry time/height, and the burned fee) and `apply_cons_pubkey_rotation` when it is applied (validator, old and new consensus addresses) ([#26619](https://github.com/cosmos/cosmos-sdk/pull/26619)).

Indexers, exchanges, and monitoring that key validators by consensus address must handle the mapping changing over a validator's lifetime.

### ML-DSA-65 Validator Consensus Keys

Cosmos SDK v0.55 registers the NIST ML-DSA-65 (FIPS 204) post-quantum signature scheme as a supported validator consensus key type ([#26436](https://github.com/cosmos/cosmos-sdk/pull/26436)). The new `cosmos.crypto.mldsa65.PubKey` / `PrivKey` proto messages, Amino routes (`cometbft/PubKeyMlDsa65`, `cometbft/PrivKeyMlDsa65`), interface-registry registration, multisig amino route, and `hd.MlDsa65Type` constant are all enabled by default.

**Action required:** none. Existing chains continue to accept only the consensus key types listed in `genesis.consensus_params.validator.pub_key_types` (still `["ed25519"]` by default). No state-machine-relevant behavior changes for chains that do not opt in.

**To opt in (new chains):** set `genesis.consensus_params.validator.pub_key_types` to `["ml_dsa_65"]` (or a list including it). Validators must then submit `MsgCreateValidator` with a `mldsa65.PubKey`. The `init` and `testnet` commands accept `--consensus-key-algo ml_dsa_65` to generate matching validator files ([#26604](https://github.com/cosmos/cosmos-sdk/pull/26604)). Test harnesses can use the new `testutil/network.Config.ValidatorConsensusKeyType` field together with `genutil.InitializeNodeValidatorFilesFromMnemonicWithKeyType` to spin up an in-process testnet pinned to ML-DSA-65.

**Operational considerations:** ML-DSA-65 keys and signatures are substantially larger than ed25519 (pubkey 1952 bytes vs 32, signature 3309 bytes vs 64). Chains enabling this key type should review `consensus_params.block.max_bytes` and gossip framing limits accordingly. The CometBFT version this release depends on expanded `MaxSignatureSize` and the per-validator `MaxCommitSigBytes` to accommodate the larger signatures; downstream applications relying on the previous fixed values may need to be re-examined.

Existing chains can combine this with [key rotation](#validator-consensus-key-rotation) to move validators to post-quantum keys: add `ml_dsa_65` to `pub_key_types` via a consensus-params update, then have validators rotate.

### ML-DSA-65 Account Keys

[#26472](https://github.com/cosmos/cosmos-sdk/pull/26472) extends ML-DSA-65 support to user account keys: keyring creation and mnemonic recovery (`--algo ml_dsa_65`), transaction signing and verification, and a new ante-handler gas cost param `SigVerifyCostMlDsa65` (added to `x/auth` params by the automatic 6 → 7 migration). No action is required; accounts using existing key types are unaffected.

### secp256k1eth Validator Consensus Keys

[#26615](https://github.com/cosmos/cosmos-sdk/pull/26615) adds `crypto/keys/secp256k1eth`, wrapping CometBFT's Ethereum-style secp256k1 consensus key implementation with SDK codec registration. Intended for EVM-compatible chains that want validator consensus addresses derived the Ethereum way; opt in via `genesis.consensus_params.validator.pub_key_types`.

### Block-STM Parallel Transaction Execution

v0.55 adds operator-facing configuration for parallel block execution with Block-STM ([#26208](https://github.com/cosmos/cosmos-sdk/pull/26208)): `block-executor` (`"sequential"`, the default, or `"block-stm"`), `block-stm-workers`, and `block-stm-pre-estimate` in `app.toml`.

To wire it, use the new `baseapp/blockexec` helper after creating your store keys (see `simapp/app.go`):

```go
stores := make([]storetypes.StoreKey, 0, len(keys))
for _, k := range keys {
    stores = append(stores, k)
}
blockexec.Apply(bApp, appOpts, stores, txConfig.TxDecoder(),
    func(storetypes.MultiStore) string { return sdk.DefaultBondDenom },
)
```

`Apply` resolves the executor from `app.toml`/flags and installs the corresponding `TxRunner`; with the default `sequential` executor it preserves today's behavior, so the wiring is safe to add unconditionally. Block-STM requires the block gas meter to remain disabled (the default since v0.54) — enabling both panics at parameter assignment.

Switching a running chain's executor is a per-node setting with identical state-transition results, but treat the first enablement as an operational rollout: test with your workload before flipping validators.

### Application-Side Mempool ABCI Methods

[#25620](https://github.com/cosmos/cosmos-sdk/pull/25620) and [#25969](https://github.com/cosmos/cosmos-sdk/pull/25969) add support for CometBFT's new application-side mempool ABCI methods. `BaseApp` exposes `SetInsertTxHandler` and `SetReapTxsHandler`; when CometBFT runs with `mempool.type = "app"`, it delegates mempool insertion and reaping to the application via `InsertTx` / `ReapTxs`. Both handlers must be set — `BaseApp` returns an error for these ABCI calls otherwise. Requires the CometBFT release this SDK version is pinned to (older versions lack socket-transport support for the new methods).

Chains using CometBFT's built-in mempool need no changes.

## Behavior Changes Affecting Dapps and Indexers

Observable changes between v0.54.x and v0.55.x that don't require code changes but may affect downstream consumers:

* **Block selection uses ante-reported gas.** Proposals are packed using the gas wanted returned by the ante handler at `CheckTx` time rather than the tx-declared gas limit ([#25338](https://github.com/cosmos/cosmos-sdk/pull/25338)). Block composition can differ for txs whose ante-reported gas diverges from their declared limit.
* **Staking emits key-rotation events** (`rotate_cons_pubkey`, `apply_cons_pubkey_rotation`), and validator consensus addresses can change over time ([#26619](https://github.com/cosmos/cosmos-sdk/pull/26619)).
* **`x/gov` `proposal_messages` event attribute** no longer has a leading comma ([#26353](https://github.com/cosmos/cosmos-sdk/pull/26353)).
* **`x/authz` prunes at most 200 expired grants per begin block** ([#26588](https://github.com/cosmos/cosmos-sdk/pull/26588)); mass-expiry cleanup now spreads across blocks.
* **`x/distribution` reward withdrawals to blocked addresses** during begin/end block fall back to the delegator/validator owner and then the community pool instead of failing ([#26406](https://github.com/cosmos/cosmos-sdk/pull/26406)). User-initiated withdrawals to blocked addresses still return `ErrUnauthorized`.
* **`x/feegrant` `Allowances` and `AllowancesByGranter` queries** now honor `PageRequest.offset` and `count_total` correctly ([#26596](https://github.com/cosmos/cosmos-sdk/pull/26596)); clients that compensated for the old off-by-page results should re-check.
