# ADR 077: x/safety, an Emergency Msg Breaker

## Changelog

* 2026-10-21: Initial draft
* 2026-10-22: Msgs run only through the router; split into the x/safety PR and the msg server PR
* 2026-10-23: Correct the downstream migration and compatibility claims; leave the removal release open; shorten
* 2026-10-23: Correct the refusal error, the keeper Msg methods, the circuit store and the migration of circuit trips

## Status

DRAFT Not Implemented

## Abstract

`x/safety` replaces `contrib/x/circuit`. One operator address stops a msg type in a single tx during an incident, and governance decides who the operator is and which msgs it may stop. The operator cannot block the gov, upgrade, safety or consensus parameter msgs, so governance can always reset a block. BaseApp checks the breaker in the `MsgServiceRouter` and in the ante handler, and code outside a module runs a msg with `msgservice.Dispatch`, which goes through the router. In a second PR, the SDK deprecates the APIs that call a msg server directly.

## Context

`contrib/x/circuit` breaks down during an incident:

* The `MsgTripCircuitBreaker` handler accepts any string, and returns success for a typo, a space-joined list of URLs or an empty list while disabling nothing.
* An authorized account can trip `MsgResetCircuitBreaker`, after which gov proposals that reset fail at execution.
* Operators configure a levels table with super-admin delegation, and the keeper compares a build-time authority with `bytes.Equal`, so only a software upgrade changes the authority.
* The module emits untyped, comma-joined event attributes, registers no errors and serves the disabled list without pagination, and its README documents CLI commands that do not exist.
* The ante decorator checks only top-level msgs, so a tripped msg inside authz `MsgExec` or a gov or group proposal reaches the mempool.
* cosmos/evm precompiles call SDK msg servers directly, past the router's breaker check, so after an operator trips `MsgSend`, an ERC20 transfer of the same coins still goes through.
* SimApp stopped wiring the module in v0.54.

## Alternatives

* **Patch `x/circuit` in place.** Validation and events need small fixes, but removing the levels table, delegation and build-time authority breaks existing integrations anyway.
* **Let gov trip.** Governance cannot pass even an expedited vote in time to stop an exploit, and an attacker who controls a proposal could block recovery msgs.
* **Expire trips.** An incident can outlast any preset timer, and an automatic unblock mid-exploit does more harm than an operator resetting late.
* **Exempt wrapper msgs from trips.** Recovery through authz and group stays open, but the operator loses any way to stop authz or group execution as a whole.
* **Check the breaker inside every msg server**, by hand or with a protoc plugin. A module that misses the check fails open, and the authz, group, consensus, crisis and nft keepers, which serve their own Msg services, still need hand-written checks.
* **Add `Dispatch` without unexporting msg servers.** Code that calls a msg server directly still compiles and skips the breaker, and the chain finds out during the incident.

## Decision

### Roles

* **authority**: `sdk.ValidateAuthority` resolves it to the consensus `AuthorityParams` override when set, and otherwise to `params.authority`, which genesis sets to the address the app passes to `NewKeeper` (the gov module account in SimApp) and the current authority hands on with `MsgUpdateParams`, as in `enterprise/poa`. The authority sets the operator and scope and resets any pattern.
* **operator**: one address in params, or none. It trips and resets patterns within its scope, effective in the same block. The authority does not trip unless it names itself the operator, which the x/safety README documents as the way to let governance trip directly.
* **scope**: a list of patterns, empty at genesis, so a new operator trips nothing until the authority grants scope.

### Patterns

A pattern is `*`, a prefix such as `/cosmos.bank.v1beta1.*` that matches on proto package segment boundaries, or an exact msg type URL. The `MsgTrip` handler rejects a pattern that matches no registered, unprotected msg type.

### Protected set

`IsAllowed` always allows msgs under `cosmos.safety.`, `cosmos.gov.` and `cosmos.upgrade.`, and `/cosmos.consensus.v1.MsgUpdateParams`.

The SDK hardcodes this set and the wrapper set below, so no chain configuration can leave out the reset path. To change either list, a chain needs an SDK release or a fork; the breaker does not protect a governance module other than `cosmos.gov`; and the SDK must add each new authz or group proto version to the wrapper list.

### Wrappers

Authz `MsgExec` and group `MsgSubmitProposal`, `MsgVote` and `MsgExec` run other msgs. The breaker blocks a wrapper only on an exact trip of its own URL. Under a `*` or prefix trip it lets the wrapper through and still blocks the tripped msgs inside it.

An authority or validator acting through an authz grant or group policy can still reach `MsgReset` and gov votes during a global halt, and the operator stops authz or group execution outright with an exact trip when a grantee is compromised. Under `*`, the group module still records proposals and votes from `MsgSubmitProposal` and `MsgVote`, and with `EXEC_TRY` runs the proposal's msgs, which the router still checks.

