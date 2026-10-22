# ADR 077: x/safety, an Emergency Msg Breaker

## Changelog

* 2026-10-21: Initial draft
* 2026-10-21: Direct msg server callers use `msgservice.Dispatch`
* 2026-10-22: Msgs run only through the router; SDK modules stop exporting their msg servers

## Status

DRAFT Not Implemented

## Abstract

`x/safety` replaces `contrib/x/circuit`. It lets one operator address stop a msg type in a single tx during an incident, and lets governance decide who that operator is and what it may stop. The authority, gov by default, sets the operator and its scope and can reset any block, but cannot block anything itself. Gov, upgrade, safety and consensus parameter msgs stay outside the operator's reach, so the operator can never lock governance out of the reset. BaseApp checks the breaker at the router and in the ante handler. The router becomes the only way to run an `sdk.Msg`: SDK modules stop exporting their msg servers, and code outside a module, such as a cosmos/evm precompile, runs a msg with `msgservice.Dispatch`, which the router checks as it checks a tx.

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

* **Router.** The keeper implements `baseapp.CircuitBreaker`, and the module wires it with `BaseApp.SetCircuitBreaker`. The router checks every msg it routes, on every call and in every exec mode, including `CheckTx`, `Simulate`, `FinalizeBlock`, block hooks, vote extensions and query contexts. It reads the breaker on each call, so a breaker set after `RegisterServices` still covers every handler.
* **Ante.** `SafetyDecorator` checks every msg in a tx and recurses into any msg that exposes `GetMessages` or `GetMsgs`, up to a depth limit. It runs in `CheckTx`, `ReCheckTx`, `Simulate` and the proposal handlers, and passes txs through in `FinalizeBlock`. A tx included after a trip then spends its sequence and fails at the router, so nobody can replay it once the operator resets. A proposal containing a tripped msg fails at submission.
* **One way to run a msg.** An `sdk.Msg` runs only through the `MsgServiceRouter`. In every SDK module the msg server type and its constructor are unexported, and the keeper package exports only `RegisterMsgServer(r grpc.ServiceRegistrar, k Keeper)`, which `RegisterServices` calls. The authz, group, consensus, crisis, nft and counter keepers stop implementing their Msg service; their Msg methods move to an unexported type. `MsgServiceRouter.HybridHandlerByMsgName`, which ran a handler by name without `ValidateBasic`, is removed. A test in `keeper_test` reaches the constructor through `export_test.go`.
* **Dispatch.** Code outside the module that owns a msg, such as a cosmos/evm precompile or another module's keeper, runs it with `msgservice.Dispatch[R](ctx, router, msg)`. `Dispatch` calls the router handler on a branch of `ctx`, so `ValidateBasic`, the breaker check and the msg server's events apply, writes the branch only on success, and returns the typed response. It does not check signers; the caller authorizes them as before.
* **Keepers for module logic.** Logic another module needs, such as authz `DispatchActions` or a bank send, stays an exported keeper method, and the breaker does not see it. Upgrade handlers call keepers, not `Dispatch`, so a trip cannot halt an upgrade. A module's calls to its own handlers, such as group running `Exec` after a vote, stay internal calls on the unexported type.
* A caller that registers a fake `grpc.ServiceRegistrar` can still capture the msg server it receives. That is deliberate abuse, not an API.
* A breaker that returns an error blocks the msg.

### Migrating from x/circuit

Chains remove `x/circuit` and add `x/safety` in the same upgrade: delete the circuit store, add the safety store, and swap the circuit keeper for the safety keeper in `SetCircuitBreaker` and the ante handler. The circuit disabled list and permissions do not migrate. After the upgrade, gov sets the operator and scope, and the operator re-trips anything that must stay blocked.

## Consequences

### Backwards Compatibility

* Existing chains need a store upgrade that adds the safety store key, and `x/circuit` chains also delete the circuit store.
* API break in every SDK module with a Msg service: `NewMsgServerImpl` is unexported, keepers that were their own msg server lose their Msg methods, each keeper package gains `RegisterMsgServer`, and `MsgServiceRouter.HybridHandlerByMsgName` is removed. `x/staking/testutil.NewHelper` takes a `msgservice.MsgRouter`.
* `msgservice.Dispatch` is new API.
* A caller that switches to `Dispatch` needs the app's `MsgServiceRouter`, so its constructor gains a router argument, and a multistore whose `CacheMultiStore` returns itself instead of a branch must be fixed first. The cosmos/evm snapshot multistore is one; without the fix a reverted EVM call keeps the msg's writes.

### Positive

* Under a selective trip, precompiles that run their SDK msgs with `Dispatch` revert. In evmd with the precompiles switched to `Dispatch`, staking `delegate`, ERC20 and WERC20 `transfer`, distribution `withdrawDelegatorRewards`, slashing `unjail` and ICS20 `transfer` revert under a trip, and `eth_call` and `eth_estimateGas` return an error. Reverting the precompiles to direct msg server calls fails every one of those tests.
* Module authors write no breaker code, and a new msg server needs nothing beyond `RegisterMsgServer`.
* A direct call to another module's msg server no longer compiles, so no in-repo or downstream caller can skip the breaker by accident.
* Downstream migration is small. cosmos/evm deleted a deprecated `erc20.NewMsgServerImpl` wrapper around the bank msg server and moved 2 test calls to `Dispatch`; its precompiles already used `Dispatch`. ibc-go needed no change. wasmd moved 5 test calls and no production code.
* Typos, empty lists and protected msgs fail at trip time instead of succeeding silently.
* No sequence of trips can block the reset or governance.
* Gov controls the operator and its scope through ordinary proposals, and can replace itself as authority.

