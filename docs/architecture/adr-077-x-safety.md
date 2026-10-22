# ADR 077: x/safety, an Emergency Msg Breaker

## Changelog

* 2026-10-21: Initial draft
* 2026-10-21: Direct msg server callers use `msgservice.Dispatch`
* 2026-10-22: Msgs run only through the router; SDK modules stop exporting their msg servers
* 2026-10-22: gogoproto types only; no autocli options or depinject; generic service commands for operators
* 2026-10-22: Split into the x/safety PR and the msg server PR; direct msg server calls get a one-release deprecation window
* 2026-10-22: Correct the authority, wrapper and ante depth descriptions; list the result code change and the full deprecation set

## Status

DRAFT Not Implemented

## Abstract

`x/safety` replaces `contrib/x/circuit`. It lets one operator address stop a msg type in a single tx during an incident, and lets governance decide who that operator is and what it may stop. The authority, gov by default, sets the operator and its scope and can reset any block, but cannot block anything itself. Gov, upgrade, safety and consensus parameter msgs stay outside the operator's reach, so the operator can never lock governance out of the reset. BaseApp checks the breaker at the router and in the ante handler. The router becomes the only way to run an `sdk.Msg`: SDK modules stop exporting their msg servers, and code outside a module, such as a cosmos/evm precompile, runs a msg with `msgservice.Dispatch`, which the router checks as it checks a tx.

The change ships in two PRs. The x/safety PR adds the module, `Dispatch` and the router fixes. It breaks no API, but a refusal by a breaker with no error of its own now returns a registered error code. The msg server PR unexports the msg servers, keeps the old constructors, keeper Msg methods and `HybridHandlerByMsgName` as deprecated wrappers for one release, and removes them in the next. Everything below assumes both have merged.

## Context

`x/circuit` lets chains disable msg types, but fails in the situations it exists for:

* `MsgTripCircuitBreaker` accepts any string. A typo, a space-joined list of URLs from the CLI, or an empty list returns success, emits a trip event, and disables nothing.
* An authorized account can trip `MsgResetCircuitBreaker` itself, and once it does, gov proposals that try to reset fail at execution.
* Permissions use a levels table with super-admin delegation. The module authority lives in the binary and the keeper compares it with `bytes.Equal`, so only a software upgrade can change it.
* Events use untyped, comma-joined attributes, the module returns unregistered errors, the disabled list query walks the whole set without pagination, and the README documents CLI commands that do not exist.
* The ante decorator checks only top-level msgs, so a tripped msg inside authz `MsgExec`, a group proposal or a gov proposal reaches the mempool.
* BaseApp checks the breaker only in the `MsgServiceRouter`. cosmos/evm precompiles, ibc-go and wasmd call msg servers or keepers directly and skip it, so a selective trip of `MsgSend` does not stop an ERC20 transfer of the same coins.
* SimApp has not wired the module since v0.54, so no reference app exercises it.

Chains need a key that stops a msg type within one block of someone finding an exploit, and governance needs to control who uses that key and how far it reaches.

## Alternatives

* **Patch `x/circuit` in place.** The validation and event fixes are small, but operators misconfigure the levels table, super-admin delegation and the build-time authority, and removing them breaks existing integrations either way.
* **Let gov trip.** A gov trip waits at least an expedited voting period, too slow to matter during an exploit, and lets an attacker who controls a proposal block recovery msgs.
* **Expire trips after a fixed time.** An incident outlasts any preset timer, and an automatic unblock mid-exploit does more harm than a manual one.
* **Protect wrapper msgs outright.** Exempting authz `MsgExec` and group execution from trips keeps recovery paths through them open, but removes the only way to stop authz or group execution as a whole.
* **Make every module depend on an x/safety keeper.** Every module would import x/safety, and every module author would still have to call it in every method.
* **Call the guard on the first line of every msg server method.** BaseApp would set the breaker on `sdk.Context`, and each method would call `msgservice.CheckMsgAllowed`. A method that forgets the line fails open with no error, and every module author pays for a check that only direct callers need.
* **Generate a guarded wrapper for every Msg service and check it at startup.** A protoc plugin removes the per-method line, but each module still wraps its msg server, modules outside the SDK must add the plugin, and the authz, group, consensus, crisis and nft keepers, which are their own msg servers, still need guards written by hand.
* **Dispatch by convention without unexporting msg servers.** Add `Dispatch` and tell outside callers to use it, but keep `NewMsgServerImpl` exported. A direct call still compiles and silently skips the breaker, and the chain finds out during the incident.
* **Wrap msg servers with reflection at registration.** Go cannot add methods to a type at run time, so the router cannot hand a direct caller a checked msg server that satisfies the module's `MsgServer` interface.

