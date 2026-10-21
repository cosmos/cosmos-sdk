# ADR 077: x/safety, an Emergency Msg Breaker

## Changelog

* 2026-10-21: Initial draft
* 2026-10-21: Generated msg server guard and startup check replace the per-method guard

## Status

DRAFT Not Implemented

## Abstract

`x/safety` replaces `contrib/x/circuit`. It lets one operator address stop a msg type in a single tx during an incident, and lets governance decide who that operator is and what it may stop. The authority, gov by default, sets the operator and its scope and can reset any block, but cannot block anything itself. Gov, upgrade, safety and consensus parameter msgs stay outside the operator's reach, so the operator can never lock governance out of the reset. BaseApp checks the breaker at the router, in the ante handler, and at the entry of every in-repo msg server, so code that calls a msg server directly, such as a cosmos/evm precompile, stops under the same trip as a tx.

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
* **Make every module depend on an x/safety keeper.** Every module would import x/safety. Setting the breaker on `sdk.Context` covers the same calls with no new dependency.
* **Guard each external repo separately.** cosmos/evm, ibc-go and wasmd could each check the breaker before calling into the SDK, but all of them reach SDK msg servers, so one check at msg server entry covers them all.
* **Call the guard on the first line of every msg server method.** Module authors have to remember it in every method, and a module that forgets it leaves direct calls unchecked with no error. A generated wrapper needs no per-method code, and a startup check stops an app that registers a msg server without it.

## Decision

### Roles

* **authority**: an address stored in params, the gov module account at genesis. As in `enterprise/poa`, the current authority sets a new authority with `MsgUpdateParams`, checked with `sdk.ValidateAuthority`. The authority sets the operator and scope and resets any pattern. It cannot trip.
* **operator**: one address stored in params, or empty. It trips and resets patterns within its scope, effective in the same block.
* **scope**: a list of patterns, empty at genesis, so a new operator can trip nothing until the authority grants scope.

### Patterns

A pattern is `*` for every msg, a prefix such as `/cosmos.bank.v1beta1.*` that matches on proto package segment boundaries, or an exact msg type URL. The `MsgTrip` handler rejects a pattern that matches no registered, unprotected msg type, so typos and no-op trips fail.

### Protected set

`IsAllowed` always allows msgs under `cosmos.safety.`, `cosmos.gov.` and `cosmos.upgrade.`, plus `/cosmos.consensus.v1.MsgUpdateParams`, whatever the disabled list contains. The SDK defines the list in code, not in params or genesis, so neither the operator nor a bad genesis file can lock out the reset.

### Wrappers

Authz `MsgExec` and group `MsgSubmitProposal`, `MsgVote` and `MsgExec` dispatch other msgs through the router. The breaker blocks a wrapper only on an exact trip of its own URL. Under a `*` or prefix trip it lets the wrapper through and still blocks the tripped msgs inside it, so recovery through an authz grant or a group policy keeps working during a global halt, and the operator can still stop authz or group execution as a whole.

### Scope changes and expiry

When gov narrows the scope, existing trips stay in place, and gov resets them in the same proposal. Trips never expire.

### Enforcement

* **Router.** The keeper implements `sdk.CircuitBreaker`, and the module wires it with `BaseApp.SetCircuitBreaker`. The router checks every msg it dispatches in `FinalizeBlock` and `Simulate`.
* **Ante.** `SafetyDecorator` checks every msg in a tx and recurses into any msg that exposes `GetMessages` or `GetMsgs`, up to a depth limit. It runs in `CheckTx`, `ReCheckTx`, `Simulate` and the proposal handlers, and passes txs through in `FinalizeBlock`. A tx included after a trip then spends its sequence and fails at the router, so nobody can replay it once the operator resets. A proposal containing a tripped msg fails at submission.
* **Msg server entry.** BaseApp sets the breaker on the `sdk.Context` of every tx, block hook, proposal handler, vote extension and query. `protoc-gen-go-msgguard`, run by `make proto-gen`, generates `NewGuardedMsgServer` for every Msg service; each of its methods calls `msgservice.CheckMsgAllowed(ctx, msg)` and then the wrapped msg server. Every in-repo `NewMsgServerImpl` returns it, protected modules included. The authz and group keepers, which are their own msg servers and which other modules call directly, call `msgservice.CheckMsgAllowed` in each msg method and implement the guard marker, so a direct call on the keeper, and group's calls to its own methods, are checked once. The router marks the msg it already checked, so the breaker checks a routed msg once, not twice, and checks a direct call, such as from a precompile, at msg server entry. `msgservice.Dispatch` runs a msg through the router for callers, such as precompiles, that want the router's handling.
* **Startup check.** With a breaker set, `MsgServiceRouter.RegisterService` panics on a msg server that does not implement `msgservice.GuardedMsgServer`, and `SetCircuitBreaker` panics when a msg service was registered before it. The `baseapp.SetUnguardedMsgServices` option lets named services of modules that have not adopted the guard register unguarded, each with a logged warning. There is no global off switch, and a chain without a breaker skips the check.
* **Upgrade handlers** run without the breaker in the context, because gov already approved the code and the breaker must not halt the chain at the upgrade height.
* A breaker that returns an error blocks the msg.

### Migrating from x/circuit

Chains remove `x/circuit` and add `x/safety` in the same upgrade: delete the circuit store, add the safety store, and swap the circuit keeper for the safety keeper in `SetCircuitBreaker` and the ante handler. The circuit disabled list and permissions do not migrate. After the upgrade, gov sets the operator and scope, and the operator re-trips anything that must stay blocked.

## Consequences

### Backwards Compatibility

