package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

type MsgServer struct {
	Keeper
}

var _ types.MsgServer = MsgServer{}

// NewMsgServerImpl returns an implementation of the protocolpool MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &MsgServer{Keeper: keeper}
}

func (k MsgServer) ClaimBudget(ctx context.Context, msg *types.MsgClaimBudget) (*types.MsgClaimBudgetResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	amount, err := k.claimFunds(sdkCtx, msg.RecipientAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimBudgetResponse{Amount: amount}, nil
}

func (k MsgServer) CreateBudget(ctx context.Context, msg *types.MsgCreateBudget) (*types.MsgCreateBudgetResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.validateAuthority(msg.GetAuthority()); err != nil {
		return nil, err
	}

	recipient, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(msg.RecipientAddress)
	if err != nil {
		return nil, err
	}

	budget, err := validateAndUpdateBudget(sdkCtx, *msg)
	if err != nil {
		return nil, err
	}

	// set budget proposal in state
	// Note: If two budgets to the same address are created, the budget would be updated with the new budget.
	err = k.Budgets.Set(sdkCtx, recipient, budget)
	if err != nil {
		return nil, err
	}
	return &types.MsgCreateBudgetResponse{}, nil
}

func (k MsgServer) FundCommunityPool(ctx context.Context, msg *types.MsgFundCommunityPool) (*types.MsgFundCommunityPoolResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	depositor, err := k.authKeeper.AddressCodec().StringToBytes(msg.Depositor)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid depositor address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	// send funds to community pool module account
	if err := k.Keeper.FundCommunityPool(sdkCtx, msg.Amount, depositor); err != nil {
		return nil, err
	}

	return &types.MsgFundCommunityPoolResponse{}, nil
}

func (k MsgServer) CommunityPoolSpend(ctx context.Context, msg *types.MsgCommunityPoolSpend) (*types.MsgCommunityPoolSpendResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	recipient, err := k.authKeeper.AddressCodec().StringToBytes(msg.Recipient)
	if err != nil {
		return nil, err
	}

	// distribute funds from community pool module account
	if err := k.Keeper.DistributeFromCommunityPool(sdkCtx, msg.Amount, recipient); err != nil {
		return nil, err
	}

	sdkCtx.Logger().Info("transferred from the community pool to recipient", "amount", msg.Amount.String(), "recipient", msg.Recipient)

	return &types.MsgCommunityPoolSpendResponse{}, nil
}

func (k MsgServer) CreateContinuousFund(ctx context.Context, msg *types.MsgCreateContinuousFund) (*types.MsgCreateContinuousFundResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	recipient, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(msg.Recipient)
	if err != nil {
		return nil, err
	}

	has, err := k.ContinuousFunds.Has(sdkCtx, recipient)
	if err != nil {
		return nil, err
	}
	if has {
		return nil, fmt.Errorf("continuous fund already exists for recipient %s", msg.Recipient)
	}

	// Validate the message fields
	err = validateContinuousFund(sdkCtx, *msg)
	if err != nil {
		return nil, err
	}

	// Check if total funds percentage exceeds 100%
	// If exceeds, we should not setup continuous fund proposal.
	totalStreamFundsPercentage := math.LegacyZeroDec()
	err = k.Keeper.ContinuousFunds.Walk(sdkCtx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (stop bool, err error) {
		totalStreamFundsPercentage = totalStreamFundsPercentage.Add(value.Percentage)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	totalStreamFundsPercentage = totalStreamFundsPercentage.Add(msg.Percentage)
	if totalStreamFundsPercentage.GT(math.LegacyOneDec()) {
		return nil, fmt.Errorf("cannot set continuous fund proposal\ntotal funds percentage exceeds 100\ncurrent total percentage: %s", totalStreamFundsPercentage.Sub(msg.Percentage).MulInt64(100).TruncateInt().String())
	}

	// Distribute funds to avoid giving this new fund more than it should get
	if err := k.IterateAndUpdateFundsDistribution(sdkCtx); err != nil {
		return nil, err
	}

	// Create continuous fund proposal
	cf := types.ContinuousFund{
		Recipient:  msg.Recipient,
		Percentage: msg.Percentage,
		Expiry:     msg.Expiry,
	}

	// Set continuous fund to the state
	err = k.ContinuousFunds.Set(sdkCtx, recipient, cf)
	if err != nil {
		return nil, err
	}

	err = k.RecipientFundDistributions.Set(sdkCtx, recipient, types.DistributionAmount{Amount: sdk.NewCoins()})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateContinuousFundResponse{}, nil
}

func (k MsgServer) WithdrawContinuousFund(ctx context.Context, msg *types.MsgWithdrawContinuousFund) (*types.MsgWithdrawContinuousFundResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	recipient, err := k.authKeeper.AddressCodec().StringToBytes(msg.RecipientAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	err = k.IterateAndUpdateFundsDistribution(sdkCtx)
	if err != nil {
		return nil, fmt.Errorf("error while iterating all the continuous funds: %w", err)
	}

	// withdraw continuous fund
	withdrawnAmount, err := k.withdrawRecipientFunds(sdkCtx, recipient)
	if err != nil {
		return nil, fmt.Errorf("error while withdrawing recipient funds for recipient: %w", err)
	}

	return &types.MsgWithdrawContinuousFundResponse{Amount: withdrawnAmount}, nil
}

func (k MsgServer) CancelContinuousFund(ctx context.Context, msg *types.MsgCancelContinuousFund) (*types.MsgCancelContinuousFundResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	recipient, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(msg.RecipientAddress)
	if err != nil {
		return nil, err
	}

	canceledHeight := sdkCtx.BlockHeight()
	canceledTime := sdkCtx.BlockTime()

	// distribute funds before withdrawing
	if err = k.IterateAndUpdateFundsDistribution(sdkCtx); err != nil {
		return nil, err
	}

	// withdraw funds if any are allocated
	withdrawnFunds, err := k.withdrawRecipientFunds(sdkCtx, recipient)
	if err != nil && !errors.Is(err, types.ErrNoRecipientFound) {
		return nil, fmt.Errorf("error while withdrawing already allocated funds for recipient %s: %w", msg.RecipientAddress, err)
	}

	if err := k.ContinuousFunds.Remove(sdkCtx, recipient); err != nil {
		return nil, fmt.Errorf("failed to remove continuous fund for recipient %s: %w", msg.RecipientAddress, err)
	}

	if err := k.RecipientFundDistributions.Remove(sdkCtx, recipient); err != nil {
		return nil, fmt.Errorf("failed to remove recipient fund distribution for recipient %s: %w", msg.RecipientAddress, err)
	}

	return &types.MsgCancelContinuousFundResponse{
		CanceledTime:           canceledTime,
		CanceledHeight:         uint64(canceledHeight),
		RecipientAddress:       msg.RecipientAddress,
		WithdrawnAllocatedFund: withdrawnFunds,
	}, nil
}

func (k MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	if err := k.validateAuthority(msg.GetAuthority()); err != nil {
		return nil, err
	}

	if err := k.Params.Set(sdkCtx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
