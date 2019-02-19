# WIP 

this spec is WIP and simply an overview of the work to be done in the PR.
Here is the current thinking as to how the Invariant-Broken-Tx PR will be operate:

Within each module invariant registration functions are defined, for example: 

``` golang
func RegisterInvariants(ak auth.AccountKeeper, ck types.CrisisKeeper) {
	ck.RegisterInvariantRoute(ModuleName +"/nonnegative-balance",NonnegativeBalanceInvariant(ak))
	ck.RegisterInvariantRoute(ModuleName +"/some-other-invariant",SomeOtherInvariant(ak))
}
```

Within `cmd/gaia/app/app.go` function `NewGaiaApp` the crisis keeper would be
initialized with all of the invariants from other modules.  

XXX QUESTION:
because the invariants are registered in `NewGaiaApp` everything would get
reregistered in the case of a blockchain reinitialization correct? So these
routes should not be required to be stored in state. 

The crisis keeper would look something like this: 

``` golang
type InvarRoute struct {
	Route     string
	Invariant sdk.Invariant
}

type Keeper struct {
	routes      []InvarRoute
	distrKeeper DistrKeeper
}

//expected distribution keeper
type DistrKeeper interface {
	MustSpendFeePool(ctx sdk.Context,amt sdk.Coins, addr sdk.AccAddress) error
}
```

here the `MustSpendFeePool` is a new function to be created in the distribution
keeper which which moves coins from `CommunityPool` to an address. It is a "Must"
function because, if the community pool is empty, the community pool should
send as much as it has left to the address (and report an error). 

the crisis module will also hold a param in the param store which is the
`InvariancesFindersFee` which is the bonus fee which anyone who
finds an invariance will receive. 

### general discussion on rewarding the invariance finder

For simplicity I propose that we start with a simply a global param (modifiable
by governance) which is the bonus fee described above. This bonus should be
large enough to compensate for the gas cost for running the transaction, and
should be paid out of the `CommunityPool`.  The hope is that governance should
update this fee once in a while to account for the change in transaction cost
of the largest invariance check transaction. 

Future iterations of this finders fee could include a new a reward based on the
calculated gas cost _in addition_ to the static reward.  However for the time
being this seems as though it may be more complexity than necessary? Thoughts?
I'm assuming we would use `CalculateGas` from the `client/utils/utils.go`,
although I'm not totally positive as to how we would parameterize that function
from the handler, without restructuring baseapp.


### handler

Here is the message struct:
```
type MsgVerifyInvariance struct {
	Sender          sdk.AccAddress 
	InvarianceRoute string        
}
```

This message is handled in the following way:
 - iterate through keeper.routes to fine msg.InvarianceRoute then run the
   associated invariant
 - if an invariance is found credit the sender account with finders fee

