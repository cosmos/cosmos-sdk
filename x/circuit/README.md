# `x/circuit`

## Concepts

Circuit Breaker is a module that is meant to avoid a chain needing to halt/shut down in the presence of a vulnerability, instead the module will allow specific messages or all messages to be disabled. When operating a chain, if it is app specific then a halt of the chain is less detrimental, but if there are applications built on top of the chain then halting is expensive due to the disturbance to applications. 

## How it works

Circuit Breaker works with the idea that an address or set of addresses have the right to block messages from being executed and/or included in the mempool. Any address with a permission is able to reset the circuit breaker for the message. 

The transactions are checked and can be rejected at two points:

* In `CircuitBreakerDecorator` [ante handler](https://docs.cosmos.network/main/learn/advanced/baseapp#antehandler):

```go reference
https://github.com/cosmos/cosmos-sdk/blob/x/circuit/v0.1.0/x/circuit/ante/circuit.go#L27-L41
``` 

* With a [message router check](https://docs.cosmos.network/main/learn/advanced/baseapp#msg-service-router):

```go reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/baseapp/msg_service_router.go#L123-L133
``` 

:::note
The `CircuitBreakerDecorator` works for most use cases, but [does not check the inner messages of a transaction](https://docs.cosmos.network/main/learn/beginner/tx-lifecycle#antehandler). This some transactions (such as `x/authz` transactions or some `x/gov` transactions) may pass the ante handler. **This does not affect the circuit breaker** as the message router check will still fail the transaction.
This tradeoff is to avoid introducing more dependencies in the `x/circuit` module. Chains can re-define the `CircuitBreakerDecorator` to check for inner messages if they wish to do so.
:::

## State

### Accounts

* AccountPermissions `0x1 | account_address  -> ProtocolBuffer(CircuitBreakerPermissions)`

```go
type level int32

const (
    // LEVEL_NONE_UNSPECIFIED indicates that the account will have no circuit
    // breaker permissions.
    LEVEL_NONE_UNSPECIFIED = iota
    // LEVEL_SOME_MSGS indicates that the account will have permission to
    // trip or reset the circuit breaker for some Msg type URLs. If this level
    // is chosen, a non-empty list of Msg type URLs must be provided in
    // limit_type_urls.
    LEVEL_SOME_MSGS
    // LEVEL_ALL_MSGS indicates that the account can trip or reset the circuit
    // breaker for Msg's of all type URLs.
    LEVEL_ALL_MSGS 
    // LEVEL_SUPER_ADMIN indicates that the account can take all circuit breaker
    // actions and can grant permissions to other accounts.
    LEVEL_SUPER_ADMIN
)

type Access struct {
	level int32 
	msgs []string // if full permission, msgs can be empty
}
```

### Disable List

List of type urls that are disabled.

* DisableList `0x2 | msg_type_url -> []byte{}` <!--- should this be stored in json to skip encoding and decoding each block, does it matter?-->

## State Transitions

### Authorize 

Authorize, is called by the module authority (default governance module account) or any account with `LEVEL_SUPER_ADMIN` to give permission to disable/enable messages to another account. There are three levels of permissions that can be granted. `LEVEL_SOME_MSGS` limits the number of messages that can be disabled. `LEVEL_ALL_MSGS` permits all messages to be disabled. `LEVEL_SUPER_ADMIN` allows an account to take all circuit breaker actions including authorizing and deauthorizing other accounts.

```protobuf
  // AuthorizeCircuitBreaker allows a super-admin to grant (or revoke) another
  // account's circuit breaker permissions.
  rpc AuthorizeCircuitBreaker(MsgAuthorizeCircuitBreaker) returns (MsgAuthorizeCircuitBreakerResponse);
```

### Trip

Trip, is called by an authorized account to disable message execution for a specific msgURL. If empty, depending on the permission level of the sender, the corresponding messages will be disabled. For example: if the sender permission level is `LEVEL_SOME_MSGS` then all messages that sender has permission will be disabled. If the sender is `LEVEL_SUPER_ADMIN` or `LEVEL_ALL_MSGS` then all msgs will be disabled.

```protobuf
  // TripCircuitBreaker pauses processing of Msg's in the state machine.
  rpc TripCircuitBreaker(MsgTripCircuitBreaker) returns (MsgTripCircuitBreakerResponse);
```

### Reset

Reset is called by an authorized account to enable execution for a specific msgURL of previously disabled message. If empty, depending on the permission level of the sender, the corresponding disabled messages will be re-enabled. For example: if the sender permission level is `LEVEL_SOME_MSGS` all messages that sender has permission will be re-enabled. If the sender is `LEVEL_SUPER_ADMIN` or `LEVEL_ALL_MSGS` then all messages will be re-enabled.

```protobuf
  // ResetCircuitBreaker resumes processing of Msg's in the state machine that
  // have been paused using TripCircuitBreaker.
  rpc ResetCircuitBreaker(MsgResetCircuitBreaker) returns (MsgResetCircuitBreakerResponse);
```

## Messages

### MsgAuthorizeCircuitBreaker

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/x/circuit/proto/cosmos/circuit/v1/tx.proto#L25-L40
```

This message is expected to fail if:

* the granter is not an account with permission level `LEVEL_SUPER_ADMIN` or the module authority

### MsgTripCircuitBreaker

```protobuf reference 
https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/x/circuit/proto/cosmos/circuit/v1/tx.proto#L47-L60
```

This message is expected to fail if:

* if the signer does not have a permission level with the ability to disable the specified type url message

### MsgResetCircuitBreaker

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/release/v0.52.x/x/circuit/proto/cosmos/circuit/v1/tx.proto#L67-L78
```

This message is expected to fail if:

* if the type url is not disabled

## Events

The circuit module emits the following events:

### Message Events

#### MsgAuthorizeCircuitBreaker

| Type    | Attribute Key | Attribute Value           |
|---------|---------------|---------------------------|
| string  | granter       | {granterAddress}          |
| string  | grantee       | {granteeAddress}          |
| string  | permission    | {granteePermissions}      |
| message | module        | circuit                   |
| message | action        | authorize_circuit_breaker |

#### MsgTripCircuitBreaker

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| string   | authority     | {authorityAddress} |
| []string | msg_urls      | []string{msg_urls} |
| message  | module        | circuit            |
| message  | action        | trip_circuit_breaker |

#### ResetCircuitBreaker

| Type     | Attribute Key | Attribute Value    |
|----------|---------------|--------------------|
| string   | authority     | {authorityAddress} |
| []string | msg_urls      | []string{msg_urls} |
| message  | module        | circuit            |
| message  | action        | reset_circuit_breaker |


## Keys

* `AccountPermissionPrefix` - `0x01`
* `DisableListPrefix` -  `0x02`

## Client

### CLI

`x/circuit` module client provides the following CLI commands:

```shell
$ <appd> tx circuit
Transactions commands for the circuit module

Usage:
  simd tx circuit [flags]
  simd tx circuit [command]

Available Commands:
  authorize   Authorize an account to trip the circuit breaker.
  disable     Disable a message from being executed
  reset       Enable a message to be executed
```

```shell
$ <appd> query circuit
Querying commands for the circuit module

Usage:
  simd query circuit [flags]
  simd query circuit [command]

Available Commands:
  account       Query a specific account's permissions
  accounts      Query all account permissions
  disabled-list Query a list of all disabled message types
```