* Existing chains need a store upgrade that adds the safety store key, and `x/circuit` chains also delete the circuit store.
* `state.Manager.SetState` gains a circuit breaker argument.
* On a chain with a breaker, every msg server must be wrapped in its generated guard or named in `SetUnguardedMsgServices`, or `RegisterServices` panics at startup. `SetCircuitBreaker` must run before `RegisterServices`.
* A plain refusal from a breaker returns `ErrUnauthorized`.
* `baseapp.CircuitBreaker` becomes a deprecated alias of `sdk.CircuitBreaker`.
* Group `MsgSubmitProposal` and `MsgVote` with `EXEC_TRY`, and `MsgCreateGroupWithPolicy`, fail when a msg they call internally is tripped, even though the outer msg passes `CheckTx`.

### Positive

* Under a selective trip, the breaker stops cosmos/evm precompile calls into SDK msg servers with no change to cosmos/evm. End-to-end tests in evmd show staking `delegate`, ERC20 `transfer` and distribution `withdrawDelegatorRewards` reverting under a trip, and `eth_call` and `eth_estimateGas` returning an error.
* Typos, empty lists and protected msgs fail at trip time instead of succeeding silently.
* No sequence of trips can block the reset or governance.
* Gov controls the operator and its scope through ordinary proposals, and can replace itself as authority.

### Negative

* Under a `*` trip the breaker blocks `MsgEthereumTx`, so during a global halt validators on an EVM chain vote with a Cosmos `MsgVote` tx, not through the gov precompile.
* Recheck evicts txs already in the mempool at trip time, and their senders rebroadcast after the reset.
* An authority set through the consensus `AuthorityParams` override that can act only through a wrapper the operator tripped by exact URL cannot recover on chain.
* `MsgCreateGroupWithPolicy` checks five inner msgs, up to 25 point lookups.

### Neutral

* Keeper calls that bypass msg servers stay outside the breaker: ibc-go transfer, packet-forward-middleware, wasmd raw IBC and burn, and EVM native value transfers. In cosmos/evm, distribution `claimRewards`, werc20 `deposit`, and the ics02 and ics20 precompiles fall in this group.
* Modules outside this repository adopt the guard by adding the plugin to their buf template and wrapping their msg servers. Until they do, chains with a breaker name their services in `SetUnguardedMsgServices`, and direct calls to them skip the breaker.
* The router checks IBC relayer msgs like any other msg. Under a trip that covers `MsgUpdateClient`, no light client on the chain updates and misbehavior submissions fail, and a client that expires during the trip needs gov `MsgRecoverClient`, which the router also refuses until reset. `/ibc.core.channel.v1.*` does not match `MsgTransfer`, so the chain keeps escrowing sends it cannot ack or time out. Inbound packets under a recv trip time out and refund on their source.
* The ICA host routes packet msgs, so it returns an error ack for a tripped msg and moves no funds, but CheckTx cannot see the msg and the `MsgRecvPacket` pays its fee.
* cosmos/evm runs Ethereum txs through an ante chain without `SafetyDecorator`, so under a `MsgEthereumTx` or `*` trip EVM txs enter the mempool, fail in `FinalizeBlock`, pay fees and use nonces. `eth_call` and `eth_estimateGas` do not check `MsgEthereumTx` itself and report success for those txs.
* wasmd routes contract bank sends, staking, gov, `Any`, Stargate and ICS20 msgs, so the router refuses them under a selective trip. `BankMsg::Burn` and `IbcMsg::SendPacket` call keepers directly, and a contract reading its `ReplyOn::Error` reply sees only `codespace: safety, code: 10`.

## Further Discussions

* Adoption of the generated guard in ibc-go, wasmd and cosmos/evm msg servers.
* A bank `SendRestriction` backed by the breaker, which would cover keeper-level transfers.
* A backport to v0.54.x for chains on that line, including cosmos/evm.

## Test Cases

* `x/safety` keeper, ante decorator and pattern tests, including fuzzing of pattern matching against the router.
* `baseapp` guard tests for proposal handlers, block hooks, vote extensions, the restarted node's check state, and Simulate running concurrently with Commit, and router tests for the startup check, the allowlist, a breaker set after registration, and the lookups each path into a guarded msg server makes.
* A golden test of `protoc-gen-go-msgguard`, a source scan that every in-repo Msg service has a generated guard and every registration uses it, and SimApp and both enterprise SimApps starting with a breaker and no allowlist.
* `tests/integration/safety`: direct msg server calls from inside `FinalizeBlock`, upgrade handlers under a trip, and group `EXEC_TRY`.
* System tests: `TestSafety*` and `TestChainUpgrade`, which upgrades a running v0.54 chain into x/safety.
* cosmos/evm: precompile tests in evmd against the real module, and a node-level script on a 4-validator testnet covering precompiles, `MsgEthereumTx` and `*` trips, native transfers and queued txs.
* ibc-go: relayed transfers, recv, ack, timeout and client update trips, client expiry and recovery, misbehavior, ICA host and packet-forward-middleware, all through real `MsgTrip` and `MsgReset` txs.
* wasmd: contract bank, staking, gov, `Any`, Stargate, burn, ICS20 and raw packet msgs, and submessage replies under trips.
* In each external repo, removing `SetCircuitBreaker`, the ante decorator or the msg server guard fails the tests that depend on it.

## References

* [#26894: x/safety module](https://github.com/cosmos/cosmos-sdk/pull/26894)
* [#26895: msg server guard](https://github.com/cosmos/cosmos-sdk/pull/26895)
* [`contrib/x/circuit`](../../contrib/x/circuit)
* [ADR 068: Preblock](./adr-068-preblock.md)
