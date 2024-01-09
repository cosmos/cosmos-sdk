package vesting

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	"cosmossdk.io/x/accounts/accountstd"
	impl "cosmossdk.io/x/accounts/internal/implementation"
	vestingtypes "cosmossdk.io/x/accounts/vesting/types/v1"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Base Vesting Account
var _ accountstd.Interface = (*BaseVestingAccount)(nil)

type getVestingFunc = func(time time.Time) sdk.Coins

// NewBaseVestingAccount creates a new BaseVestingAccount object.
func NewBaseVestingAccount(d accountstd.Dependencies) (*BaseVestingAccount, error) {
	baseVestingAccount := &BaseVestingAccount{
		OriginalVesting:  sdk.NewCoins(),
		DelegatedFree:    sdk.NewCoins(),
		DelegatedVesting: sdk.NewCoins(),
		AddressCodec:     d.AddressCodec,
		EndTime:          0,
	}

	return baseVestingAccount, nil
}

type BaseVestingAccount struct {
	OriginalVesting  sdk.Coins
	DelegatedFree    sdk.Coins
	DelegatedVesting sdk.Coins
	AddressCodec     address.Codec
	// Vesting end time, as unix timestamp (in seconds).
	EndTime int64
}

type StateTransitionRecord struct {
	DelegatedFree    sdk.Coins
	DelegatedVesting sdk.Coins
}

type StateTransitionRecords struct {
	Records []StateTransitionRecord
}

// NewStateTransitionRecords creates a new StateTransitionRecords object.
func (bva BaseVestingAccount) NewInitialStateTransitionRecords() StateTransitionRecords {
	return StateTransitionRecords{
		Records: []StateTransitionRecord{{
			DelegatedFree:    bva.DelegatedFree,
			DelegatedVesting: bva.DelegatedVesting,
		}},
	}
}

// --------------- execute -----------------

// TrackDelegation tracks a delegation amount for any given vesting account type
// given the amount of coins currently vesting and the current account balance
// of the delegation denominations.
//
// CONTRACT: The account's coins, delegation coins, vesting coins, and delegated
// vesting coins must be sorted.
func (bva *BaseVestingAccount) TrackDelegation(
	balance, vestingCoins, amount, delegatedFree, delegatedVesting sdk.Coins,
) (newDelegatedFree, newDelegatedVesting sdk.Coins) {
	for _, coin := range amount {
		baseAmt := balance.AmountOf(coin.Denom)
		vestingAmt := vestingCoins.AmountOf(coin.Denom)
		delVestingAmt := delegatedVesting.AmountOf(coin.Denom)

		// Panic if the delegation amount is zero or if the base coins does not
		// exceed the desired delegation amount.
		if coin.Amount.IsZero() || baseAmt.LT(coin.Amount) {
			panic("delegation attempt with zero coins or insufficient funds")
		}

		// compute x and y per the specification, where:
		// X := min(max(V - DV, 0), D)
		// Y := D - X
		x := math.MinInt(math.MaxInt(vestingAmt.Sub(delVestingAmt), math.ZeroInt()), coin.Amount)
		y := coin.Amount.Sub(x)

		newDelegatedFree = delegatedFree
		newDelegatedVesting = delegatedVesting
		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			newDelegatedVesting = delegatedVesting.Add(xCoin)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			newDelegatedFree = delegatedFree.Add(yCoin)
		}
	}

	return newDelegatedFree, newDelegatedFree
}

