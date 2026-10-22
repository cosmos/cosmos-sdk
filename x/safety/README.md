---
sidebar_position: 1
---

# `x/safety`

## Abstract

`x/safety` lets a chain stop specific msg types from executing, fast, without a
chain halt or a software upgrade. One operator address can disable ("trip") msgs
by exact type URL or by proto package prefix in a single tx, and the authority
(governance by default) or the operator can re-enable ("reset") them.

The module replaces `contrib/x/circuit`. It keeps what circuit got right:
collections state and the BaseApp `CircuitBreaker` hook. It replaces the
permission levels table with a single operator,
moves the authority into params, and adds a protected set so that a trip can
never block its own reset.

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
current authority can hand the role to a new address with `MsgUpdateParams`.
When a chain sets the override, the override account is the authority and
`params.authority` has no effect, so only a change of the override moves the
role.

The authority cannot trip. A chain that wants governance to trip directly sets
the operator to the same address as the authority.

Either role may be an account that acts through an authz grantee or a group
policy. Authz `MsgExec` and the group proposal msgs are in the
[wrapper set](#wrapper-set), which a `*` or prefix trip does not block, so
those paths reach `MsgReset` and votes during an incident. Exact trips of the
wrappers do block them; see [Known gaps](#known-gaps).

### Patterns

A pattern takes one of three forms. Nothing else is valid.

| Form | Example | Matches |
| --- | --- | --- |
| `*` | `*` | every msg |
| prefix | `/cosmos.bank.*` | every msg whose type URL starts with `/cosmos.bank.` |
| exact | `/cosmos.bank.v1beta1.MsgSend` | that msg type URL |

A prefix pattern is a leading `/`, one or more dot-separated proto identifiers,
then `.*`. Prefixes match at segment boundaries: `/cosmos.bank.*` covers
`/cosmos.bank.v1beta1.MsgSend` but not `/cosmos.banking.v1.MsgX`. A pattern is
at most 256 bytes.

A scope entry `S` covers a pattern `P` when `S` is `*`, `S` equals `P`, or `S`
is a prefix `X.*` and `P` starts with `X.`. Only `*` covers `*`.

### Protected set

These msg type URLs are hardcoded and always allowed, whatever the disabled
list contains:

* every type URL under `/cosmos.safety.`
* every type URL under `/cosmos.gov.`
* every type URL under `/cosmos.upgrade.`
* `/cosmos.consensus.v1.MsgUpdateParams`

`MsgReset`, `MsgUpdateParams`, governance and upgrades always execute when
signed directly or run by x/gov, so no trip and no genesis file can lock gov,
a directly signing authority, or the operator out of recovery. A trip of `*`
still leaves this set running.

### Wrapper set

These msg type URLs are hardcoded. Each one only dispatches other msgs through
the router, which checks every msg it dispatches:

* `/cosmos.authz.v1beta1.MsgExec`
* `/cosmos.group.v1.MsgSubmitProposal`
* `/cosmos.group.v1.MsgVote`
* `/cosmos.group.v1.MsgExec`

Only an exact disabled entry for its own URL blocks a wrapper. `*` and prefix
entries such as `/cosmos.authz.*` or `/cosmos.group.*` do not, in `IsAllowed`,
in the ante decorator, or when `MsgTrip` checks what a pattern blocks. Under
`*`, a wrapper still runs, and the router admits the msgs it dispatches from
the protected set and the other wrappers and refuses the rest. An authority,
operator or validator that acts through an authz grantee or a group policy
still sends `MsgReset`, `MsgUpdateParams` and gov votes. A wrapper dispatched
this way still writes state: under `*`, authz `MsgExec` runs a granted group
`MsgVote`, and the group counts the vote.

Wrappers are not protected. The operator trips one by its exact URL when its
scope covers that URL; a scope of `*` or `/cosmos.authz.*` covers
`/cosmos.authz.v1beta1.MsgExec`. The exact trip halts that wrapper's handler,
and `MsgTrip` accepts it while `*` or a prefix entry stands.

The group wrappers do not each stand alone. Group `MsgSubmitProposal` and
`MsgVote` with `exec` set to `EXEC_TRY` call the group keeper's vote and exec
directly, without a group `MsgVote` or `MsgExec` reaching the router. To halt
group proposal execution, trip all three group wrappers. The router still
checks every msg in a proposal, whichever path executes it.

To cut off a compromised grantee that trips with the operator's grant:

1. Trip `/cosmos.authz.v1beta1.MsgExec`.
2. Reset the grantee's trips. The granter's `MsgRevoke` is an ordinary authz
   msg, so `*` blocks it.
3. Revoke the grant.
4. Reset `/cosmos.authz.v1beta1.MsgExec`.

### Enforcement

`Keeper.IsAllowed(ctx, typeURL)` implements `baseapp.CircuitBreaker`:

1. A protected type URL is allowed.
2. The type URL is blocked if the disabled list contains the exact type URL.
3. A type URL in the wrapper set is allowed.
4. The type URL is blocked if the disabled list contains `*` or any
   segment-boundary prefix of it. For `/cosmos.bank.v1beta1.MsgSend` the
   prefixes are `/cosmos.*`, `/cosmos.bank.*` and `/cosmos.bank.v1beta1.*`.
5. Otherwise it is allowed.

A blocked msg returns `false` with `ErrMsgDisabled`, naming the disabled entry
that blocks it. The router passes that error on unchanged, so a msg blocked at
execution fails with code 10 in the `safety` codespace, the same error the ante
decorator returns.

A check of a protected type URL runs no lookup. Any other check runs at most
two point lookups plus one per package segment, and one for a wrapper; it
stops at the first disabled entry that blocks the msg, and never walks the
disabled list.

The check runs in two places:

* The `MsgServiceRouter` calls it for every msg it routes, including msgs that
  authz `MsgExec`, gov proposals and other modules dispatch through the router,
  and msgs run with `msgservice.Dispatch`. This is the authoritative gate.
* The optional `ante.CircuitBreakerDecorator` from `x/auth/ante`, which
  `ante.HandlerOptions.CircuitBreaker` adds, calls it for every top-level msg
  and, recursively, every msg returned by `GetMessages() ([]sdk.Msg, error)`
  (authz `MsgExec`) or `GetMsgs() ([]sdk.Msg, error)` (gov v1 and group
  `MsgSubmitProposal`), up to `MaxNestedMsgDepth` levels. It opens protected
  wrappers such as gov `MsgSubmitProposal` too, so a proposal containing a
  tripped msg fails `CheckTx`, and it opens a wrapper that `*` lets through, so
  authz `MsgExec` wrapping a tripped msg fails too. Placed before fee deduction,
  it turns away tripped txs in `CheckTx`, so they never enter the mempool and
  never pay fees. A tx turned away in `CheckTx` keeps its sequence, so its
  signed bytes stay valid: anyone with a copy, such as the RPC node that
  received it, can broadcast it after the reset, and it executes. To cancel it,
  sign another tx at the same sequence. A tx signed with a timeout, such as
  `--timeout-duration`, expires on its own. A tx that entered the mempool
  before a trip meets the decorator again: CometBFT rechecks the mempool after
  every commit by default, and `ReCheckTx` evicts the tx, and with an app-side
  mempool `PrepareProposal` removes it. Either way it pays no fee and keeps its
  sequence, so, as with a tx turned away in `CheckTx`, its signed bytes execute
  after the reset unless the sender signs another tx at the same sequence or
  the tx times out. In `FinalizeBlock` the decorator passes every tx on to the
  router. A tx that reaches `FinalizeBlock` after a trip, such as one later in
  the same block as the trip or any tx on a node with recheck off, fails at the
  router, pays its fee and uses its sequence, so its signed bytes cannot
  execute again once the trip is reset. That holds with the default
  `NoOpMempool`, where `ProcessProposal` accepts every block. With an
  app-side mempool, `ProcessProposal` runs the ante chain on every tx in the
  proposal, and the SDK's default handler rejects the whole proposal when one
  tx fails, so validators reject a block that contains a tripped msg. The
  SDK's default `PrepareProposal` runs the same decorator and leaves tripped
  txs out; a proposer whose `PrepareProposal` skips the ante chain can have
  its block rejected. Gov or group still creates a proposal
  submitted in such a tx, because the router admits the wrapper and checks the
  msgs only when the proposal executes; while the trip stands, the router
  blocks the tripped msg there and the proposal fails. The
  decorator checks every msg in a tx against the disabled list as it stands
  before the tx runs, so a tx that resets a pattern and then sends a msg the
  pattern blocked fails in `CheckTx`. Send that msg in a later tx.

### Gas

The check charges no gas. The keeper runs its lookups on an infinite gas
meter that replaces the context's meter for the check only, so neither the tx
nor the block gas meter pays for them. A msg uses the same gas with x/safety
wired as with no circuit breaker set, both in `FinalizeBlock` and in
simulation, so `--gas auto` estimates do not rise. The same holds for
`msgservice.Dispatch` and for EVM precompiles that call it.

Charging for the check would make every msg pay for the breaker whether or
not anything is tripped. It would also inflate simulated gas: the ante
decorator checks in `CheckTx` and simulation but not in `FinalizeBlock`, so
users would pay fees for gas that never runs. The work stays bounded without
gas: per msg, at most two lookups plus one per package segment, on a type URL
that the interface registry knows or, through the `IsAllowed` query, one of
at most 256 bytes.

### Direct callers

The router checks only the msgs it routes. A direct call to a msg server
method, such as `bankkeeper.NewMsgServerImpl(k).Send(ctx, msg)`, skips the
check, so a trip does not stop it. Code outside the module that owns a msg,
such as an EVM precompile or another module's keeper, runs the msg with
`msgservice.Dispatch` instead:

```go
resp, err := msgservice.Dispatch[*banktypes.MsgSendResponse](ctx, app.MsgServiceRouter(), msg)
```

`Dispatch` calls the router handler for the msg on a branch of `ctx`: it runs
`ValidateBasic`, the circuit breaker check and the msg server, and returns the
typed response. It uses the same gas as a routed msg, which is the same gas
as with no circuit breaker set, and a tripped msg fails
with `ErrMsgDisabled`. The router checks on every call whatever the exec mode,
so the check applies to `Dispatch` from a tx in `CheckTx`, simulation or
`FinalizeBlock`, from `PreBlock`, `BeginBlock` and `EndBlock`, from vote
extensions, and from query contexts such as the one `eth_call` runs in.

`Dispatch` writes the branch and emits the msg server's events on `ctx` only
when the msg succeeds, so a failed `Dispatch` leaves `ctx` unchanged. This
needs a multistore whose `CacheMultiStore` returns a branch. The cosmos/evm
snapshot multistore returns itself, so fix that store before a precompile
switches to `Dispatch`, or a reverted EVM call keeps the msg's writes.
`Dispatch` does not emit the `message` event (`action`, `sender`, `module`) or
the `msg_index` attribute that a msg in a tx gets, so an indexer keyed on
`message.action` does not see it.

`Dispatch` does not check the msg's signers: the caller authorizes them. The
msg must be the gogoproto type, not the pulsar type from `cosmossdk.io/api`.
A msg with no registered handler fails with `ErrUnknownRequest`, and a nil
router returns an error.

Module logic that another module needs, such as authz `DispatchActions` or a
bank keeper send, is a keeper method, and a trip does not cover it. Upgrade
handlers call keepers, not `Dispatch`, so a trip cannot halt an upgrade. A
module's calls to its own handlers, such as group running `Exec` after a vote
or the gov v1beta1 server wrapping the v1 server, stay internal to the module.

## State

| Collection | Prefix | Key | Value |
| --- | --- | --- | --- |
| `Params` | `0x00` | | `Params{authority, operator, scope}` |
| `DisabledList` | `0x01` | pattern `string` | |

## Messages

Every msg runs stateless checks first: addresses decode with the chain address
codec, pattern lists are non-empty, patterns are well-formed and unique. Trip
and reset are all-or-nothing: if any pattern fails, the msg changes nothing.

### MsgUpdateParams

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/safety/v1/tx.proto
```

Replaces the params. The signer must pass `sdk.ValidateAuthority` against the
current `params.authority`. The new params need a valid authority, an empty or
valid operator, and well-formed, unique scope patterns. Both addresses must be
in canonical form (the string the address codec encodes), because
`sdk.ValidateAuthority` compares strings. The msg never changes the disabled
list; narrowing the scope leaves existing disabled entries in place, and the
authority must reset them explicitly.

### MsgTrip

Adds patterns to the disabled list. The signer must equal the non-empty
`params.operator`, in canonical form, so `EventTrip` names the operator exactly
as params does. Each pattern must:

* be covered by a scope entry,
* block at least one msg type URL registered in the `InterfaceRegistry` that
  is not protected, so typos and no-op trips fail,
* not already be in the disabled list.

A wrapper counts only for its exact URL. `/cosmos.authz.*` is a valid trip
because it blocks `MsgGrant` and the other authz msgs, and it leaves `MsgExec`
running.

A pattern under a broader disabled entry is accepted, and so is a pattern
beside a broader one in the same msg. The operator uses this to pin a msg
before the broad entry is reset: with `*` disabled, it trips
`/cosmos.bank.v1beta1.MsgSend`, then resets `*`, and `MsgSend` never reopens.
After the authority narrows the scope below a broad trip, the operator pins
the msgs it still covers the same way before the authority resets the broad
entry.

A broad pattern can also be tripped over narrower entries already in the list.
Resetting it later leaves those entries in place; check `DisabledList` after a
reset.

### MsgReset

Removes patterns from the disabled list. If the signer passes the authority
check it may reset any pattern. Otherwise the signer must be the non-empty
operator and each pattern must be covered by its scope. Both roles match the
signer string exactly, as `sdk.ValidateAuthority` does. Each pattern must be in
the disabled list exactly as given: resetting `/cosmos.bank.*` does not remove
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
| `IsAllowed` | `GET /cosmos/safety/v1/is_allowed?msg_type_url=...` | whether the type URL may execute |

`IsAllowed` returns `InvalidArgument` for anything other than an exact msg type
URL, such as a URL without the leading `/`, a prefix pattern or `*`.

## Genesis

```protobuf
message GenesisState {
  Params params = 1;
  repeated string disabled_patterns = 2;
}
```

The default genesis sets the authority to the one passed to `NewKeeper`, no
operator, an empty scope, and an empty disabled list. An operator the authority appoints can trip and reset
nothing until the authority also grants it scope. `ValidateGenesis` and
`InitGenesis` check the params and that every disabled pattern is well-formed
and unique; `InitGenesis` fails on an invalid state. `ExportGenesis` output
always imports cleanly. Genesis may list a protected type URL; `IsAllowed`
ignores it.

## Wiring

x/safety supports manual wiring only. It has no depinject module config and
no autocli options, and its proto types are gogoproto only, with no pulsar
types in `cosmossdk.io/api`.

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

The router reads the breaker on every call, so `SetCircuitBreaker` can run
before or after `ModuleManager.RegisterServices`.

Register `safety.NewAppModule(app.SafetyKeeper)` with the module manager, add
the store key, and add `safetytypes.ModuleName` to the init and export genesis
orders, after `genutiltypes.ModuleName` in the init order. A chain that adds
the module in an upgrade adds `safetytypes.StoreKey` to
`StoreUpgrades.Added`; `RunMigrations` then runs the default genesis.

### Genesis order

Order `safety` after `genutil` in `InitGenesis`. Genutil delivers gentxs
through the router during `InitChain`, and the router checks them against the
disabled list. With `safety` first, a genesis entry that covers
`/cosmos.staking.v1beta1.MsgCreateValidator`, such as `*` or
`/cosmos.staking.*`, blocks every gentx and `InitChain` panics. With `safety`
after `genutil`, gentxs run against an empty disabled list.

### Ante handler

Set `CircuitBreaker` in the `x/auth/ante` handler options. `NewAnteHandler`
then puts `ante.CircuitBreakerDecorator` right after `SetUpContextDecorator`,
ahead of fee deduction:

```go
anteHandler, err := ante.NewAnteHandler(ante.HandlerOptions{
	AccountKeeper:   app.AccountKeeper,
	BankKeeper:      app.BankKeeper,
	// ...
	CircuitBreaker: app.SafetyKeeper,
})
```

An app that builds its own chain adds
`ante.NewCircuitBreakerDecorator(app.SafetyKeeper)` after
`SetUpContextDecorator` and before `DeductFeeDecorator`. An app with more than
one ante chain, such as cosmos/evm with its separate chain for Ethereum txs,
must add the decorator to each, or the txs that chain handles reach a block
and pay fees; see [EVM chains](#evm-chains).

## Operating

x/safety defines no CLI code. `simd`, and any app whose root command builds
autocli from `runtimeservices.ExtractAutoCLIOptions`, derives generic commands
from the module's Msg and Query services. `--from` fills the signer field.

Trip and reset with the operator key:

```shell
simd tx safety trip --patterns '/cosmos.bank.v1beta1.MsgSend,/cosmos.staking.*' \
  --from operator --chain-id <chain-id> --node <rpc> --fees 2000stake
simd tx safety reset --patterns /cosmos.bank.v1beta1.MsgSend --patterns '/cosmos.staking.*' \
  --from operator --chain-id <chain-id> --node <rpc> --fees 2000stake
```

`--patterns` takes a comma-separated list, repeats, or both. Quote `*` and
prefix patterns so the shell does not expand them. The authority resets with
the same `reset` command from its own key. An authority key that signs directly
updates params with:

```shell
simd tx safety update-params \
  --params '{"authority":"cosmos1...","operator":"cosmos1...","scope":["/cosmos.bank.*"]}' \
  --from authority --chain-id <chain-id> --node <rpc> --fees 2000stake
```

A gov authority sends `MsgUpdateParams` or `MsgReset` in a proposal file for
`simd tx gov submit-proposal`.

Query:

```shell
simd q safety params --node <rpc>
simd q safety disabled-list --node <rpc>
simd q safety is-allowed --msg-type-url /cosmos.bank.v1beta1.MsgSend --node <rpc>
```

`is-allowed` omits `allowed` when it is false, since the output skips default
values.

### Without the generic commands

An app binary that does not build autocli this way still registers the msg
types, so the operator signs tx JSON with it and queries the REST gateway. Save
as `trip.json`:

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
```

A reset uses the same steps with `"@type": "/cosmos.safety.v1.MsgReset"` and
`"sender"` in place of `"operator"`. Give each pattern its own array entry.

```shell
curl <api>/cosmos/safety/v1/params
curl <api>/cosmos/safety/v1/disabled_list
curl '<api>/cosmos/safety/v1/is_allowed?msg_type_url=/cosmos.bank.v1beta1.MsgSend'
```

## IBC, EVM and wasm

### IBC relaying

A relayer delivers packets, acks, timeouts and client updates as ordinary
msgs, and the router checks each one. Effects of an exact trip on the tripped
chain:

| Trip | Effect |
| --- | --- |
| `/ibc.core.channel.v1.MsgRecvPacket` | Inbound packets wait. A packet that times out refunds on its source chain. A relayer can still deliver a packet after the reset and before its timeout. |
| `/ibc.core.channel.v1.MsgAcknowledgement` | Acks for outbound packets fail, so the chain refunds an error-acked packet only after the reset. A timeout cannot refund a packet the counterparty already received. |
| `/ibc.core.client.v1.MsgUpdateClient` | No light client on the chain updates, so recv and ack fail proof verification. Misbehavior submissions fail too, since ibc-go v11 takes misbehavior in `MsgUpdateClient`. |

Under `/ibc.core.channel.v1.*` the router refuses recv, ack and timeout but
not `/ibc.applications.transfer.v1.MsgTransfer`, which lives in another
package, so the chain keeps escrowing new sends that it can neither ack nor
time out until the reset. To stop ICS20 too, add the `MsgTransfer` URL or trip
`/ibc.*`.

End any trip that covers `MsgUpdateClient`, such as `/ibc.core.client.v1.*`,
`/ibc.*` or `*`, within the trusting period of the chain's clients. A client
that expires during the trip stays expired after the reset, and only gov
`MsgRecoverClient` revives it. x/safety does not protect `MsgRecoverClient`, so
the ante decorator refuses a gov proposal that wraps it, and a proposal that
reaches a block passes the vote and fails at execution. Reset the trip before
submitting the recovery proposal.

To stop transfers without freezing clients, trip
`/ibc.applications.transfer.v1.MsgTransfer` and `MsgRecvPacket` by exact URL,
and leave acks, timeouts and client updates open so packets already in flight
settle.

Under `*` every IBC msg fails, and gov resets the trip as usual. After the
reset, packets still inside their timeout relay and ack. The destination
refuses a packet whose timeout passed during the trip, and the source chain
refunds it on timeout.

### ICS20 transfers

The router refuses a routed `MsgTransfer` that a trip covers.
packet-forward-middleware calls the transfer keeper's `Transfer` msg server
method directly, so a forward still escrows and sends under a `MsgTransfer`
trip; see [Direct keeper calls](#direct-keeper-calls). The cosmos/evm ICS20
precompile calls the same method. Once it runs the msg with
`msgservice.Dispatch`, the router refuses the tripped msg and the precompile
call reverts with nothing escrowed; see [Direct callers](#direct-callers).

### Interchain account hosts

The ICA host routes the msgs in a packet through the router. When the router
refuses a tripped msg, the host fails the packet with an error ack,
`ABCI code: 10: error handling packet: see events for details`, and moves no
funds. The x/safety reason appears only in the host chain's events. The ante
decorator cannot see msgs inside packet data, so CheckTx admits the
`MsgRecvPacket`, which succeeds and pays its fee. Protected msgs still run:
under `/cosmos.*` a gov `MsgVote` sent through an interchain account executes.
Under `*` the host chain refuses the `MsgRecvPacket` itself; see
[Interchain accounts](#interchain-accounts).

### EVM chains

cosmos/evm sends every EVM tx as `/cosmos.evm.vm.v1.MsgEthereumTx`, and the
router checks it like any msg. Under an exact trip of it, or `*`, the router
refuses every EVM tx, precompile calls included, so during a `*` trip
validators vote with a Cosmos gov `MsgVote` tx, not the gov precompile.

cosmos/evm runs Ethereum txs through a separate ante chain. Without
`CircuitBreakerDecorator` in it, under a `MsgEthereumTx` or `*` trip the proposer
includes EVM txs in a block, where they fail with code 10, pay their fee, use
their nonce and return no receipt. Add the decorator to that chain; see
[Ante handler](#ante-handler).

With the decorator, the EVM mempool still answers `CheckTx` with code 0 and
`eth_sendRawTransaction` with a tx hash, because it queues the tx before the
ante chain runs. The decorator refuses the tx when the pool tries to promote
it, so the tx never becomes pending, never reaches a block, and uses no nonce
and pays no fee. The sender gets no error and no receipt. The tx stays queued and
becomes pending after the reset.

`eth_call` and `eth_estimateGas` route no `MsgEthereumTx`, so they never check
it. Under a `MsgEthereumTx` or `*` trip a value transfer estimates at 21000
gas, `eth_call` of it succeeds, and a gov precompile vote simulates as a
success, though neither tx executes. A precompile that runs a tripped msg with
`msgservice.Dispatch` fails in `eth_call` and `eth_estimateGas` too, because
the router checks the msg in the query context.

An EVM value transfer moves the native denom without a bank msg, so the router
never checks it against a `/cosmos.bank.v1beta1.MsgSend` trip. A precompile
that runs its SDK msg with `msgservice.Dispatch` reverts under a selective
trip of that msg. A precompile that calls a msg server directly, or calls a
keeper, skips the trip; see [Direct callers](#direct-callers) and
[Direct keeper calls](#direct-keeper-calls).

### CosmWasm contracts

wasmd turns `BankMsg::Send`, staking, gov, `CosmosMsg::Any`, Stargate and
`IbcMsg::Transfer` msgs into SDK msgs and routes them through the
`MsgServiceRouter`, so when the router refuses a tripped SDK msg, the contract
execution fails and moves no funds. CheckTx admits the outer
`MsgExecuteContract`, since the decorator cannot see the msgs a contract emits.
Gov msgs stay protected: under `/cosmos.*` a contract's gov vote executes and
its bank send fails.

Two wasmd paths call keepers directly and skip a selective trip:

* `BankMsg::Burn` calls the bank keeper. The bank module defines no `MsgBurn`,
  so `MsgTrip` of `/cosmos.bank.v1beta1.MsgBurn` fails with code 7.
* `IbcMsg::SendPacket` calls the IBC channel keeper and goes through under
  `/ibc.*`.

To stop those two, trip `MsgExecuteContract` by exact URL, `/cosmwasm.wasm.*`
or `*`; the router then refuses every contract execution.

A contract that sends a tripped msg as a submessage with `ReplyOn::Error`
receives the failure in its reply and continues. wasmd redacts the error to
`codespace: safety, code: 10`, so the contract sees neither the type URL nor
the pattern.

## Known gaps

### Direct keeper calls

The router check only sees msgs that pass through the `MsgServiceRouter`. Code
that calls a keeper or a msg server directly skips a selective trip; see
[Direct callers](#direct-callers). Known paths:

* cosmos/evm precompile methods that call SDK keepers, such as distribution
  `claimRewards` and WERC20 `deposit`,
* EVM native value transfers,
* the ibc-go transfer keeper's `Transfer` and `SendTransfer` methods, which
  packet-forward-middleware and other modules call,
* the ibc-go channel keeper's `SendPacket`, which any module or middleware can
  call to send a packet,
* wasmd `BankMsg::Burn` and `IbcMsg::SendPacket`. wasmd routes the other
  contract msgs; see [CosmWasm contracts](#cosmwasm-contracts).

A `*` trip still blocks every msg outside the protected and wrapper sets that
enters through a tx, directly or inside a wrapper, so it stops those paths at
the tx msg that starts them. A selective trip of, for example, `/cosmos.bank.*`
does not stop a contract or precompile from moving funds through the bank
keeper.

### Exact wrapper trips

An authority that can act only through a wrapper loses that path while the
operator keeps exact trips of the wrappers it uses. An account that signs
through an authz grantee loses it to a trip of `/cosmos.authz.v1beta1.MsgExec`.
A group policy resets through a group proposal, and group `MsgVote` and
`MsgExec` act only on proposals that already exist, so an exact trip of group
`MsgSubmitProposal` alone locks out a group policy with no reset proposal
pending. An exact trip of group `MsgVote` alone locks out a group whose
decision policy needs votes from members who are not proposers: group
`MsgSubmitProposal` with `exec` set to `EXEC_TRY` counts only the proposers'
votes, and the other members cannot vote. An exact trip of group `MsgExec`
alone leaves group `MsgVote` with `exec` set to `EXEC_TRY`, which executes the
proposal. The path returns when the operator resets those trips. Otherwise
recovery needs a chain-wide authority change, such as a gov proposal that sets
the consensus `AuthorityParams` override, or a software upgrade. Neither works
when the locked-out authority is the override itself: consensus
`MsgUpdateParams`, upgrade `MsgSoftwareUpgrade` and x/safety `MsgReset` and
`MsgUpdateParams` all check the override, so a gov proposal of any of them
fails. Until the operator resets the trip, recovery then needs validators to
halt and switch to a new binary off chain. Use gov or a directly signing
account as the authority, and as the override.

### Grantees under `*`

A `*` trip leaves authz `MsgExec` running, so a grantee can still run the
protected msgs and the wrappers the granter granted it, such as a gov deposit
from the granter's funds or a group vote in the granter's name, while `*`
blocks the granter's `MsgRevoke`. Trip `/cosmos.authz.v1beta1.MsgExec` by its
exact URL to stop it.

### Interchain accounts

An authority that acts through an interchain account needs `MsgRecvPacket` and
client updates, which a trip of `*` or `/ibc.*` blocks. Neither is a wrapper
or protected. Use gov, a directly signing account or a group policy as the
authority. A group policy is still exposed to
[exact wrapper trips](#exact-wrapper-trips), with no on-chain recovery when it
is the consensus override.

## Changes from x/circuit

| x/circuit | x/safety |
| --- | --- |
| An empty type URL list tripped or reset nothing and succeeded | Empty lists are rejected |
| Any string could be tripped | A pattern must match a registered, unprotected msg |
| `MsgResetCircuitBreaker` could be tripped, locking out recovery | Protected set, always allowed |
| Permission levels table and super admin delegation | One operator in params |
| Authority fixed at build time, compared with `bytes.Equal` | Authority in params, checked with `sdk.ValidateAuthority` |
| Revoked permissions left state that failed genesis validation | Genesis export always imports |
| Untyped events with comma-joined attributes | Typed proto events |
| Unregistered errors | Registered errors: codespace `safety`, or the shared SDK `ErrUnauthorized` and `ErrInvalidAddress` |
| Unpaginated disabled list query | Paginated |
| Ante decorator checked top-level msgs only | Recurses into nested msgs |
| CLI examples quoted several type URLs as one argument | Generic `--patterns` flag takes a comma-separated list or repeats |
| Exact type URLs only | Exact, prefix and `*` patterns |
