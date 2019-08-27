package vesting

/* ------------- README.md ------------------

Goal: Design more modular vesting accounts for the sdk, and use that design to build kava's specific vesting accounts.

Aim is to remove vesting logic from auth, bank, and staking modules, while making the vesting code modular enough to allow a wide variety of vesting accounts.

Turns out this is quite a hard.

## Reasons

**Account permissions problem**
The goal is to allow many types of vesting accounts, but you want the sdk to work with any of them. So a pattern is to create accounts as structs that fulfill a generic interface.
You also want permissions on what modules can use vesting coins to be set in the accounts structs (eg some vesting accounts could allow staking, but others do not).
However, if the sdk modules are using an account interface method for moving coins (ie all coin movements call bank.Send), then permissions cannot be controlled from within the account stucts.
An account can’t control who can send coins when it doesn’t know who called sendCoins.

consequences:
 * can’t have accounts control access to vesting coins AND have a generic interface to coin transfers that modules use (ie bank.SendCoins())
 * can’t have a generic function for coin transfers AND have staking still done in the same way as normal through the cli

**Vesting coins utilisation problem**
Ideally modules should be agnostic to whether they are using vesting coins or normal ones (ie there should be no code specific to vesting in staking, dist, etc).
However, in most modules, using a vesting coin in place of a normal (vested) coin will necessitate changes to the module’s logic.
Particularly for conditional vesting. ie what happens if a user doesn’t meet the conditions, so their coins don't vest, but they’ve already deposited them in a gov proposal?
Or when a user doesn't meet the conditions but has already staked all their vesting coins?
Using vesting coins to stake requires custom logic, this must be placed somewhere.

**Iterating accounts problem**
Iterating through all accounts is expensive. So doing it every block makes things slow.
So 'paying out' newly vested coins by updating state (like decrement a “vesting” var and increment a “vested” one) can’t be done every block.
Amount vested must be calculated when someone is trying to move the coins.

Can’t place the required logic for vesting accounts in an account struct that is accessed only through a generic sendCoins function.

## Design Proposal

Vesting needs specific code for each module that may use vesting coins.
Don’t want this in the actual modules. It can’t go in an account struct stored in auth (without expanding the account interface)
Therefore put in a vesting module that calls public methods on other modules (ie stake).

### Draft Implementation

Remove all vesting related logic from auth, bank, supply, stake. Put all vesting logic in it's own module.

Store vesting coins in the module and have users explicitly remove them after vest.

If users want to stake vesting coins they must do so by submitting messages to the vesting module, that then sort of routes them to the staking module.

This implementation contains 3 vesting account types: delayed and conditional from the current sdk, and conditional which is kava specific.

#### Implementation Notes

Improve UX by routing all staking commands through vesting for people who have vesting accounts,
ie allow them to delegate from both vesting and vested in one msg.
Then they don't need to interact with staking for their normal account.

Coin Storage:  
Want some kind of account abstraction in this module to group logic for schedules n stuff.
Make sense to store coins in those accouts.
Storing them as auth.Account's in auth make delegation easy, but allows coins to be sent to them.
Storing coins in module account makes delegation a bit more complex
Given that you need to store all (un/re)delegations if using the module account, probably makes sense to just use actual accounts in auth.

Slashing
wrong: instead of the debt collection mechanism the begin blocker could trigger a slash event (stakingKeeoer.Slash()) that instantly removes coins.
This would slash all the delegators rather than just the validator
TODO: what if an unbonding delegation is slashed? Should vesting unbond more to cover the debt or just accept it?
*/


// -------------- types/account.go ---------------

// Vesting account is an interface for internal use by the vesting module. These are not stored in the auth module.
type VestingAccount interface {
	SubtractFreeVestedCoins(time) sdk.Coins
}

// A Delayed vesting account releases coins all at once at a specific datetime
type DelayedVestingAccount struct {
	owner sdk.Address
	currentBalance sdk.Coins

	endTime
}
func (dva DelayedVestingAccount) SubtractFreeVestedCoins(time) {
	// based off current time calculate how many coins should now be vested
	// subtract all the free coins in the account up to the above amount
}

