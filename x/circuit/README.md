# Circuit Breaker - block specific functionality in the state machine


## Concepts

Circuit Breaker is a module that is meant to avoid a chain needing to halt/shut down in the presence of a vulnerability, instead the module will allow specific messages or all messages to be disabled. When operating a chain, if it is app specific then a halt of the chain is less detrimental, but if there are applications built on top of the chain then halting is expensive due to the disturbance to applications. 

Circuit Breaker works with the idea that an address or set of addresses have the right to block messages from being executed and/or included in the mempool. Any address with a permission is able to reset the circuit breaker for the message. 

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

Trip, is called by an account to disable message execution for a specific msgURL. 

```protobuf
  // TripCircuitBreaker pauses processing of Msg's in the state machine.
  rpc TripCircuitBreaker(MsgTripCircuitBreaker) returns (MsgTripCircuitBreakerResponse);
```

### Reset

Reset is called to enable execution of a previously disabled message. 

```protobuf
  // ResetCircuitBreaker resumes processing of Msg's in the state machine that
  // have been been paused using TripCircuitBreaker.
  rpc ResetCircuitBreaker(MsgResetCircuitBreaker) returns (MsgResetCircuitBreakerResponse);
```

## Messages

### MsgAuthorizeCircuitBreaker

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/circuit/v1/tx.proto#L25-L75
```
This message is expected to fail if:

* the granter is not an account with permission level `LEVEL_SUPER_ADMIN` or the module authority
* if the type urls does not exist <!-- TODO: is this possible?-->

### MsgTripCircuitBreaker

```protobuf reference 
https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/circuit/v1/tx.proto#L77-L93
```

This message is expected to fail if:

* if the signer does not have a permission level with the ability to disable the specified type url message
* if the type urls does not exist <!-- TODO: is this possible?-->

### MsgResetCircuitBreaker

```protobuf reference
https://github.com/cosmos/cosmos-sdk/blob/main/proto/cosmos/circuit/v1/tx.proto#L95-109
```

This message is expected to fail if:

* if the type urls does not exist <!-- TODO: is this possible?-->
* if the type url is not disabled

* `## Events` - list and describe event tags used
* `## Client` - list and describe CLI commands and gRPC and REST endpoints