### Scope changes and expiry

When the authority narrows the scope, existing trips stay in place; gov resets them in a separate proposal, since a reset that names an entry the operator already removed fails the whole proposal. Trips never expire.

### Enforcement

* **Router.** The keeper implements `baseapp.CircuitBreaker`, and the app wires it with `BaseApp.SetCircuitBreaker`. The router checks every msg it routes, in every exec mode, including block hooks, vote extensions and query contexts. An x/safety refusal returns `ErrMsgDisabled` from the safety codespace; a breaker that refuses with no error of its own gets `sdkerrors.ErrCircuitBreakerTripped`. A hybrid handler checks the msg type it was registered for, not the type of the request it receives.
* **Ante.** `HandlerOptions.CircuitBreaker` adds `ante.CircuitBreakerDecorator`, which checks every msg in a tx and recurses into any msg with `GetMessages` or `GetMsgs`, as deep as the tx decoder unpacks. It runs in `CheckTx`, `ReCheckTx`, `Simulate`, and in `PrepareProposal` and `ProcessProposal` when the app runs the ante chain there. It passes txs through in `FinalizeBlock`, so a tx included after a trip spends its sequence and fails at the router, and cannot execute after the reset.
* **Dispatch.** Code outside the module that owns a msg, such as a cosmos/evm precompile, runs it with `msgservice.Dispatch[R](ctx, router, msg)`. `Dispatch` calls the router handler on a branch of `ctx`, so `ValidateBasic`, the breaker and the msg server's events apply, writes the branch only on success, and returns the typed response. The caller still authorizes the signers.
* **Keepers.** Logic that other modules need, such as a bank send or authz `SaveGrant`, stays an exported keeper method that the breaker does not see. Authz `DispatchActions` runs each msg through the router, so a trip still blocks it. Upgrade handlers call keepers, not `Dispatch`, so a trip cannot halt an upgrade; x/safety and x/consensus export `SetParams` for that. The circuit migration's `MsgTrip` is the one exception: it is protected, so no trip blocks it.
* **Gas.** The keeper reads the disabled list on an infinite gas meter, so the check charges no gas, and a msg uses the same gas as on a chain with no circuit breaker. Per msg the keeper makes at most two point lookups plus one per package segment.
* A breaker that returns an error blocks the msg.

### Unexported msg servers

Each SDK module unexports its msg server type and exports `RegisterMsgServer`, which `RegisterServices` calls. The authz, group, consensus, crisis, nft and counter keepers move their Msg handlers to an unexported type and keep deprecated Msg methods on the keeper, so they still satisfy their `MsgServer` interface until those methods go. Once they go, code runs an SDK `sdk.Msg` only through the router, short of a caller that passes `RegisterMsgServer` a fake registrar to capture the msg server.

Deprecated wrappers, kept until a future release removes them:

* `NewMsgServerImpl` in each module, and `NewLegacyMsgServerImpl` in `x/gov`
* the Msg methods on the authz, consensus, crisis, nft, group and counter keepers
* `enterprise/poa`'s `NewMsgServer` and `MsgServer`
* `x/staking/testutil.NewHelper`, replaced by `NewHelperWithRouter`
* `MsgServiceRouter.HybridHandlerByMsgName`, which skips `ValidateBasic`

A guard test fails when an SDK package exports any other msg server type, constructor or Msg handler.

### Wiring

x/safety uses gogoproto types only: no `cosmossdk.io/api` types, no autocli options and no depinject config. Apps wire it by hand, and operators use the `tx safety` and `q safety` commands that autocli derives from the module's services.

### Migrating from x/circuit

In one upgrade, the chain stops mounting the circuit store, whose data stays in the database outside the app hash, adds the safety store, and swaps the circuit keeper for the safety keeper in `SetCircuitBreaker` and the ante handler. x/safety does not read the circuit disabled list or permissions, so the upgrade handler sets the operator and scope, then dispatches a `MsgTrip` from the operator for the msg type URLs circuit had tripped, and the upgrade block refuses them. `MsgTrip` refuses a URL that matches no registered msg, so a typo fails the upgrade instead of disabling nothing.

## Consequences

### Backwards Compatibility

In the x/safety PR, the SDK adds `x/safety`, `msgservice.Dispatch` and `MsgRouter`, `ante.HandlerOptions.CircuitBreaker` and `NewCircuitBreakerDecorator`, and `sdkerrors.ErrCircuitBreakerTripped`. Three results change:

