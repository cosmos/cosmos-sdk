# ADR 030: Authorization Module

## Changelog

- 2019-11-06: Initial Draft
- 2020-10-12: Updated Draft
- 2020-11-13: Accepted

## Status

Accepted

## Abstract

This ADR defines the `x/authz` module which allows accounts to grant authorizations to perform actions
on behalf of that account to other accounts.

## Context

The concrete use cases which motivated this module include:
- the desire to delegate the ability to vote on proposals to other accounts besides the account which one has
delegated stake
- "sub-keys" functionality, as originally proposed in [\#4480](https://github.com/cosmos/cosmos-sdk/issues/4480) which
is a term used to describe the functionality provided by this module together with
the `fee_grant` module from [ADR 029](./adr-029-fee-grant-module.md) and the [group module](https://github.com/regen-network/cosmos-modules/tree/master/incubator/group).

The "sub-keys" functionality roughly refers to the ability for one account to grant some subset of its capabilities to
other accounts with possibly less robust, but easier to use security measures. For instance, a master account representing
an organization could grant the ability to spend small amounts of the organization's funds to individual employee accounts.
Or an individual (or group) with a multisig wallet could grant the ability to vote on proposals to any one of the member
keys.

The current
implementation is based on work done by the [Gaian's team at Hackatom Berlin 2019](https://github.com/cosmos-gaians/cosmos-sdk/tree/hackatom/x/delegation).

## Decision

We will create a module named `authz` which provides functionality for
granting arbitrary privileges from one account (the _granter_) to another account (the _grantee_). Authorizations
must be granted for a particular `Msg` service methods one by one using an implementation
of `Authorization`.

### Types

Authorizations determine exactly what privileges are granted. They are extensible
and can be defined for any `Msg` service method even outside of the module where
the `Msg` method is defined. `Authorization`s use the new `ServiceMsg` type from
ADR 031.

#### Authorization

```go
type Authorization interface {
	// MethodName returns the fully-qualified Msg service method name as described in ADR 031.
	MethodName() string

	// Accept determines whether this grant permits the provided sdk.ServiceMsg to be performed, and if
	// so provides an upgraded authorization instance.
	// Returns:
	// + allow: true if msg is authorized
	// + updated: new Authorization instance which should overwrite the current one with new state
	// + delete: true if Authorization has been exhausted and can be deleted from state
	Accept(msg sdk.ServiceMsg, block abci.Header) (allow bool, updated Authorization, delete bool)
}
```

For example a `SendAuthorization` like this is defined for `MsgSend` that takes
a `SpendLimit` and updates it down to zero:

```go
type SendAuthorization struct {
	// SpendLimit specifies the maximum amount of tokens that can be spent
	// by this authorization and will be updated as tokens are spent. If it is
	// empty, there is no spend limit and any amount of coins can be spent.
	SpendLimit sdk.Coins
}

func (cap SendAuthorization) MethodName() string {
	return "/cosmos.bank.v1beta1.Msg/Send"
}

func (cap SendAuthorization) Accept(msg sdk.ServiceMsg, block abci.Header) (allow bool, updated Authorization, delete bool) {
	switch req := msg.Request.(type) {
	case bank.MsgSend:
		left, invalid := cap.SpendLimit.SafeSub(req.Amount)
		if invalid {
			return false, nil, false
		}
		if left.IsZero() {
			return true, nil, true
		}
		return true, SendAuthorization{SpendLimit: left}, false
	}
	return false, nil, false
}
```

A different type of capability for `MsgSend` could be implemented
using the `Authorization` interface with no need to change the underlying
`bank` module.

### `Msg` Service

```proto
service Msg {
  // GrantAuthorization grants the provided authorization to the grantee on the granter's
  // account with the provided expiration time.
  rpc GrantAuthorization(MsgGrantAuthorization) returns (MsgGrantAuthorizationResponse);

  // ExecAuthorized attempts to execute the provided messages using
  // authorizations granted to the grantee. Each message should have only
  // one signer corresponding to the granter of the authorization.
  // The grantee signing this message must have an authorization from the granter.
  rpc ExecAuthorized(MsgExecAuthorized) returns (MsgExecAuthorizedResponse)


  // RevokeAuthorization revokes any authorization corresponding to the provided method name on the
  // granter's account that has been granted to the grantee.
  rpc RevokeAuthorization(MsgRevokeAuthorization) returns (MsgRevokeAuthorizationResponse);
}

message MsgGrantAuthorization{
  string granter = 1;
  string grantee = 2;
  google.protobuf.Any authorization = 3 [(cosmos_proto.accepts_interface) = "Authorization"];
  google.protobuf.Timestamp expiration = 4;
}

message MsgExecAuthorized {
    string grantee = 1;
    repeated google.protobuf.Any msgs = 2;
}

message MsgRevokeAuthorization{
  string granter = 1;
  string grantee = 2;
  string method_name = 3;
}
```

### Router Middleware

The `authz` `Keeper` will expose a `DispatchActions` method which allows other modules to send `ServiceMsg`s
to the router based on `Authorization` grants:

```go
type Keeper interface {
	// DispatchActions routes the provided msgs to their respective handlers if the grantee was granted an authorization
	// to send those messages by the first (and only) signer of each msg.
    DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.ServiceMsg) sdk.Result`
}
```

This allows the functionality provided by `authz` to be used for future inter-module object capabilities
permissions as described in [ADR 033](https://github.com/cosmos/cosmos-sdk/7459)

### CLI

#### `tx exec` Method

When a CLI user wants to run a transaction on behalf of another account using `MsgExecAuthorized`, they
can use the `exec` method. For instance `gaiacli tx gov vote 1 yes --from <grantee> --generate-only | gaiacli tx authz exec --send-as <granter> --from <grantee>`
would send a transaction like this:

```go
MsgExecAuthorized {
  Grantee: mykey,
  Msgs: []sdk.SericeMsg{
    ServiceMsg {
      MethodName:"/cosmos.gov.v1beta1.Msg/Vote"
      Request: MsgVote {
	    ProposalID: 1,
	    Voter: cosmos3thsdgh983egh823
	    Option: Yes
      }
    }
  }
}
```

#### `tx grant <grantee> <authorization> --from <granter>`

This CLI command will send a `MsgGrantAuthorization` transaction. `authorization` should be encoded as
JSON on the CLI.

#### `tx revoke <grantee> <method-name> --from <granter>`

This CLI command will send a `MsgRevokeAuthorization` transaction.

### Built-in Authorizations

#### `SendAuthorization`

```proto
// SendAuthorization allows the grantee to spend up to spend_limit coins from
// the granter's account.
message SendAuthorization {
  repeated cosmos.base.v1beta1.Coin spend_limit = 1;
}
```

#### `GenericAuthorization`

```proto
// GenericAuthorization gives the grantee unrestricted permissions to execute
// the provide method on behalf of the granter's account.
message GenericAuthorization {
  string method_name = 1;
}
```

## Consequences

### Positive

- Users will be able to authorize arbitrary actions on behalf of their accounts to other
users, improving key management for many use cases
- The solution is more generic than previously considered approaches and the
`Authorization` interface approach can be extended to cover other use cases by
SDK users

### Negative

### Neutral

## References

- Initial Hackatom implementation: https://github.com/cosmos-gaians/cosmos-sdk/tree/hackatom/x/delegation
- Post-Hackatom spec: https://gist.github.com/aaronc/b60628017352df5983791cad30babe56#delegation-module
- B-Harvest subkeys spec: https://github.com/cosmos/cosmos-sdk/issues/4480