### Negative

* Every chain and module that called an SDK msg server from code must move to `Dispatch` or a keeper call before it builds against this release.
* A module outside the SDK that still exports its msg server, and any keeper method that runs a msg's logic, remain direct paths the breaker does not see.
* Under a `*` trip the breaker blocks `MsgEthereumTx`, so during a global halt validators on an EVM chain vote with a Cosmos `MsgVote` tx, not through the gov precompile.
* Recheck evicts txs already in the mempool at trip time, and their senders rebroadcast after the reset.
* An authority set through the consensus `AuthorityParams` override that can act only through a wrapper the operator tripped by exact URL cannot recover on chain.

### Neutral

* Keeper calls stay outside the breaker: packet-forward-middleware's call to the transfer keeper's `Transfer` method, wasmd raw IBC and burn, EVM native value transfers, and the cosmos/evm distribution `claimRewards` and WERC20 `deposit` precompile methods.
* The router checks IBC relayer msgs like any other msg. Under a trip that covers `MsgUpdateClient`, no light client on the chain updates and misbehavior submissions fail, and a client that expires during the trip needs gov `MsgRecoverClient`, which the router also refuses until reset. `/ibc.core.channel.v1.*` does not match `MsgTransfer`, so the chain keeps escrowing sends it cannot ack or time out. Inbound packets under a recv trip time out and refund on their source.
* The ICA host routes packet msgs, so it returns an error ack for a tripped msg and moves no funds, but CheckTx cannot see the msg and the `MsgRecvPacket` pays its fee.
* cosmos/evm runs Ethereum txs through a separate ante chain. Without `SafetyDecorator` in it, under a `MsgEthereumTx` or `*` trip EVM txs fail in `FinalizeBlock`, pay fees and use nonces. With it, `CheckTx` and `eth_sendRawTransaction` still answer success, because the EVM mempool queues the tx before the ante chain runs, but the tx never becomes pending and pays no fee. `eth_call` and `eth_estimateGas` do not check `MsgEthereumTx` itself and report success for a value transfer.
* wasmd routes contract bank sends, staking, gov, `Any`, Stargate and ICS20 msgs, so the router refuses them under a selective trip. `BankMsg::Burn` and `IbcMsg::SendPacket` call keepers directly, and a contract reading its `ReplyOn::Error` reply sees only `codespace: safety, code: 10`.

## Further Discussions

* Switching cosmos/evm precompiles and packet-forward-middleware to `Dispatch`.
* A bank `SendRestriction` backed by the breaker, which would cover keeper-level transfers.
* A backport to v0.54.x for chains on that line, including cosmos/evm.

## Test Cases

* `x/safety` keeper, ante decorator and pattern tests, including fuzzing of pattern matching against the router.
* `types/msgservice` tests of `Dispatch`: the typed response, a tripped msg, an unknown msg, a pulsar msg, a wrong response type, a nil router, and a failed msg leaving `ctx` unchanged.
* `baseapp` tests that the router checks `Dispatch` in `CheckTx`, `Simulate`, `FinalizeBlock`, block hooks, vote extensions and a query context, and hybrid handler tests for the registered msg type and a breaker set or removed after registration.
* `tests/integration/safety`: `Dispatch` from block hooks, a `FinalizeBlock` tx and a query context in SimApp under a trip and after the reset.
* `types/msgservice`: a guard test that fails when any SDK package exports a msg server type, constructor, alias or Msg handler func.
* System tests: `TestSafety*` and `TestChainUpgrade`, which upgrades a running v0.54 chain into x/safety.
* cosmos/evm: precompile tests in evmd with the precompiles switched to `Dispatch`, ante tests of `SafetyDecorator` under `MsgEthereumTx` and `*` trips, and a node-level script on a 4-validator testnet covering precompiles, `MsgEthereumTx` and `*` trips, native transfers and queued txs.
* ibc-go: relayed transfers, recv, ack, timeout and client update trips, client expiry and recovery, misbehavior, ICA host and packet-forward-middleware, all through real `MsgTrip` and `MsgReset` txs.
* wasmd: contract bank, staking, gov, `Any`, Stargate, burn, ICS20 and raw packet msgs, and submessage replies under trips.
* In each external repo, removing `SetCircuitBreaker` or the ante decorator fails the tests that depend on it, and in cosmos/evm so does reverting the precompiles to direct msg server calls.

## References

* [#26894: x/safety, `msgservice.Dispatch` and unexported msg servers](https://github.com/cosmos/cosmos-sdk/pull/26894), which folds in #26892 and #26898
* [`contrib/x/circuit`](../../contrib/x/circuit)
* [ADR 068: Preblock](./adr-068-preblock.md)
