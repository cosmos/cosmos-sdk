/*
Package fee_grant provides functionality for delegating sub-permissions
from one account (key) to another account (key).

The first implementation allows the delegation for the payment of transaction fees.
Effectively, this allows for a user to pay fees using the balance of an account
different from their own. Example use cases would be allowing a key on a device to
pay for fees using a master wallet, or a third party service allowing users to
pay for transactions without ever really holding their own tokens. This package
provides ways for specifying fee allowances such that delegating fees
to another account can be done with clear and safe restrictions.

A user would delegate fees to a user using MsgDelegateFeeAllowance and revoke
that delegation using MsgRevokeFeeAllowance. In both cases Granter is the one
who is delegating fees and Grantee is the one who is receiving the delegation.
So grantee would correspond to the one who is signing a transaction and the
granter would be the address they place in DelegatedFee.FeeAccount.

The fee allowance that a grantee receives is specified by an implementation of
the FeeAllowance interface. Two FeeAllowance implementations are provided in
this package: BasicFeeAllowance and PeriodicFeeAllowance.

In order to integrate this into an application, we must use the DeductDelegatedFeeDecorator
ante handler from this package instead of the default DeductFeeDecorator from auth.
To allow handling txs from empty accounts (with fees paid from an existing account),
we have to re-order the decorators as well. You can see an example in
`x/delegate_fees/internal/ante/fee_test.go:newAnteHandler()`
*/
package fee_grant