## Decision

### Roles

* **authority**: checked with `sdk.ValidateAuthority`. Without the consensus `AuthorityParams` override, it is the address in params, the gov module account at genesis, and as in `enterprise/poa` the current authority sets a new one with `MsgUpdateParams`. With the override set, the override account is the authority and the address in params has no effect. The authority sets the operator and scope and resets any pattern. It cannot trip.
* **operator**: one address stored in params, or empty. It trips and resets patterns within its scope, effective in the same block.
* **scope**: a list of patterns, empty at genesis, so a new operator can trip nothing until the authority grants scope.

### Patterns

A pattern is `*` for every msg, a prefix such as `/cosmos.bank.v1beta1.*` that matches on proto package segment boundaries, or an exact msg type URL. The `MsgTrip` handler rejects a pattern that matches no registered, unprotected msg type, so typos and no-op trips fail.

### Protected set

`IsAllowed` always allows msgs under `cosmos.safety.`, `cosmos.gov.` and `cosmos.upgrade.`, plus `/cosmos.consensus.v1.MsgUpdateParams`, whatever the disabled list contains.

The SDK hardcodes this set and the wrapper set below. A list in params or genesis could leave out the reset path: a bad genesis file, or a params update that leaves out `MsgReset` or gov, would let a trip block its own reset, and the only way back would be a halt and a new binary. A configured wrapper list could also exempt an ordinary msg from `*`. A hardcoded list cannot be misconfigured on chain, but changing either list takes an SDK release or a fork. The protected set covers governance only under `cosmos.gov.`, so a chain that governs through another module gets no protection for it. The wrapper list names versioned URLs, so a new authz or group proto version must be added to it.

### Wrappers

Authz `MsgExec` and group `MsgSubmitProposal`, `MsgVote` and `MsgExec` run other msgs. The breaker blocks a wrapper only on an exact trip of its own URL. Under a `*` or prefix trip it lets the wrapper through and still blocks the tripped msgs it dispatches.

An authority or validator that acts through an authz grant or a group policy needs the wrapper during a global halt, or `*` locks it out of `MsgReset` and gov votes. The operator also needs to stop authz or group execution as a whole, such as when a grantee is compromised. The exact-trip rule covers both: `*` and prefixes leave recovery through wrappers open, and an exact trip of each wrapper stops it. Letting a wrapper through does open something: under `*`, group `MsgSubmitProposal` and `MsgVote` still record proposals and votes, and with `EXEC_TRY` run the proposal without a group `MsgVote` or `MsgExec` reaching the router. That is the price of keeping recovery through group policies open.

### Scope changes and expiry

When gov narrows the scope, existing trips stay in place, and gov resets them in the same proposal. Trips never expire.

### Enforcement

