# ADR 058: Authz hooks

## Changelog

* 2022-08-02: Initial Draft

## Status

DRAFT Not Implemented

## Abstract

In order to extend the use cases covered by the `x/authz` module, we propose to implement a hook system so that external modules can perform stateful checks and/or state changes while handling a `MsgExec`. 

## Context

Currently, when handling `MsgExec` it is not possible to execute stateful checks before accepting an authorization, nor it is possible to perform state changes after it has been accepted. This makes it impossible to correctly implement some use cases like fee-based authorizations (as described inside [\#11583](https://github.com/cosmos/cosmos-sdk/issues/11583)).

## Decision

In order to improve the number of use cases that can be handled though the usage of `x/authz`, we proposed to implement a hook system within the authz keeper, similar to what other modules already use. 

Particularly, we have thought about the following hook structure: 

```go
type AuthzHooks interface {
    AfterAuthorizationAccepted(ctx sdk.Context, grantee sdk.AccAddress, authorization Authorization) error
}
```

This can then be used as follows within the `x/auth` keeper structure: 

```go
type Keeper struct {
    hooks authz.AuthzHooks
    // ...
}

// OnAuthorizationAccepted must be called after an authorization has been accepted
func (k Keeper) OnAuthorizationAccepted(ctx sdk.Context, authorization authz.Authorization) error {
    if k.hooks != nil {
        return k.hooks.OnAuthorizationAccepted(ctx, authorization)
    }
    return nil
}

// DispatchActions attempts to execute the provided messages via authorization
// grants from the message signer to the grantee.
func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) ([][]byte, error) {
    // [...]

        authorization, err := grant.GetAuthorization()
        if err != nil {
            return nil, err
        }
        
        resp, err := authorization.Accept(ctx, msg)
        if err != nil {
            return nil, err
        }
        
        if resp.Delete {
            err = k.DeleteGrant(ctx, grantee, granter, sdk.MsgTypeURL(msg))
        } else if resp.Updated != nil {
            err = k.update(ctx, grantee, granter, resp.Updated)
        }
        if err != nil {
            return nil, err
        }
        
        if !resp.Accept {
            return nil, sdkerrors.ErrUnauthorized
        }
        
        err = k.OnAuthorizationAccepted(ctx, grantee, authorization)
        if err != nil {
            return nil, err
        }
        
    // [...]
}
```

The above `AuthzHooks` interface can then be implemented by external modules to perform stateful checks and state changes if needed: 

```go
// FeeAuthorization represents an authorization that requires a fee to be executed
type FeeAuthorization struct {
    // Authorization to execute
    authz.Authorization
    
    // Amount of fees to be paid to execute the authorization
    FeeAmount sdk.Coins
    
    // Recipient that will receive the fees
    FeeRecipient string
}

// OnAuthorizationAccepted is implemented within the x/bank keeper
func (k *Keeper) OnAuthorizationAccepted(ctx sdk.Context, grantee sdk.AccAddress, authorization authz.Authorization) error {
    if feeAuthorization, ok := authorization.(*FeeAuthorization); ok {
        recipient, err := sdk.AccAddressFromBech32(feeAuthorization.FeeRecipient)
        if err != nil {
            return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, err)
        }
        
        err = k.SendCoins(ctx, grantee, recipient, feeAuthorization.FeeAmount)
        if err != nil {
            return err
        }
    }
}

```


## Consequences

### Backwards Compatibility

The existing authorization is not impacted, and everything can still work as expected.

### Positive

* enables a wider range of use cases to be supported

### Negative

{negative consequences}

### Neutral

* other modules might need to adapt by implementing the hooks if they want to handle some custom authorization types

## Further Discussions

The `FeeAuthorization` is just an example of authorization that could be implemented with this new flow and can be implemented along with this new flow.

## References

* https://github.com/cosmos/cosmos-sdk/issues/11583
