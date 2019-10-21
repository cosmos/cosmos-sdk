# `msg_delegation` Module

The `msg_delegation` module provides support for granting arbitrary capabilities
from one account (the granter) to another account (the grantee). Capabilities
must be granted for a particular type of `sdk.Msg` one by one using an implementation
of `Capability`.

## Types

Capabilities determine exactly what action is delegated. They are extensible
and can be defined for any sdk.Msg type even outside of the module where the Msg is defined.

### Capability

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
using the `Capability` interface with new need to change the underlying
`bank` module.

## Messages

### `MsgGrant`

```go
// MsgGrant grants the provided capability to the grantee on the granter's
// account with the provided expiration time
type MsgGrant struct {
	Granter    sdk.AccAddress `json:"granter"`
	Grantee    sdk.AccAddress `json:"grantee"`
	Capability Capability     `json:"capability"`
    // Expiration specifies the expiration time of the grant
	Expiration time.Time      `json:"expiration"`
}
```

### MsgRevoke

```go
// MsgRevoke revokes whatever any capability on the granter's account with
// the provided msg type that has been granted to the grantee
type MsgRevoke struct {
	Granter sdk.AccAddress `json:"granter"`
	Grantee sdk.AccAddress `json:"grantee"`
	MsgType sdk.Msg        `json:"msg_type"`
}

```

### MsgExecDelegated

```go
// MsgExecDelegated attempts to execute the provided messages using
// capabilities granted to the signer. Each message should have only
// one signer corresponding to the granter of the capability.
type MsgExecDelegated struct {
	Grantee sdk.AccAddress `json:"grantee"`
	Msgs   []sdk.Msg      `json:"msg"`
}
```

## Keeper

The message delegation keeper receives a reference to the `BaseApp` `Router`

### `DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) sdk.Result`

`DispatchActions attempts to execute the provided messages via capability
grants from the message signer to the grantee..