* **Router.** The keeper implements `baseapp.CircuitBreaker`, and the app wires it with `BaseApp.SetCircuitBreaker`. The router checks every msg it routes, on every call and in every exec mode, including `CheckTx`, `Simulate`, `FinalizeBlock`, block hooks, vote extensions and query contexts. It reads the breaker on each call, so a breaker set after `RegisterServices` still covers every handler.
* **Ante.** `ante.CircuitBreakerDecorator` in `x/auth/ante`, which `HandlerOptions.CircuitBreaker` adds, checks every msg in a tx and recurses into any msg that exposes `GetMessages` or `GetMsgs`, as deep as the tx decoder unpacks `Any` values, so it adds no nesting rule of its own. It runs in `CheckTx`, `ReCheckTx` and `Simulate`, and in `PrepareProposal` and `ProcessProposal` only with an app-side mempool; with the default `NoOpMempool` neither proposal handler runs the ante chain. It passes txs through in `FinalizeBlock`. A tx included after a trip then spends its sequence and fails at the router, so it cannot execute again once the operator resets. A gov or group proposal containing a tripped msg fails in `CheckTx`; one that reaches `FinalizeBlock` is created and fails at execution while the trip stands.
* **One way to run a msg.** An `sdk.Msg` runs only through the `MsgServiceRouter`. In every SDK module the msg server type is unexported, and the keeper package exports `RegisterMsgServer(r grpc.ServiceRegistrar, ...)`, which `RegisterServices` calls with the keepers the msg server needs: the module keeper in most modules, the account keeper in `x/auth`, and the account and bank keepers in `x/auth/vesting`, whose function lives in package `vesting`. The authz, group, consensus, crisis, nft and counter keepers stop implementing their Msg service; their Msg methods move to an unexported type. For one release, `NewMsgServerImpl`, the keeper Msg methods, `enterprise/poa`'s `NewMsgServer` and `MsgServer`, `x/staking/testutil.NewHelper` and `MsgServiceRouter.HybridHandlerByMsgName`, which runs a handler by name without `ValidateBasic`, stay as deprecated wrappers; the next release removes them. A guard test fails on any other exported msg server type, constructor or Msg handler. A test in `keeper_test` reaches the constructor through `export_test.go`.
* **Dispatch.** Code outside the module that owns a msg, such as a cosmos/evm precompile or another module's keeper, runs it with `msgservice.Dispatch[R](ctx, router, msg)`. `Dispatch` calls the router handler on a branch of `ctx`, so `ValidateBasic`, the breaker check and the msg server's events apply, writes the branch only on success, and returns the typed response. It does not check signers; the caller authorizes them as before.
* **Keepers for module logic.** Logic another module needs, such as authz `DispatchActions` or a bank send, stays an exported keeper method, and the breaker does not see it. Upgrade handlers call keepers, not `Dispatch`, so a trip cannot halt an upgrade; the consensus keeper gains `SetParams` for handlers that ran consensus `MsgUpdateParams`, and a SimApp test fails if its upgrade handler calls `Dispatch`. A module's calls to its own handlers, such as group running `Exec` after a vote, stay internal calls on the unexported type.
* A caller that registers a fake `grpc.ServiceRegistrar` can still capture the msg server it receives. That is deliberate abuse, not an API.
* A breaker that returns an error blocks the msg.
* **Gas.** The breaker check charges no gas. The keeper runs its disabled-list lookups on an infinite gas meter in place of the context's meter, so the router, the ante decorator and `Dispatch` charge a tx the same gas as with no breaker set, and `Simulate` stays as close to `FinalizeBlock` gas as without one. Charging would add 5000 gas to a bank send at the router, and, because the decorator checks in `Simulate` but not in `FinalizeBlock`, another 5000 to its gas estimate that users pay fees for and never use. The work stays bounded without gas: per msg, at most two point lookups plus one per package segment, on a type URL the interface registry knows or, through the `IsAllowed` query, one of at most 256 bytes.

### Wiring and client

* **Manual wiring, gogoproto types only.** x/safety has no pulsar types in `cosmossdk.io/api`, no autocli options and no depinject config. Apps wire it by hand. The module defines no CLI code; operators trip, reset and query with the generic `tx safety` and `q safety` commands that autocli derives from the module's Msg and Query services.

### Migrating from x/circuit

