# ADR 030: Msg Authorization Module

## Changelog

- 2019-11-06: Initial Draft
- 2020/08/18: Updated Draft

## Status

Accepted

## Context

At Hackatom Berlin in June 2019, the Gaians team worked on a `delegation` module that is described in the spec below.
Prior to that, the B-Harvest team had worked on a [subkeys spec](https://github.com/cosmos/cosmos-sdk/issues/4480) that
covered similar use cases. In discussions after the Hackathon, the `delegation` module approach was deemed to
be more generic and there was community desire for the work to continue. This resulted in an ICF grant to continue
this work along with the fee grants and key groups modules.

## Decision

We will create a module named `msg_authorization`. The `msg_authorization` module provides support for granting arbitrary capabilities
from one account (the granter) to another account (the grantee). Capabilities
must be granted for a particular type of `sdk.Msg` one by one using an implementation
of `Capability`.

### Types

Capabilities determine exactly what action is delegated. They are extensible
and can be defined for any sdk.Msg type even outside of the module where the Msg is defined.

#### Capability

```go
type Capability interface {
	// MsgType returns the type of Msg's that this capability can accept
	MsgType() sdk.Msg
	// Accept determines whether this grant allows the provided action, and if
	// so provides an upgraded capability grant
	Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Capability, delete bool)
}
```

For example a `SendCapability` like this is defined for `MsgSend` that takes
a `SpendLimit` and updates it down to zero:

```go
type SendCapability struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this capability and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}

func (cap SendCapability) MsgType() sdk.Msg {
	return bank.MsgSend{}
}

func (cap SendCapability) Accept(msg sdk.Msg, block abci.Header) (allow bool, updated Capability, delete bool) {
	switch msg := msg.(type) {
	case bank.MsgSend:
		left, invalid := cap.SpendLimit.SafeSub(msg.Amount)
		if invalid {
			return false, nil, false
		}
		if left.IsZero() {
			return true, nil, true
		}
		return true, SendCapability{SpendLimit: left}, false
	}
	return false, nil, false
}
```

A different type of capability for `MsgSend` could be implemented
using the `Capability` interface with no need to change the underlying
`bank` module.

### Messages

#### `MsgGrant`

```go
// MsgGrant grants the provided capability to the grantee on the granter's
// account with the provided expiration time.
type MsgGrant struct {
	Granter    sdk.AccAddress `json:"granter"`
	Grantee    sdk.AccAddress `json:"grantee"`
	Capability Capability     `json:"capability"`
    // Expiration specifies the expiration time of the grant
	Expiration time.Time      `json:"expiration"`
}
```

#### MsgRevoke

```go
// MsgRevoke revokes any capability with the provided sdk.Msg type on the
// granter's account with that has been granted to the grantee.
type MsgRevoke struct {
	Granter sdk.AccAddress `json:"granter"`
	Grantee sdk.AccAddress `json:"grantee"`
    // CapabilityMsgType is the type of sdk.Msg that the revoked Capability refers to.
    // i.e. this is what `Capability.MsgType()` returns
	CapabilityMsgType sdk.Msg `json:"capability_msg_type"`
}
```

#### MsgExecDelegated

```go
// MsgExecDelegated attempts to execute the provided messages using
// capabilities granted to the grantee. Each message should have only
// one signer corresponding to the granter of the capability.
type MsgExecDelegated struct {
	Grantee sdk.AccAddress `json:"grantee"`
	Msgs   []sdk.Msg      `json:"msg"`
}
```

### Keeper

#### Constructor: `NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router sdk.Router) Keeper`

The message delegation keeper receives a reference to the baseapp's `Router` in order
to dispatch delegated messages.

#### `DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) sdk.Result`

`DispatchActions` attempts to execute the provided messages via capability
grants from the message signer to the grantee.

#### `Grant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, capability Capability, expiration time.Time)`

Grants the provided capability to the grantee on the granter's account with the provided expiration
time. If there is an existing capability grant for the same `sdk.Msg` type, this grant
overwrites that.

#### `Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg)`

Revokes any capability for the provided message type granted to the grantee by the granter.

#### `GetCapability(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg) (cap Capability, expiration time.Time)`

Returns any `Capability` (or `nil`), with the expiration time, granted to the grantee by the granter for the provided msg type.

### CLI

#### `--send-as` Flag

When a CLI user wants to run a transaction as another user using `MsgExecDelegated`, they
can use the `--send-as` flag. For instance `gaiacli tx gov vote 1 yes --from mykey --send-as cosmos3thsdgh983egh823`
would send a transaction like this:

```go
MsgExecDelegated {
  Grantee: mykey,
  Msgs: []sdk.Msg{
    MsgVote {
	  ProposalID: 1,
	  Voter: cosmos3thsdgh983egh823
	  Option: Yes
    }
  }
}
```
#### `tx grant <grantee> <capability> --from <granter>`

This CLI command will send a `MsgGrant` tx. `capability` should be encoded as
JSON on the CLI.

#### `tx revoke <grantee> <capability-msg-type> --from <granter>`

This CLI command will send a `MsgRevoke` tx. `capability-msg-type` should be encoded as
JSON on the CLI.

### Built-in Capabilities

#### `SendCapability`

```go
type SendCapability struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this capability and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}
```

#### `GenericCapability`

```go
// GenericCapability grants the permission to execute any transaction of the provided
// sdk.Msg type without restrictions
type GenericCapability struct {
    // MsgType is the type of Msg this capability grant allows
    MsgType sdk.Msg
}
```

## Status

Accepted

## Consequences

### Positive

- Users will be able to authorize arbitrary permissions on their accounts to other
users, simplifying key management for some use cases
- The solution is more generic than previously considered approaches and the
`Capability` interface approach can be extended to cover other use cases by 
SDK users

### Negative

### Neutral

## References

- Initial Hackatom implementation: https://github.com/cosmos-gaians/cosmos-sdk/tree/hackatom/x/delegation
- Post-Hackatom spec: https://gist.github.com/aaronc/b60628017352df5983791cad30babe56#delegation-module
- B-Harvest subkeys spec: https://github.com/cosmos/cosmos-sdk/issues/4480
