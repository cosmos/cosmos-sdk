---
sidebar_position: 1
---

# `x/safety`

## Abstract

`x/safety` stops chosen msg types from executing without a chain halt or a
software upgrade. One operator address disables ("trips") msgs by exact type
URL, proto package prefix or `*` in a single tx, and the authority (governance
by default) or the operator re-enables ("resets") them. It replaces
`contrib/x/circuit`. The design and its alternatives are in
[ADR 077](../../docs/architecture/adr-077-x-safety.md).

## Contents

* [Concepts](#concepts)
    * [Roles](#roles)
    * [Patterns](#patterns)
    * [Protected set](#protected-set)
    * [Wrapper set](#wrapper-set)
    * [Enforcement](#enforcement)
    * [Gas](#gas)
    * [Direct callers](#direct-callers)
* [State](#state)
* [Messages](#messages)
* [Events](#events)
* [Queries](#queries)
* [Genesis](#genesis)
* [Wiring](#wiring)
* [Operating](#operating)
    * [Incident runbook](#incident-runbook)
    * [Before an incident](#before-an-incident)
    * [Commands](#commands)
* [IBC, EVM and wasm](#ibc-evm-and-wasm)
* [Known gaps](#known-gaps)
* [Changes from x/circuit](#changes-from-xcircuit)

## Concepts

### Roles

| Role | Stored in | Can |
| --- | --- | --- |
| authority | `params.authority`, defaults to the gov module account | update params, reset any disabled pattern |
| operator | `params.operator`, empty means no operator | trip and reset patterns covered by `params.scope` |

The authority check uses `sdk.ValidateAuthority`. Without the consensus
`AuthorityParams` override, `params.authority` is the authority, and the
current authority hands the role to a new address with `MsgUpdateParams`.
When a chain sets the override, the override account is the authority and
`params.authority` has no effect.

The authority cannot trip. To let governance trip directly, set the operator
to the authority's address.

Either role may act through an authz grantee or a group policy; see
[Wrapper set](#wrapper-set) and [Exact wrapper trips](#exact-wrapper-trips).

### Patterns

| Form | Example | Matches |
| --- | --- | --- |
| `*` | `*` | every msg |
| prefix | `/cosmos.bank.*` | every msg whose type URL starts with `/cosmos.bank.` |
| exact | `/cosmos.bank.v1beta1.MsgSend` | that msg type URL |

A prefix is a leading `/`, one or more dot-separated proto identifiers, then
`.*`. It matches at segment boundaries: `/cosmos.bank.*` covers
`/cosmos.bank.v1beta1.MsgSend` but not `/cosmos.banking.v1.MsgX`. A pattern is
at most 256 bytes.

A scope entry `S` covers a pattern `P` when `S` is `*`, `S` equals `P`, or `S`
is a prefix `X.*` and `P` starts with `X.`. Only `*` covers `*`.

### Protected set

Always allowed, whatever the disabled list contains:

* every type URL under `/cosmos.safety.`
* every type URL under `/cosmos.gov.`
* every type URL under `/cosmos.upgrade.`
* `/cosmos.consensus.v1.MsgUpdateParams`

The set is hardcoded. A `*` trip leaves it running.

### Wrapper set

These msgs only dispatch other msgs through the router, which checks each
dispatched msg:

* `/cosmos.authz.v1beta1.MsgExec`
* `/cosmos.group.v1.MsgSubmitProposal`
* `/cosmos.group.v1.MsgVote`
* `/cosmos.group.v1.MsgExec`

Only an exact disabled entry for its own URL blocks a wrapper. `*` and prefix
entries such as `/cosmos.authz.*` do not, in `IsAllowed`, in the ante
decorator, or when `MsgTrip` checks what a pattern blocks. Under `*` a wrapper
runs, and the router admits the msgs it dispatches from the protected set and
the other wrappers and refuses the rest. A wrapper dispatched this way still
writes state: under `*`, authz `MsgExec` runs a granted group `MsgVote`, and
the group counts the vote.

The operator trips a wrapper by its exact URL when its scope covers that URL;
`*` and `/cosmos.authz.*` both cover `/cosmos.authz.v1beta1.MsgExec`.

Group `MsgSubmitProposal` and `MsgVote` with `exec` set to `EXEC_TRY` run the
group's vote and exec without a group `MsgVote` or `MsgExec` reaching the
router. To halt group proposal execution, trip all three group wrappers.

To cut off a compromised grantee that trips with the operator's grant:

1. Trip `/cosmos.authz.v1beta1.MsgExec`.
2. Reset the grantee's trips. `*` blocks the granter's `MsgRevoke`.
3. Revoke the grant.
4. Reset `/cosmos.authz.v1beta1.MsgExec`.

### Enforcement

`Keeper.IsAllowed(ctx, typeURL)` implements `baseapp.CircuitBreaker`:

1. A protected type URL is allowed.
2. The type URL is blocked if the disabled list contains it exactly.
3. A type URL in the wrapper set is allowed.
4. The type URL is blocked if the disabled list contains `*` or any
   segment-boundary prefix of it. For `/cosmos.bank.v1beta1.MsgSend` the
   prefixes are `/cosmos.*`, `/cosmos.bank.*` and `/cosmos.bank.v1beta1.*`.
5. Otherwise it is allowed.

A blocked msg fails with `ErrMsgDisabled`, code 10 in the `safety` codespace,
naming the disabled entry that blocks it. The keeper logs
`msg refused by disabled pattern` at debug level with `type_url` and
`pattern`. Per check it runs at most two point lookups plus one per package
segment, and none for a protected type URL.

The check runs in two places.

**Router.** The `MsgServiceRouter` checks every msg it routes, in every exec
mode: tx msgs, msgs that authz `MsgExec`, gov and group proposals dispatch,
and msgs run with `msgservice.Dispatch`. The router reads the breaker on each
call, so `SetCircuitBreaker` can run before or after `RegisterServices`.

**Ante decorator.** `ante.CircuitBreakerDecorator` in `x/auth/ante`, which
`ante.HandlerOptions.CircuitBreaker` adds:

* checks every top-level msg and, recursively, every msg returned by
  `GetMessages() ([]sdk.Msg, error)` (authz `MsgExec`) or
  `GetMsgs() ([]sdk.Msg, error)` (gov v1 and group `MsgSubmitProposal`), up to
  `MaxNestedMsgDepth` levels, including inside protected and wrapper msgs.
* runs in `CheckTx`, `ReCheckTx` and simulation, and passes every tx through
  in `FinalizeBlock`.
* runs in `PrepareProposal` and `ProcessProposal` only with an app-side
  mempool. The SDK default `PrepareProposal` then leaves tripped txs out, and
  the default `ProcessProposal` rejects a block that contains one. With the
  default `NoOpMempool` neither handler runs the ante chain.
* checks against the disabled list as it stands before the tx, so a tx that
  resets a pattern and then sends a msg the pattern blocked fails. Send that
  msg in a later tx.

What happens to a tx that contains a tripped msg:

| Where | Result | Fee | Sequence |
| --- | --- | --- | --- |
| `CheckTx` | refused, never enters the mempool | not paid | kept |
| mempool at trip time | evicted by `ReCheckTx` (on by default in CometBFT), or left out by an app-side mempool's `PrepareProposal` | not paid | kept |
| `FinalizeBlock`, such as a tx later in the trip's block or on a node with recheck off | fails at the router | paid | used |

A tx that keeps its sequence executes if anyone broadcasts its signed bytes
after the reset. To cancel it, sign another tx at the same sequence. A tx
signed with `--timeout-duration` expires on its own.

A gov or group proposal that contains a tripped msg fails `CheckTx`. One that
reaches `FinalizeBlock` is created, because the router admits the protected
or wrapper msg, and fails at execution while the trip stands.

### Gas

The check charges no gas. The keeper runs its lookups on an infinite gas
meter, so a msg uses the same gas with x/safety wired as with no circuit
breaker set, in `FinalizeBlock` and in simulation. The same goes for
`msgservice.Dispatch` and EVM precompiles that call it.

### Direct callers

The router checks only the msgs it routes. A direct call to a msg server
method, such as `bankkeeper.NewMsgServerImpl(k).Send(ctx, msg)`, skips the
check. Code outside the module that owns a msg, such as an EVM precompile or
another module's keeper, runs the msg with `msgservice.Dispatch`:

```go
resp, err := msgservice.Dispatch[*banktypes.MsgSendResponse](ctx, app.MsgServiceRouter(), msg)
```

`Dispatch` runs the router handler on a branch of `ctx`: `ValidateBasic`, the
circuit breaker check and the msg server. It writes the branch and emits the
msg server's events only on success, and returns the typed response.

* The multistore's `CacheMultiStore` must return a branch. The cosmos/evm
  snapshot multistore returns itself; fix it before a precompile switches to
  `Dispatch`, or a reverted EVM call keeps the msg's writes.
* It emits no `message` event (`action`, `sender`, `module`) and no
  `msg_index` attribute.
* It does not check signers; the caller authorizes them.
* The msg must be the gogoproto type, not the pulsar type from
  `cosmossdk.io/api`. A msg with no handler fails with `ErrUnknownRequest`,
  and a nil router returns an error.

Keeper methods, such as authz `DispatchActions` or a bank keeper send, skip
the check. Upgrade handlers call keepers, not `Dispatch`, so a trip cannot
halt an upgrade.

## State

| Collection | Prefix | Key | Value |
| --- | --- | --- | --- |
| `Params` | `0x00` | | `Params{authority, operator, scope}` |
| `DisabledList` | `0x01` | pattern `string` | |

## Messages

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/safety/v1/tx.proto
```

Every msg runs stateless checks first: addresses decode with the chain address
codec, pattern lists are non-empty, patterns are well-formed and unique. Trip
and reset are all-or-nothing.

### MsgUpdateParams

Replaces the params. The signer must pass `sdk.ValidateAuthority`. The new
params need a valid authority, an empty or valid operator, and well-formed,
unique scope patterns. Both addresses must be in the canonical form the
address codec encodes. The msg never changes the disabled list: narrowing the
scope leaves existing entries in place, and the authority resets them.

### MsgTrip

Adds patterns to the disabled list. The signer must equal the non-empty
`params.operator`, in canonical form. Each pattern must:

* be covered by a scope entry,
* block at least one registered msg type URL that is not protected,
* not already be in the disabled list.

A wrapper counts only for its exact URL: `/cosmos.authz.*` is a valid trip
because it blocks `MsgGrant`, and it leaves `MsgExec` running.

A pattern under or beside a broader entry is accepted. With `*` disabled, the
operator trips `/cosmos.bank.v1beta1.MsgSend`, then resets `*`, and `MsgSend`
stays blocked. A broad pattern can also be tripped over narrower entries;
resetting it later leaves them in place.

### MsgReset

Removes patterns from the disabled list. A signer that passes the authority
check may reset any pattern. Otherwise the signer must be the non-empty
operator and each pattern must be covered by its scope. Each pattern must be
in the list exactly as given: resetting `/cosmos.bank.*` does not remove
`/cosmos.bank.v1beta1.MsgSend`.

## Events

| Event | Fields | Emitted by |
| --- | --- | --- |
| `cosmos.safety.v1.EventTrip` | `operator`, `patterns` | `MsgTrip` |
| `cosmos.safety.v1.EventReset` | `sender`, `patterns` | `MsgReset` |
| `cosmos.safety.v1.EventParamsUpdated` | `authority`, `params` | `MsgUpdateParams` |

## Queries

| gRPC | REST | Returns |
| --- | --- | --- |
| `Params` | `GET /cosmos/safety/v1/params` | the params |
| `DisabledList` | `GET /cosmos/safety/v1/disabled_list` | a page of disabled patterns |
| `IsAllowed` | `GET /cosmos/safety/v1/is_allowed?msg_type_url=...` | whether the type URL may execute and, if not, the blocking entry in `blocked_by` |
| `MatchingMsgs` | `GET /cosmos/safety/v1/matching_msgs?pattern=...` | a page of the registered msg type URLs a trip of the pattern would block, sorted |

`IsAllowed` returns `InvalidArgument` for anything but an exact msg type URL.

`MatchingMsgs` is the dry run for a trip. It applies the protected and wrapper
rules and ignores the operator scope and the disabled list.

## Genesis

```protobuf
message GenesisState {
  Params params = 1;
  repeated string disabled_patterns = 2;
}
```

The default genesis sets the authority passed to `NewKeeper`, no operator, an
empty scope and an empty disabled list. `InitGenesis` fails on invalid params
or a malformed or duplicate pattern. `ExportGenesis` output always imports.
`IsAllowed` ignores a protected type URL listed in genesis.

## Wiring

x/safety supports manual wiring only: no depinject module config, no autocli
options, and gogoproto types only, with no pulsar types in `cosmossdk.io/api`.

```go
app.SafetyKeeper = safetykeeper.NewKeeper(
	appCodec,
	runtime.NewKVStoreService(keys[safetytypes.StoreKey]),
	app.AccountKeeper.AddressCodec(),
	interfaceRegistry,
	authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
app.SetCircuitBreaker(app.SafetyKeeper)
```

Register `safety.NewAppModule(app.SafetyKeeper)` with the module manager, add
the store key, and add `safetytypes.ModuleName` to the init and export genesis
orders.

### Genesis order

Order `safety` after `genutil` in `InitGenesis`. Genutil delivers gentxs
through the router during `InitChain`. With `safety` first, a genesis entry
that covers `/cosmos.staking.v1beta1.MsgCreateValidator` blocks every gentx and
`InitChain` panics.

### Ante handler

Set `CircuitBreaker` in the `x/auth/ante` handler options. `NewAnteHandler`
puts `ante.CircuitBreakerDecorator` right after `SetUpContextDecorator`, ahead
of fee deduction:

```go
anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
	AccountKeeper:  app.AccountKeeper,
	BankKeeper:     app.BankKeeper,
	// ...
	CircuitBreaker: app.SafetyKeeper,
})
```

An app that builds its own chain adds
`ante.NewCircuitBreakerDecorator(app.SafetyKeeper)` after
`SetUpContextDecorator` and before `DeductFeeDecorator`, in every ante chain it
has, such as the cosmos/evm chain for Ethereum txs.

### Upgrade handler

A chain that adds the module in an upgrade adds `safetytypes.StoreKey` to
`StoreUpgrades.Added`; `RunMigrations` then runs the default genesis. To name
an operator in the same upgrade, set the params right after `RunMigrations`:

```go
app.UpgradeKeeper.SetUpgradeHandler(UpgradeName,
	func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		vm, err := app.ModuleManager.RunMigrations(ctx, app.Configurator(), fromVM)
		if err != nil {
			return nil, err
		}
		return vm, app.SafetyKeeper.SetParams(ctx, safetytypes.Params{
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			Operator:  "cosmos1...", // the operator account
			Scope:     []string{safetytypes.Wildcard},
		})
	},
)
```

`SetParams` validates the params first. The operator can trip from the
upgrade block on.

## Operating

### Incident runbook

1. Preview the trip: `q safety matching-msgs --pattern <pattern>`.
2. Trip with the operator key.
3. Confirm: `q safety disabled-list`, and `q safety is-allowed` for the
   exploit's msg, which names the blocking entry in `blocked_by`.
4. Patch first, then reset. Ship the fix, by software upgrade if needed (gov
   and upgrade msgs run under any trip), and reset only after it is live.
   Refused txs keep their sequence and execute if rebroadcast after the reset.

| Exploit path | Trip | Stays open | Watch |
| --- | --- | --- | --- |
| one msg type | its exact URL | everything else | keeper callers of the same logic below |
| one module | its prefix, such as `/cosmos.bank.*` | other modules, the module's wrappers | contracts and precompiles that call the module's keeper |
| ICS20 | `/ibc.applications.transfer.v1.MsgTransfer` and `/ibc.core.channel.v1.MsgRecvPacket`, exact | acks, timeouts, client updates, so in-flight packets settle | packet-forward-middleware retries a timed-out forward through the transfer keeper |
| IBC core or light client | `/ibc.*` or `/ibc.core.client.v1.*` | non-IBC msgs | client trusting periods: reset before the first client expires |
| EVM tx, precompile, precompile keeper call or native transfer | `/cosmos.evm.vm.v1.MsgEthereumTx`, exact | Cosmos txs, including gov `MsgVote` | EVM txs queue in the EVM mempool with no error and run after the reset |
| wasm `BankMsg::Burn` or `IbcMsg::SendPacket` | `/cosmwasm.wasm.v1.MsgExecuteContract`, exact, or `/cosmwasm.wasm.*` | non-wasm msgs | none |
| wasm contract sending a routed SDK msg | that SDK msg | the contract's other msgs | `ReplyOn::Error` contracts continue after the failure |
| compromised authz grantee | `/cosmos.authz.v1beta1.MsgExec`, exact | direct txs | see [Wrapper set](#wrapper-set) |
| unknown | `*` | protected set and wrappers | grantees under `*`; IBC client trusting periods; EVM validators vote with a Cosmos `MsgVote` |

Any trip that covers `MsgUpdateClient` (`/ibc.core.client.v1.*`, `/ibc.*` or
`*`) has a deadline: the shortest trusting period of the chain's clients. A
client that expires during the trip needs gov `MsgRecoverClient`, which the
router refuses until the reset.

### Before an incident

* Appoint the operator and grant its scope right after the upgrade that adds
  the module, with `SetParams` in the upgrade handler or a gov proposal.
* Confirm the node binary has the `tx safety` and `q safety` commands, or keep
  the tx JSON from [Without the generic commands](#without-the-generic-commands)
  ready.
* Put the ante check in every ante chain, including the Ethereum tx chain on
  an EVM chain.
* Write down the trusting period of every IBC client on the chain.

### Commands

x/safety defines no CLI code. `simd`, and any app whose root command builds
autocli from `runtimeservices.ExtractAutoCLIOptions`, derives generic commands
from the module's Msg and Query services. `--from` fills the signer field.

```shell
simd tx safety trip --patterns '/cosmos.bank.v1beta1.MsgSend,/cosmos.staking.*' \
  --from operator --chain-id <chain-id> --node <rpc> --fees 2000stake
simd tx safety reset --patterns /cosmos.bank.v1beta1.MsgSend --patterns '/cosmos.staking.*' \
  --from operator --chain-id <chain-id> --node <rpc> --fees 2000stake
simd tx safety update-params \
  --params '{"authority":"cosmos1...","operator":"cosmos1...","scope":["/cosmos.bank.*"]}' \
  --from authority --chain-id <chain-id> --node <rpc> --fees 2000stake
```

`--patterns` takes a comma-separated list, repeats, or both. Quote `*` and
prefix patterns. The authority resets with the same `reset` command. A gov
authority sends `MsgUpdateParams` or `MsgReset` in a proposal file for
`simd tx gov submit-proposal`.

```shell
simd q safety params --node <rpc>
simd q safety disabled-list --node <rpc>
simd q safety is-allowed --msg-type-url /cosmos.bank.v1beta1.MsgSend --node <rpc>
simd q safety matching-msgs --pattern '/cosmos.bank.*' --node <rpc>
```

For a blocked msg, `is-allowed` prints `blocked_by` and omits `allowed`, since
the output skips default values.

#### Without the generic commands

The binary still registers the msg types, so the operator signs tx JSON and
queries the REST gateway. Save as `trip.json`:

```json
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.safety.v1.MsgTrip",
        "operator": "cosmos1...",
        "patterns": ["/cosmos.bank.v1beta1.MsgSend", "/cosmos.staking.*"]
      }
    ]
  },
  "auth_info": {
    "fee": {"amount": [{"denom": "stake", "amount": "2000"}], "gas_limit": "200000"}
  },
  "signatures": []
}
```

```shell
simd tx sign trip.json --from operator --chain-id <chain-id> --node <rpc> > trip.signed.json
simd tx broadcast trip.signed.json --node <rpc>
curl '<api>/cosmos/safety/v1/is_allowed?msg_type_url=/cosmos.bank.v1beta1.MsgSend'
```

A reset uses `"@type": "/cosmos.safety.v1.MsgReset"` and `"sender"` in place
of `"operator"`.

## IBC, EVM and wasm

### IBC relaying

Relayers deliver packets, acks, timeouts and client updates as ordinary msgs,
and the router checks each one.

| Exact trip | Effect |
| --- | --- |
| `/ibc.core.channel.v1.MsgRecvPacket` | Inbound packets wait. One that times out refunds on its source chain; one delivered after the reset and before its timeout executes. |
| `/ibc.core.channel.v1.MsgAcknowledgement` | Acks fail, so the chain refunds an error-acked packet only after the reset. |
| `/ibc.core.client.v1.MsgUpdateClient` | No light client on the chain updates, so recv and ack fail proof verification. Misbehavior submissions fail too, since ibc-go v11 takes misbehavior in `MsgUpdateClient`. |

`/ibc.core.channel.v1.*` does not cover
`/ibc.applications.transfer.v1.MsgTransfer`, so the chain keeps escrowing
sends it can neither ack nor time out until the reset. Add the `MsgTransfer`
URL or trip `/ibc.*`.

x/safety does not protect `MsgRecoverClient`, so while a trip covers it the
ante decorator refuses a gov proposal that wraps it, and a proposal that
reaches a block fails at execution.

Under `*` every IBC msg fails. After the reset, packets inside their timeout
relay and ack, and packets past it refund on their source.

### ICS20 transfers

The router refuses a routed `MsgTransfer` under a trip. packet-forward-middleware
calls the transfer keeper's `Transfer` method directly, so a forward still
escrows and sends. The cosmos/evm ICS20 precompile reverts with nothing
escrowed once it runs the msg with `msgservice.Dispatch`.

### Interchain account hosts

The ICA host routes packet msgs through the router. For a tripped msg the host
returns an error ack, `ABCI code: 10: error handling packet: see events for
details`, and moves no funds; the x/safety reason appears only in the host
chain's events. `CheckTx` cannot see msgs inside packet data, so the
`MsgRecvPacket` succeeds and pays its fee. Protected msgs run: under
`/cosmos.*` a gov `MsgVote` sent through an interchain account executes.

### EVM chains

cosmos/evm sends every EVM tx as `/cosmos.evm.vm.v1.MsgEthereumTx`. Under an
exact trip of it, or `*`, the router refuses every EVM tx, precompile calls
included.

* Without `CircuitBreakerDecorator` in the Ethereum tx ante chain, tripped EVM
  txs reach a block, fail with code 10, pay their fee, use their nonce and
  return no receipt.
* With it, `CheckTx` returns code 0 and `eth_sendRawTransaction` a tx hash,
  because the EVM mempool queues the tx before the ante chain runs. The tx
  never becomes pending, pays no fee and uses no nonce, and becomes pending
  after the reset.
* `eth_call` and `eth_estimateGas` route no `MsgEthereumTx`, so under its trip
  a value transfer estimates at 21000 gas and a gov precompile vote simulates
  as a success. A precompile that runs a tripped msg with `Dispatch` fails in
  both.
* An EVM value transfer moves the native denom without a bank msg, so a
  `/cosmos.bank.v1beta1.MsgSend` trip does not stop it.

### CosmWasm contracts

wasmd routes `BankMsg::Send`, staking, gov, `CosmosMsg::Any`, Stargate and
`IbcMsg::Transfer` msgs through the router, so a tripped SDK msg fails the
contract execution and moves no funds. `CheckTx` admits the outer
`MsgExecuteContract`. Gov msgs stay protected.

`BankMsg::Burn` and `IbcMsg::SendPacket` call keepers directly and skip a
selective trip. The bank module defines no `MsgBurn`, so a trip of
`/cosmos.bank.v1beta1.MsgBurn` fails with code 7.

A contract that sends a tripped msg as a submessage with `ReplyOn::Error`
continues, and sees only `codespace: safety, code: 10` in its reply.

## Known gaps

### Direct keeper calls

Code that calls a keeper or a msg server directly skips a selective trip.
Known paths:

* cosmos/evm precompile methods that call SDK keepers, such as distribution
  `claimRewards` and WERC20 `deposit`,
* EVM native value transfers,
* the ibc-go transfer keeper's `Transfer` and `SendTransfer`, which
  packet-forward-middleware calls,
* the ibc-go channel keeper's `SendPacket`,
* wasmd `BankMsg::Burn` and `IbcMsg::SendPacket`.

A `*` trip stops each of these at the tx msg that starts it.

### Exact wrapper trips

An authority that acts only through a wrapper loses that path while the
operator keeps an exact trip of the wrapper:

* an authz grantee, under a trip of `/cosmos.authz.v1beta1.MsgExec`;
* a group policy with no reset proposal pending, under a trip of group
  `MsgSubmitProposal`;
* a group whose decision policy needs votes from non-proposers, under a trip
  of group `MsgVote`.

The path returns when the operator resets the trip. Otherwise recovery needs a
gov proposal that sets the consensus `AuthorityParams` override, or a software
upgrade. When the locked-out authority is the override itself, neither works,
and recovery needs validators to switch to a new binary off chain. Use gov or
a directly signing account as the authority and as the override.

### Grantees under `*`

A `*` trip leaves authz `MsgExec` running, so a grantee can still run the
protected msgs and wrappers it was granted, such as a gov deposit from the
granter's funds, while `*` blocks the granter's `MsgRevoke`. Trip
`/cosmos.authz.v1beta1.MsgExec` by its exact URL to stop it.

### Interchain accounts

An authority that acts through an interchain account needs `MsgRecvPacket` and
client updates, which `*` and `/ibc.*` block. Use gov, a directly signing
account or a group policy as the authority.

## Changes from x/circuit

| x/circuit | x/safety |
| --- | --- |
| Empty type URL lists succeeded and changed nothing | Empty lists are rejected |
| Any string could be tripped | A pattern must match a registered, unprotected msg |
| `MsgResetCircuitBreaker` could be tripped | Protected set, always allowed |
| Permission levels table and super admin delegation | One operator in params |
| Authority fixed at build time, compared with `bytes.Equal` | Authority in params, checked with `sdk.ValidateAuthority` |
| Revoked permissions left state that failed genesis validation | Genesis export always imports |
| Untyped events with comma-joined attributes | Typed proto events |
| Unregistered errors | Registered errors |
| Unpaginated disabled list query | Paginated |
| Ante decorator checked top-level msgs only | Recurses into nested msgs |
| Exact type URLs only | Exact, prefix and `*` patterns |