Chains remove `x/circuit` and add `x/safety` in the same upgrade: delete the circuit store, add the safety store, and swap the circuit keeper for the safety keeper in `SetCircuitBreaker` and the ante handler. The circuit disabled list and permissions do not migrate. After the upgrade, gov sets the operator and scope, and the operator re-trips anything that must stay blocked.

## Consequences

### Backwards Compatibility

* Existing chains need a store upgrade that adds the safety store key, and `x/circuit` chains also delete the circuit store.
* The x/safety PR adds API only: the module, `msgservice.Dispatch` and `ante.HandlerOptions.CircuitBreaker`. It changes one result: when a breaker refuses a msg without an error of its own, as `contrib/x/circuit` does, the router returns `sdkerrors.ErrCircuitBreakerTripped` (codespace `sdk`, code 43) in place of an unregistered error that reached clients as codespace `undefined`, code 1. A circuit chain that only bumps the SDK sees the new code in tx results, which feed the results hash, so a v0.54.x backport is state machine breaking. Importing the deprecated `contrib/x/circuit` package also trips staticcheck SA1019.
* The msg server PR deprecates `NewMsgServerImpl`, the keeper Msg methods, `enterprise/poa`'s `NewMsgServer` and `MsgServer`, `x/staking/testutil.NewHelper` and `MsgServiceRouter.HybridHandlerByMsgName`, adds `RegisterMsgServer` to each keeper package, and removes the deprecated calls in the next release. Nothing breaks at compile time, but downstream staticcheck reports SA1019 at each call, in app code and in tests, until the chain migrates.
* A caller that switches to `Dispatch` needs the app's `MsgServiceRouter`, so its constructor gains a router argument, and a multistore whose `CacheMultiStore` returns itself instead of a branch must be fixed first. The cosmos/evm snapshot multistore is one; without the fix a reverted EVM call keeps the msg's writes.

### Positive

* Under a selective trip, precompiles that run their SDK msgs with `Dispatch` revert. In evmd with the precompiles switched to `Dispatch`, staking `delegate`, ERC20 and WERC20 `transfer`, distribution `withdrawDelegatorRewards`, slashing `unjail` and ICS20 `transfer` revert under a trip, and `eth_call` and `eth_estimateGas` return an error. Reverting the precompiles to direct msg server calls fails every one of those tests.
* Module authors write no breaker code, and a new msg server needs nothing beyond `RegisterMsgServer`.
* Once the deprecated calls are removed, a direct call to another module's msg server no longer compiles, so no in-repo or downstream caller can skip the breaker by accident. Until then, staticcheck flags each downstream call. In the SDK, only the deprecated wrappers and their own tests call them.
* Downstream migration is small. cosmos/evm deleted a deprecated `erc20.NewMsgServerImpl` wrapper around the bank msg server and moved 2 test calls to `Dispatch`; its precompiles already used `Dispatch`. ibc-go needed no change. wasmd moved 5 test calls and no production code.
* Typos, empty lists and protected msgs fail at trip time instead of succeeding silently.
* No sequence of trips can block the reset or governance.
* Gov controls the operator and its scope through ordinary proposals, and can replace itself as authority.

### Negative

* Every chain and module that called an SDK msg server from code must move to `Dispatch` or a keeper call before the release that removes the deprecated calls.
* A module outside the SDK that still exports its msg server, and any keeper method that runs a msg's logic, remain direct paths the breaker does not see.
* Under a `*` trip the breaker blocks `MsgEthereumTx`, so during a global halt validators on an EVM chain vote with a Cosmos `MsgVote` tx, not through the gov precompile.
* Recheck evicts txs already in the mempool at trip time, and their senders rebroadcast after the reset.
* An app that does not build autocli from `runtimeservices.ExtractAutoCLIOptions` has no `tx safety` commands, so its operators sign `MsgTrip` and `MsgReset` as tx JSON. x/safety has no depinject config, so an app adds it with manual wiring, and a `runtime.App` registers it with `RegisterStores` and `RegisterModules` before `Load`.
* An authority set through the consensus `AuthorityParams` override that can act only through a wrapper the operator tripped by exact URL cannot recover on chain.
* Only the authority removes the operator, by gov vote by default. Until the vote passes, a compromised operator can trip anything in its scope and re-trip after each reset.
* `MsgReset` needs every named entry present, so a gov reset fails if the operator changes the list during voting.
* While a trip covers a msg, a gov proposal that contains it fails `CheckTx`, so a fix that is not a software upgrade needs the trip narrowed first.