// A continuous vesting account releases coins continually between a start and end datetime.
type ContinuousVestingAccount struct {
	owner sdk.Address
	currentBalance sdk.Coins

	startTime
	endTime
	totalAmountToVest
}
func (cva ContinuousVestingAccount) SubtractFreeVestedCoins(time) {
	// based off current time calculate how many coins should now be vested
	// subtract all the free coins in the account up to the above amount
}

// A conditional vesting account releases coins at fixed intervals provided the associated validator has signed enough blocks in this interval
// As a consequence of this, if coins are currently delegated then they must be seized. This is done in the begin blocker.
type ConditionalVestingAccount struct {
	owner sdk.Address
	coins sdk.Coins

	startTime
	periodLength
	numPeriods
	totalAmountToVest

	validator sdk.ValOperAddress
	upTimePct sdk.Dec
	signingHistory []
	debt sdk.Coins
}

func (cva ConditionalVestingAccount) SubtractFreeVestedCoins(time) {
	// based off current time and signing history calculate how many coins should now be vested
	// subtract all the free coins in the account up to the above amount
}

// ----------- types/msg.go ---------------

// MsgWithdrawVested withdraws any newly vested to the owners normal account
type MsgWithdrawVested struct {
	Owner sdk.Address
}


// Define msgs for each msg from staking that you want to enable for vesting coins.
// Here DelegatorAddress is the address of the vesting account owner

type MsgDelegateVesting struct {
	*staking.MsgDelegate
}
// staking.MsgDelegate:
// type MsgDelegate struct {
// 	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
// 	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
// 	Amount           sdk.Coin       `json:"amount" yaml:"amount"`
// }

type MsgBeginRedelegateVesting struct {
	*staking.MsgBeginRedelegate
}
// staking.MsgBeginRedelegate:
// type MsgBeginRedelegate struct {
// 	DelegatorAddress    sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
// 	ValidatorSrcAddress sdk.ValAddress `json:"validator_src_address" yaml:"validator_src_address"`
// 	ValidatorDstAddress sdk.ValAddress `json:"validator_dst_address" yaml:"validator_dst_address"`
// 	Amount              sdk.Coin       `json:"amount" yaml:"amount"`
// }

type MsgUndelegateVesting struct {
	*staking.MsgUndelegate
}
// staking.MsgUndelegate:
// type MsgUndelegate struct {
// 	DelegatorAddress sdk.AccAddress `json:"delegator_address" yaml:"delegator_address"`
// 	ValidatorAddress sdk.ValAddress `json:"validator_address" yaml:"validator_address"`
// 	Amount           sdk.Coin       `json:"amount" yaml:"amount"`
// }


// ------------- handler.go ----------------


func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgWithdrawVested:
			// get vesting account ( keeper.GetVestingAccount(msg.Owner) )
			// coins := account.SubtractFreeVestedCoins()
			// supply.SendFromModuleToAccount(vestingModuleName, account.Owner, coins)

		case types.MsgDelegateVesting:
			// Do what staking handler does when delegating, mainly `stakingKeeper.Delegate`
			// Delegate from the vesting module account (staking needs to be modified to allow delegation from module accounts)
			// Record amounts delegated from this owner to different validators (need to know this to know where and how much to undelegate from)

		case types.MsgBeginRedelegateVesting:
			// same idea as above

		case types.MsgUndelegateVesting:
			// same idea as above

		default:
			errMsg := fmt.Sprintf("unrecognized vesting message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}


// ------------ abci.go --------------------

// This is not needed for non-conditional vesting accounts
func BeginBlocker(req abci.BeginBlockerRequest) {
	// For each conditional vesting account:

		// Record validator signatures into account structs
		// Follow how slashing module does it
		// This information is used to limit withdraws from accounts

		
		// If this block is the end of a period:
			// If num signed blocks is under threshold
				// Validator shouldn't get this period's coins, so mark coins for removal by incrementing account.Debt

		// Collect debt:
			// Remove coins (and debt) from vesting account (and vesting module account) if available
			// If there is still debt left:
				// Trigger undelegation of that amount. Once it unbonds this begin blocker will remove it.
}

// ------------- keeper.go --------------

type Keeper struct {
	// ...
}
func (k Keeper) GetVestingAccount(owner sdk.Address) vestingAccount {
	// fetch from db
}
func (k Keeper) SetVestingAccount(account vestingAccount) {
	// store in db
}
