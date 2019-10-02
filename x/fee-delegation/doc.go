/*
package fee_delegation provides functionality for delegating the payment of
transaction fees from one account to another account. Effectively, this
allows for a user to pay fees using the balance of an account different from
their own. Example use cases would be allowing a key on a device to pay for fees
using a master wallet, or a third party service allowing users to pay for
transactions without ever really holding their own tokens. This package
provides ways for specifying fee allowances such that delegating fees
to another account can be done with clear and safe restrictions.

In order to integrate this into an application, the "ante handler" which deducts
fees must call Keeper.AllowDelegatedFees to check if
the provided StdTx.Fee can be delegated from the Std.TxFeeAccount address
to the first signer of the transaction. An example usage would be:

allow := feeDelegationKeeper.AllowDelegatedFees(ctx, signers[0], stdTx.FeeAccount, stdTx.Fee.Amount)

A user would delegate fees to a user using MsgDelegateFeeAllowance and revoke
that delegation using MsgRevokeFeeAllowance. In both cases Granter is the one
who is delegating fees and Grantee is the one who is receiving the delegation.
So grantee would correspond to the one who is signing a transaction and the
granter would be the address they place in StdTx.FeeAccount.

The fee allowance that a grantee receives is specified by an implementation of
the FeeAllowance interface. Two FeeAllowance implementations are provided in
this package: BasicFeeAllowance and PeriodicFeeAllowance.
 */
package fee_delegation