### Neutral

* IBC relayer msgs, the ICA host, cosmos/evm Ethereum txs and wasmd contract msgs each meet the breaker differently, and some keeper calls skip it; the [x/safety README](../../x/safety/README.md#ibc-evm-and-wasm) lists the behavior and the runbook for each.

## Further Discussions

* Switching cosmos/evm precompiles and packet-forward-middleware to `Dispatch`.
* A bank `SendRestriction` backed by the breaker, which would cover keeper-level transfers.
* A backport to v0.54.x for chains on that line, including cosmos/evm.

## Test Cases

* `x/safety` keeper and pattern tests, `ante.CircuitBreakerDecorator` tests, and fuzzing of pattern parsing and of the keeper against a model.
* `types/msgservice` tests of `Dispatch`: the typed response, a tripped msg, an unknown msg, a pulsar msg, a wrong response type, a nil router, and a failed msg leaving `ctx` unchanged.
* `baseapp` tests that the router checks `Dispatch` in `CheckTx`, `Simulate`, `FinalizeBlock`, block hooks, vote extensions and a query context, and hybrid handler tests for the registered msg type and a breaker set or removed after registration.
* `tests/integration/safety`: delivered and simulated gas of a bank send, a delegation, an authz exec and a gov proposal match SimApp with no breaker set, with an empty disabled list and with 1000 entries.
* `tests/integration/safety`: `Dispatch` from block hooks, a `FinalizeBlock` tx and a query context in SimApp under a trip and after the reset.
* `types/msgservice`: a guard test that fails when any SDK package exports a msg server type, constructor, alias or Msg handler func outside the deprecated wrappers, or when a listed wrapper loses its `Deprecated:` doc.
* `simapp`: a test that fails when an upgrade handler calls `Dispatch`, and `x/consensus` tests of `Keeper.SetParams`.
* `tests/integration/safety`: an upgrade off `contrib/x/circuit` that restarts SimApp without the circuit wiring, and a `runtime.App` with x/safety wired by hand.
* System tests: `TestSafety*` and `TestChainUpgrade`, which upgrades a running v0.54 chain into x/safety.
* cosmos/evm: precompile tests in evmd with the precompiles switched to `Dispatch`, ante tests of `ante.CircuitBreakerDecorator` under `MsgEthereumTx` and `*` trips, and a node-level script on a 4-validator testnet covering precompiles, `MsgEthereumTx` and `*` trips, native transfers and queued txs.
* ibc-go: relayed transfers, recv, ack, timeout and client update trips, client expiry and recovery, misbehavior, ICA host and packet-forward-middleware, all through real `MsgTrip` and `MsgReset` txs.
* wasmd: contract bank, staking, gov, `Any`, Stargate, burn, ICS20 and raw packet msgs, and submessage replies under trips.
* In each external repo, removing `SetCircuitBreaker` or the ante decorator fails the tests that depend on it, and in cosmos/evm so does reverting the precompiles to direct msg server calls.

## References

* [#26906: the x/safety PR](https://github.com/cosmos/cosmos-sdk/pull/26906): the module, `msgservice.Dispatch` and the router fixes, which folds in [#26892](https://github.com/cosmos/cosmos-sdk/pull/26892) and [#26898](https://github.com/cosmos/cosmos-sdk/pull/26898)
* [#26894: the msg server PR](https://github.com/cosmos/cosmos-sdk/pull/26894), unexported msg servers with a one-release deprecation window
* [`contrib/x/circuit`](../../contrib/x/circuit)
* [ADR 068: Preblock](./adr-068-preblock.md)
