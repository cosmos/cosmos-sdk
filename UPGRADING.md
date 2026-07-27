# Upgrade Reference

This document provides a reference for upgrading from `v0.54.x` to `v0.55.x` of Cosmos SDK. If you are upgrading directly from `v0.53.x`, see [Upgrading from v0.53.x](#upgrading-from-v053x) after reading the breaking changes below.

For a full list of changes, see the [Changelog](https://github.com/cosmos/cosmos-sdk/blob/release/v0.55.x/CHANGELOG.md).

The headline changes in this release are the removal of three legacy surfaces (`x/params`, `x/protocolpool`, and `SIGN_MODE_TEXTUAL`), a reworked app-side mempool interface, and validator consensus key rotation in `x/staking`. Key rotation ships enabled for every chain that upgrades — see [Validator Consensus Key Rotation](#validator-consensus-key-rotation) — and requires one line of wiring in `app.go`. Everything else in the new-features list (ML-DSA-65 keys, secp256k1eth keys, config-driven Block-STM wiring) is opt-in.

## Table of Contents

* [Breaking Changes](#breaking-changes)
    * [CometBFT Upgrade](#cometbft-upgrade)
    * [Removed: x/params](#removed-xparams)
    * [Removed: x/protocolpool](#removed-xprotocolpool)
    * [Removed: SIGN_MODE_TEXTUAL](#removed-sign_mode_textual)
    * [Mempool Interface Changes](#mempool-interface-changes)
    * [Staking: Key Rotation Wiring and Interface Changes](#staking-key-rotation-wiring-and-interface-changes)
    * [genutil: ExportGenesisFileWithTime Signature](#genutil-exportgenesisfilewithtime-signature)
    * [Upgrade Handler and Store Migrations](#upgrade-handler-and-store-migrations)
* [Upgrading from v0.53.x](#upgrading-from-v053x)
* [New Features and Non-Breaking Changes](#new-features-and-non-breaking-changes)
    * [Validator Consensus Key Rotation](#validator-consensus-key-rotation)
    * [ML-DSA-65 Validator Consensus Keys](#ml-dsa-65-validator-consensus-keys)
    * [ML-DSA-65 Account Keys](#ml-dsa-65-account-keys)
    * [secp256k1eth Validator Consensus Keys](#secp256k1eth-validator-consensus-keys)
    * [Block-STM Configuration](#block-stm-configuration)
* [Behavior Changes Affecting Dapps and Indexers](#behavior-changes-affecting-dapps-and-indexers)

## Breaking Changes

### CometBFT Upgrade

Cosmos SDK v0.55 requires CometBFT `v0.40.0` (the v0.54.x line shipped with `v0.39.x`, ending at `v0.39.3` in v0.54.3). Bump your app's `go.mod` to match the SDK's pin. Relevant changes in CometBFT v0.40.0:

* Expanded `MaxSignatureSize` and per-validator `MaxCommitSigBytes` to accommodate post-quantum (ML-DSA-65) signatures.
* A fix for the application-side mempool (`mempool.type = "app"`, supported since CometBFT v0.39.2 / SDK v0.54.3): the default socket transport was missing the `InsertTx` / `ReapTxs` cases, causing node self-kill ([cometbft#5958](https://github.com/cometbft/cometbft/pull/5958)). Chains using an app-side mempool over the socket transport need v0.40.0.
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
* `ExtMempool.RemoveWithReason` and the `RemoveReason` type, introduced in v0.54, are unchanged.

Custom `PrepareProposal` handlers that iterate the mempool should read gas from `PooledTx.GasWanted` rather than re-deriving it from the tx.

### Staking: Key Rotation Wiring and Interface Changes

`x/staking` now requires a `key_rotation_fee_pool` module account with burn permissions — the staking keeper panics at construction if it is missing (`x/staking/keeper/keeper.go`). Add it to your `maccPerms`:

```go
maccPerms = map[string][]string{
    // ...existing entries...
    stakingtypes.KeyRotationFeePoolName: {authtypes.Burner},
}
```

This is required for **all** chains upgrading to v0.55, whether or not validators are expected to use [key rotation](#validator-consensus-key-rotation).

Key rotation also touches two staking keeper surfaces that external modules may implement or consume:

* The `StakingHooks` interface gains `AfterValidatorConsKeyUpdated(ctx context.Context, oldConsAddr, newConsAddr sdk.ConsAddress, valAddr sdk.ValAddress) error`, called when a rotation is applied. Custom `StakingHooks` implementations must add this method (returning `nil` is fine if you don't need the notification).
* The staking keeper adds `ValidatorByHistoricalConsAddr(ctx, consAddr)`, which resolves a validator from a consensus address it used before a rotation. Modules that map consensus addresses to validators can no longer assume that mapping is immutable — see [Validator Consensus Key Rotation](#validator-consensus-key-rotation).

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

#### Module Migrations

Two module consensus-version bumps ship in this release and run automatically via `RunMigrations` in your upgrade handler:

* `x/staking` 5 → 6: adds the `key_rotation_fee` param, defaulting to `1000000` of the bond denom ([#26485](https://github.com/cosmos/cosmos-sdk/pull/26485)). `Params.Validate` requires the fee denom to equal `bond_denom` ([#26613](https://github.com/cosmos/cosmos-sdk/pull/26613)).
* `x/auth` 6 → 7: adds the `SigVerifyCostMlDsa65` param with its default value ([#26472](https://github.com/cosmos/cosmos-sdk/pull/26472)).

#### Reference Upgrade Handler

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

## Upgrading from v0.53.x

Skipping v0.54 and upgrading directly from `v0.53.x` to `v0.55.x` is supported as a single coordinated upgrade: one binary swap, one upgrade handler, one halt height. Work through the [v0.53.x → v0.54.x upgrade reference](https://github.com/cosmos/cosmos-sdk/blob/release/v0.54.x/UPGRADING.md) first — all of its required changes still apply — then apply this guide on top. The v0.54 hop's highlights, so you know what you're signing up for:

* CometBFT `v0.38.x` → `v0.39.x` (LibP2P, `AdaptiveSync`); from v0.53 you jump straight to the `v0.40.0` release v0.55 pins.
* Consolidation of `cosmossdk.io/x/*` vanity modules into `github.com/cosmos/cosmos-sdk/x/*`, plus the Log v2 and Store v2 moves.
* `x/gov` keeper-initialization and `GovHooks` interface changes, `x/epochs` and `x/bank` wiring updates, and the `x/circuit` / `x/nft` / `x/crisis` deprecations.
* IBC v11 (if your chain uses IBC).

Where the two hops interact, land directly on the v0.55 state instead of transiting through v0.54's:

* **Skip transient wiring.** Don't adopt v0.54 reference-app wiring that v0.55 removes in the same hop: the SIGN_MODE_TEXTUAL tx-config setup, `x/protocolpool` (if your v0.53 app didn't already wire it), and `distrkeeper.WithExternalCommunityPool`. Go straight to the v0.55 forms shown in this guide.
* **Module constructors.** v0.54's constructor signatures still carried the legacy `exported.Subspace` arguments; use the v0.55 signatures from [Removed: x/params](#removed-xparams) directly.
* **Custom mempools.** Implement the v0.55 `Mempool` interface ([Mempool Interface Changes](#mempool-interface-changes)) directly; don't bother with the v0.54 shape.
* **Module migrations are cumulative.** `RunMigrations` walks each module from its v0.53 consensus version to the v0.55 target in one pass (`x/auth` 5 → 6 → 7, `x/staking` 5 → 6). No manual intervention is needed beyond the standard upgrade handler.
* **Store upgrades.** The v0.53 → v0.54 hop required no store additions or deletions, so the combined store upgrade is exactly the snippet in [Upgrade Handler and Store Migrations](#upgrade-handler-and-store-migrations): delete `protocolpool` only if your v0.53 app had wired it, and `params` if its store was still mounted (more likely on a v0.53-era app). Use a single upgrade name, e.g. `v053-to-v055`.

Test the full jump on a mainnet-state export before scheduling it: the two-version migration path gets far less ecosystem mileage than the single-version one.

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

Indexers, exchanges, and monitoring that key validators by consensus address must handle the mapping changing over a validator's lifetime. On-chain, `keeper.ValidatorByHistoricalConsAddr` resolves a validator from a rotated-away consensus address.

Chains built on the enterprise `x/poa` module have their own `MsgRotateConsPubKey` with different semantics — no fee, no rate limit, an admin override, and a same-block swap with no rotation history. Because the old consensus address is gone immediately, modules that attribute `LastCommit` signatures or vote extensions by consensus address need extra care across the swap; see the PoA guide below for the caveats and the operator runbook.

For an overview of key rotation and the operator procedures, see [Key rotation](https://docs.cosmos.network/sdk/latest/keys/key-rotation), [Rotate a consensus key, Staking](https://docs.cosmos.network/sdk/latest/keys/rotate-validator-key), and [Rotate a consensus key, PoA](https://docs.cosmos.network/sdk/latest/keys/rotate-validator-key-poa).

### ML-DSA-65 Validator Consensus Keys

Cosmos SDK v0.55 registers the NIST ML-DSA-65 (FIPS 204) post-quantum signature scheme as a supported validator consensus key type ([#26436](https://github.com/cosmos/cosmos-sdk/pull/26436)). The new `cosmos.crypto.mldsa65.PubKey` / `PrivKey` proto messages, Amino routes (`cometbft/PubKeyMlDsa65`, `cometbft/PrivKeyMlDsa65`), interface-registry registration, multisig amino route, and `hd.MlDsa65Type` constant are all enabled by default.

**Action required:** none. Existing chains continue to accept only the consensus key types listed in `genesis.consensus_params.validator.pub_key_types` (still `["ed25519"]` by default). No state-machine-relevant behavior changes for chains that do not opt in.

**To opt in (new chains):** set `genesis.consensus_params.validator.pub_key_types` to `["ml_dsa_65"]` (or a list including it). Validators must then submit `MsgCreateValidator` with a `mldsa65.PubKey`. The `init` and `testnet` commands accept `--consensus-key-algo ml_dsa_65` to generate matching validator files ([#26604](https://github.com/cosmos/cosmos-sdk/pull/26604)). Test harnesses can use the new `testutil/network.Config.ValidatorConsensusKeyType` field together with `genutil.InitializeNodeValidatorFilesFromMnemonicWithKeyType` to spin up an in-process testnet pinned to ML-DSA-65.

**Operational considerations:** ML-DSA-65 keys and signatures are substantially larger than ed25519 (pubkey 1952 bytes vs 32, signature 3309 bytes vs 64). Chains enabling this key type should review `consensus_params.block.max_bytes` and gossip framing limits accordingly. The cometbft commit lift in this release expanded `MaxSignatureSize` and the per-validator `MaxCommitSigBytes` to accommodate the larger signatures; downstream applications relying on the previous fixed values may need to be re-examined.

**Warning — IBC counterparties must upgrade first.** IBC light clients on counterparty chains verify your validator set's commit signatures using the counterparty's own compiled-in crypto. A counterparty running a stack that predates ML-DSA-65 support cannot verify signatures from the new key type: once validators holding sufficient voting power sign with it, your headers fail verification there, IBC packet flow with that chain stops, and the client eventually expires. Before enabling a new consensus key type on a chain with live IBC connections, coordinate so every counterparty chain is running a CometBFT/SDK stack that can verify it — the counterparty only needs the verification code on its nodes, not the key type in its own `pub_key_types`.

Existing chains can combine this with [key rotation](#validator-consensus-key-rotation) to move validators to post-quantum keys: add `ml_dsa_65` to `pub_key_types` via a consensus-params update, then have validators rotate.

For the concepts and operator guides, see [Post-quantum keys](https://docs.cosmos.network/sdk/latest/keys/post-quantum-keys), [Enable ML-DSA keys](https://docs.cosmos.network/sdk/latest/keys/enable-ml-dsa-keys), and [Migrate a validator to ML-DSA](https://docs.cosmos.network/sdk/latest/keys/migrate-validator-ml-dsa).

### ML-DSA-65 Account Keys

[#26472](https://github.com/cosmos/cosmos-sdk/pull/26472) extends ML-DSA-65 support to user account keys: keyring creation and mnemonic recovery (`--algo ml_dsa_65`), transaction signing and verification, and a new ante-handler gas cost param `SigVerifyCostMlDsa65` (added to `x/auth` params by the automatic 6 → 7 migration). No action is required; accounts using existing key types are unaffected.

See [Create an ML-DSA account](https://docs.cosmos.network/sdk/latest/keys/create-ml-dsa-account) and [Post-quantum keys](https://docs.cosmos.network/sdk/latest/keys/post-quantum-keys).

### secp256k1eth Validator Consensus Keys

[#26615](https://github.com/cosmos/cosmos-sdk/pull/26615) adds `crypto/keys/secp256k1eth`, wrapping CometBFT's Ethereum-style secp256k1 consensus key implementation with SDK codec registration. Intended for EVM-compatible chains that want validator consensus addresses derived the Ethereum way; opt in via `genesis.consensus_params.validator.pub_key_types`.

The IBC counterparty warning from the [ML-DSA-65 section](#ml-dsa-65-validator-consensus-keys) applies here too: counterparty chains must run a stack that can verify secp256k1eth signatures before your validators adopt the key type, or IBC connections with them will break.

See [Post-quantum keys](https://docs.cosmos.network/sdk/latest/keys/post-quantum-keys) for how the consensus key types compare.

### Block-STM Configuration

Block-STM parallel execution itself is not new — the engine (`baseapp/txnrunner`) and the `SetBlockSTMTxRunner` hook shipped in v0.54.x, wired programmatically per chain. v0.55 adds standard operator-facing configuration ([#26208](https://github.com/cosmos/cosmos-sdk/pull/26208)): `block-executor` (`"sequential"`, the default, or `"block-stm"`), `block-stm-workers`, and `block-stm-pre-estimate` in `app.toml`, plus a `baseapp/blockexec` helper that resolves them and installs the runner. Chains that already call `SetBlockSTMTxRunner` directly can keep that wiring or switch to the helper.

To adopt the config-driven wiring, call `blockexec.Apply` after creating your store keys (see `simapp/app.go`):

```go
stores := make([]storetypes.StoreKey, 0, len(keys))
for _, k := range keys {
    stores = append(stores, k)
}
blockexec.Apply(bApp, appOpts, stores, txConfig.TxDecoder(),
    func(storetypes.MultiStore) string { return sdk.DefaultBondDenom },
)
```

`Apply` resolves the executor from `app.toml`/flags and installs the corresponding `TxRunner`; with the default `sequential` executor it preserves today's behavior, so the wiring is safe to add unconditionally. Block-STM is incompatible with the block gas meter (disabled by default since v0.54): `Apply` disables the meter automatically when `block-stm` is selected, but chains wiring `SetBlockSTMTxRunner` directly must call `SetDisableBlockGasMeter(true)` first or the runner installation panics.

Switching a running chain's executor is a per-node setting with identical state-transition results, but treat the first enablement as an operational rollout: test with your workload before flipping validators.

## Behavior Changes Affecting Dapps and Indexers

Observable changes between v0.54.x and v0.55.x that don't require code changes but may affect downstream consumers:

* **Block selection uses ante-reported gas.** Proposals are packed using the gas wanted returned by the ante handler at `CheckTx` time rather than the tx-declared gas limit ([#25338](https://github.com/cosmos/cosmos-sdk/pull/25338)). Block composition can differ for txs whose ante-reported gas diverges from their declared limit.
* **Staking emits key-rotation events** (`rotate_cons_pubkey`, `apply_cons_pubkey_rotation`), and validator consensus addresses can change over time ([#26619](https://github.com/cosmos/cosmos-sdk/pull/26619)).
* **`x/gov` `proposal_messages` event attribute** no longer has a leading comma ([#26353](https://github.com/cosmos/cosmos-sdk/pull/26353)).
* **`x/authz` prunes at most 200 expired grants per begin block** ([#26588](https://github.com/cosmos/cosmos-sdk/pull/26588)); mass-expiry cleanup now spreads across blocks.
* **`x/distribution` reward withdrawals to blocked addresses** during begin/end block fall back to the delegator/validator owner and then the community pool instead of failing ([#26406](https://github.com/cosmos/cosmos-sdk/pull/26406)). User-initiated withdrawals to blocked addresses still return `ErrUnauthorized`.
* **`x/feegrant` `Allowances` and `AllowancesByGranter` queries** now honor `PageRequest.offset` and `count_total` correctly ([#26596](https://github.com/cosmos/cosmos-sdk/pull/26596)); clients that compensated for the old off-by-page results should re-check.