* When a breaker refuses without an error of its own, as `contrib/x/circuit` does, the tx now fails with codespace `sdk`, code 43, where it failed with codespace `undefined`, code 1. The chain hashes tx results into the results hash, so a backport to v0.54.x or v0.55.x breaks the state machine.
* A hybrid handler refuses a request when its own msg type is tripped, whatever the request's type.
* `CollectionPaginate` and `CollectionFilteredPaginate` report the real total when `count_total` is set and the offset runs past the last item.

The SDK also deprecates `contrib/x/circuit`, so staticcheck reports SA1019 on its import.

In the msg server PR, the SDK adds `RegisterMsgServer` to each keeper package, `NewHelperWithRouter` and consensus `Keeper.SetParams`, and deprecates the APIs listed under [Unexported msg servers](#unexported-msg-servers). Nothing breaks at compile time; staticcheck reports SA1019 at each call to a deprecated function until the caller migrates. It does not flag a keeper registered as its own `MsgServer`, since that path calls no deprecated function by name.

### Positive

* Module authors write no breaker code.
* The `MsgTrip` handler rejects typos, empty lists and protected msgs.
* The operator cannot block the reset or governance with any sequence of trips.
* Gov controls the operator and scope through ordinary proposals and can hand off the authority.
* Once maintainers remove the deprecated APIs, code that calls another module's msg server directly no longer compiles, short of capturing it through a fake registrar.

### Negative

* Downstream code that calls SDK msg servers must move to `Dispatch` or a keeper call before the removal. On `main` today, cosmos/evm's staking, distribution, gov, slashing and ERC20 precompiles construct SDK msg servers in production code, and seven of its test helpers do the same. wasmd calls them only in four tests. ibc-go calls none.
* A `Dispatch` caller needs the app's `MsgServiceRouter`, and a multistore whose `CacheMultiStore` returns itself instead of a branch must be fixed first. The cosmos/evm snapshot multistore is one; without the fix a reverted EVM call keeps the msg's writes.
* Modules outside the SDK that export their msg servers, and keeper methods that run a msg's logic, remain paths the breaker does not see.
* A `*` trip blocks `MsgEthereumTx`, so validators on an EVM chain vote with a Cosmos `MsgVote` during a global halt.
* Recheck evicts txs already in the mempool at trip time; senders rebroadcast after the reset.
* Without autocli from `runtimeservices.ExtractAutoCLIOptions`, operators sign `MsgTrip` and `MsgReset` as tx JSON.
* An override authority that can act only through a wrapper the operator tripped by exact URL cannot recover on chain.
* Only the authority removes the operator. Until a gov vote passes, a compromised operator can trip anything in its scope and re-trip after each reset.
* `MsgReset` needs every named entry present, so a gov reset fails if the operator changes the list during the vote.
* While a trip covers a msg, a gov proposal containing it fails `CheckTx`, so the operator narrows the trip before gov can propose a fix that is not a software upgrade.

### Neutral

* The breaker treats IBC relayer msgs, the ICA host, Ethereum txs and wasm contract msgs differently. The x/safety README lists the behavior and runbook for each.

## Further Discussions

* Switching cosmos/evm precompiles and packet-forward-middleware to `Dispatch`.
* A bank `SendRestriction` backed by the breaker, to cover keeper-level transfers.
* A backport to v0.54.x and v0.55.x for chains on those lines, including cosmos/evm.
* The release that removes the deprecated msg server APIs.

## Test Cases

* x/safety keeper, pattern and ante decorator tests, with fuzzing of pattern parsing and of the keeper against a model.
* `Dispatch` tests in `types/msgservice`, and `baseapp` tests that the router checks `Dispatch` in every exec mode and that hybrid handlers check their own msg type.
* SimApp integration tests: gas of a bank send, a delegation, an authz exec and a gov proposal matches SimApp with no breaker, with 0 and 1000 disabled entries; `Dispatch` from block hooks, `FinalizeBlock` and queries; an upgrade off `contrib/x/circuit`; a `runtime.App` wired by hand; and a check that no upgrade handler calls `Dispatch`.
* System tests `TestSafety*`, and `TestChainUpgrade`, which upgrades a running v0.54 chain into x/safety.
* Local test branches of cosmos/evm, ibc-go and wasmd, built against both PRs, cover precompiles switched to `Dispatch`, `MsgEthereumTx` and `*` trips, relayed IBC transfers, ICA and packet-forward-middleware, and wasm contract msgs under trips. Removing `SetCircuitBreaker` fails the tests that depend on it.

## References

* [#26906](https://github.com/cosmos/cosmos-sdk/pull/26906): x/safety, `msgservice.Dispatch` and the router fixes
* [#26910](https://github.com/cosmos/cosmos-sdk/pull/26910): unexported msg servers and the deprecated wrappers
* [`contrib/x/circuit`](../../contrib/x/circuit)