// TrackUndelegation tracks an undelegation amount by setting the necessary
// values by which delegated vesting and delegated vesting need to decrease and
// by which amount the base coins need to increase.
//
// NOTE: The undelegation (bond refund) amount may exceed the delegated
// vesting (bond) amount due to the way undelegation truncates the bond refund,
// which can increase the validator's exchange rate (tokens/shares) slightly if
// the undelegated tokens are non-integral.
//
// CONTRACT: The account's coins and undelegation coins must be sorted.
func (bva *BaseVestingAccount) TrackUndelegation(
	amount, delegatedFree, delegatedVesting sdk.Coins,
) (newDelegatedFree, newDelegatedVesting sdk.Coins) {
	for _, coin := range amount {
		// panic if the undelegation amount is zero
		if coin.Amount.IsZero() {
			panic("undelegation attempt with zero coins")
		}
		delegatedFreeAmount := delegatedFree.AmountOf(coin.Denom)
		delegatedVestingAmount := delegatedVesting.AmountOf(coin.Denom)

		// compute x and y per the specification, where:
		// X := min(DF, D)
		// Y := min(DV, D - X)
		x := math.MinInt(delegatedFreeAmount, coin.Amount)
		y := math.MinInt(delegatedVestingAmount, coin.Amount.Sub(x))

		newDelegatedFree = delegatedFree
		newDelegatedVesting = delegatedVesting
		if !x.IsZero() {
			xCoin := sdk.NewCoin(coin.Denom, x)
			newDelegatedFree = delegatedFree.Sub(xCoin)
		}

		if !y.IsZero() {
			yCoin := sdk.NewCoin(coin.Denom, y)
			newDelegatedVesting = delegatedVesting.Sub(yCoin)
		}
	}

	return newDelegatedFree, newDelegatedVesting
}

// ExecuteMessages handle the execution of codectypes Any messages
// and update the vesting account DelegatedFree and DelegatedVesting
// when delegate or undelegate is trigger.
func (bva *BaseVestingAccount) ExecuteMessages(
	ctx context.Context, msg *vestingtypes.MsgExecuteMessages, getVestingFunc getVestingFunc,
) (
	*vestingtypes.MsgExecuteMessagesResponse, error,
) {
	originalContext := accountstd.OriginalContext(ctx)
	sdkctx := sdk.UnwrapSDKContext(originalContext)

	// Keep track of all delegation related transitions
	stateRecords := bva.NewInitialStateTransitionRecords()
	for _, m := range msg.ExecutionMessages {
		protoMsg, err := impl.UnpackAnyRaw(m)
		if err != nil {
			return nil, err
		}

		previousStateRecord := stateRecords.Records[len(stateRecords.Records)-1]

		typeUrl := codectypes.MsgTypeURL(protoMsg)
		switch typeUrl {
		case "/cosmos.staking.v1beta1.MsgDelegate":
			msgDelegate, ok := protoMsg.(*stakingtypes.MsgDelegate)
			if !ok {
				return nil, fmt.Errorf("Invalid proto msg for type: %s", typeUrl)
			}

			// Query account balance for the delegated denom
			balanceQueryReq := banktypes.NewQueryBalanceRequest(sdk.AccAddress(msgDelegate.DelegatorAddress), msgDelegate.Amount.Denom)
			resp, err := accountstd.QueryModule[banktypes.QueryBalanceResponse](ctx, balanceQueryReq)
			if err != nil {
				return nil, err
			}

			// Apply track delegation with previous state in the record
			newDelegatedFree, newDelegatedVesting := bva.TrackDelegation(
				sdk.Coins{*resp.Balance},
				getVestingFunc(sdkctx.BlockHeader().Time),
				sdk.Coins{msgDelegate.Amount},
				previousStateRecord.DelegatedFree,
				previousStateRecord.DelegatedVesting,
			)
			stateRecords.Records = append(stateRecords.Records, StateTransitionRecord{
				DelegatedFree:    newDelegatedFree,
				DelegatedVesting: newDelegatedVesting,
			})
		case "/cosmos.staking.v1beta1.MsgUndelegate":
			msgUndelegate, ok := protoMsg.(*stakingtypes.MsgUndelegate)
			if !ok {
				return nil, fmt.Errorf("Invalid proto msg for type: %s", typeUrl)
			}

			// Apply track delegation with previous state in the record
			newDelegatedFree, newDelegatedVesting := bva.TrackUndelegation(
				sdk.Coins{msgUndelegate.Amount},
				previousStateRecord.DelegatedFree,
				previousStateRecord.DelegatedVesting,
			)
			stateRecords.Records = append(stateRecords.Records, StateTransitionRecord{
				DelegatedFree:    newDelegatedFree,
				DelegatedVesting: newDelegatedVesting,
			})
		default:
			fmt.Println("Contiue with the execution")
		}
	}

	// execute messages
	responses, err := accountstd.ExecModuleAnys(ctx, msg.ExecutionMessages)
	if err != nil {
		return nil, err
	}

	// Apply the lastest delegation state to account
	// when execute is successfull. If no delegate or
	// undelegate action involve then apply the initial
	// state which is the account current state.
	newestStateRecord := stateRecords.Records[len(stateRecords.Records)-1]
	bva.DelegatedFree = newestStateRecord.DelegatedFree
	bva.DelegatedVesting = newestStateRecord.DelegatedVesting

	return &vestingtypes.MsgExecuteMessagesResponse{ExecutionMessagesResponse: responses}, nil
}

// --------------- Query -----------------

// LockedCoinsFromVesting returns all the coins that are not spendable (i.e. locked)
// for a vesting account given the current vesting coins. If no coins are locked,
// an empty slice of Coins is returned.
//
// CONTRACT: Delegated vesting coins and vestingCoins must be sorted.
func (bva BaseVestingAccount) LockedCoinsFromVesting(vestingCoins sdk.Coins) sdk.Coins {
	lockedCoins := vestingCoins.Sub(vestingCoins.Min(bva.DelegatedVesting)...)
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

// QueryOriginalVesting returns a vesting account's original vesting amount
func (bva BaseVestingAccount) QueryOriginalVesting(ctx context.Context, _ *vestingtypes.QueryOriginalVestingRequest) (
	*vestingtypes.QueryOriginalVestingResponse, error,
) {
	return &vestingtypes.QueryOriginalVestingResponse{
		OriginalVesting: bva.OriginalVesting,
	}, nil
}

// QueryDelegatedFree returns a vesting account's delegation amount that is not
// vesting.
func (bva BaseVestingAccount) QueryDelegatedFree(ctx context.Context, _ *vestingtypes.QueryDelegatedFreeRequest) (
	*vestingtypes.QueryDelegatedFreeResponse, error,
) {
	return &vestingtypes.QueryDelegatedFreeResponse{
		DelegatedFree: bva.DelegatedFree,
	}, nil
}

// QueryDelegatedVesting returns a vesting account's delegation amount that is
// still vesting.
func (bva BaseVestingAccount) QueryDelegatedVesting(ctx context.Context, _ *vestingtypes.QueryDelegatedVestingRequest) (
	*vestingtypes.QueryDelegatedVestingResponse, error,
) {
	return &vestingtypes.QueryDelegatedVestingResponse{
		DelegatedVesting: bva.DelegatedVesting,
	}, nil
}

// QueryEndTime returns a vesting account's end time
func (bva BaseVestingAccount) QueryEndTime(ctx context.Context, _ *vestingtypes.QueryEndTimeRequest) (
	*vestingtypes.QueryEndTimeResponse, error,
) {
	return &vestingtypes.QueryEndTimeResponse{
		EndTime: bva.EndTime,
	}, nil
}

// Validate checks for errors on the account fields
func (bva BaseVestingAccount) Validate() error {
	if bva.EndTime < 0 {
		return errors.New("end time cannot be negative")
	}

	if !bva.OriginalVesting.IsValid() || !bva.OriginalVesting.IsAllPositive() {
		return fmt.Errorf("invalid coins: %s", bva.OriginalVesting.String())
	}

	if !(bva.DelegatedVesting.IsAllLTE(bva.OriginalVesting)) {
		return errors.New("delegated vesting amount cannot be greater than original vesting amount")
	}

	return nil
}

// Only for implementing account interface, base vesting account
// served as a base for other types of vesting account only
// and should not be initialize as a stand alone vesting account type.
func (bva BaseVestingAccount) RegisterInitHandler(builder *accountstd.InitBuilder) {
}

func (bva BaseVestingAccount) RegisterExecuteHandlers(builder *accountstd.ExecuteBuilder) {

}

func (bva BaseVestingAccount) RegisterQueryHandlers(builder *accountstd.QueryBuilder) {
	accountstd.RegisterQueryHandler(builder, bva.QueryOriginalVesting)
	accountstd.RegisterQueryHandler(builder, bva.QueryDelegatedFree)
	accountstd.RegisterQueryHandler(builder, bva.QueryDelegatedVesting)
	accountstd.RegisterQueryHandler(builder, bva.QueryEndTime)
}
