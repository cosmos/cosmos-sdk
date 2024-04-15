package keeper

import (
	"context"
	errorspkg "errors"
	"fmt"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
	amount, err := k.claimFunds(ctx, msg.RecipientAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimBudgetResponse{Amount: amount}, nil
}

func (k MsgServer) SubmitBudgetProposal(ctx context.Context, msg *types.MsgSubmitBudgetProposal) (*types.MsgSubmitBudgetProposalResponse, error) {
	if err := k.validateAuthority(msg.GetAuthority()); err != nil {
		return nil, err
	}

	recipient, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(msg.RecipientAddress)
	if err != nil {
		return nil, err
	}

	budgetProposal, err := k.validateAndUpdateBudgetProposal(ctx, *msg)
	if err != nil {
		return nil, err
	}

	// set budget proposal in state
	// Note: If two budgets to the same address are created, the budget would be updated with the new budget.
	err = k.BudgetProposal.Set(ctx, recipient, *budgetProposal)
	if err != nil {
		return nil, err
	}
	return &types.MsgSubmitBudgetProposalResponse{}, nil
}

func (k MsgServer) FundCommunityPool(ctx context.Context, msg *types.MsgFundCommunityPool) (*types.MsgFundCommunityPoolResponse, error) {
	depositor, err := k.authKeeper.AddressCodec().StringToBytes(msg.Depositor)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid depositor address: %s", err)
	}

	if err := validateAmount(msg.Amount); err != nil {
		return nil, err
	}

	// send funds to community pool module account
	if err := k.Keeper.FundCommunityPool(ctx, msg.Amount, depositor); err != nil {
		return nil, err
	}

	return &types.MsgFundCommunityPoolResponse{}, nil
}

func (k MsgServer) CommunityPoolSpend(ctx context.Context, msg *types.MsgCommunityPoolSpend) (*types.MsgCommunityPoolSpendResponse, error) {
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
	if err := k.Keeper.DistributeFromCommunityPool(ctx, msg.Amount, recipient); err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("transferred from the community pool to recipient", "amount", msg.Amount.String(), "recipient", msg.Recipient)

	return &types.MsgCommunityPoolSpendResponse{}, nil
}

func (k MsgServer) CreateContinuousFund(ctx context.Context, msg *types.MsgCreateContinuousFund) (*types.MsgCreateContinuousFundResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	recipient, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(msg.Recipient)
	if err != nil {
		return nil, err
	}

	// Validate the message fields
	err = k.validateContinuousFund(ctx, *msg)
	if err != nil {
		return nil, err
	}

	// Check if total funds percentage exceeds 100%
	// If exceeds, we should not setup continuous fund proposal.
	totalStreamFundsPercentage := math.ZeroInt()
	err = k.Keeper.RecipientFundPercentage.Walk(ctx, nil, func(key sdk.AccAddress, value math.Int) (stop bool, err error) {
		totalStreamFundsPercentage = totalStreamFundsPercentage.Add(value)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	percentage := msg.Percentage.MulInt64(100)
	totalStreamFundsPercentage = totalStreamFundsPercentage.Add(percentage.TruncateInt())
	if totalStreamFundsPercentage.GT(math.NewInt(100)) {
		return nil, fmt.Errorf("cannot set continuous fund proposal\ntotal funds percentage exceeds 100\ncurrent total percentage: %v", totalStreamFundsPercentage.Sub(percentage.TruncateInt()))
	}

	// Create continuous fund proposal
	cf := types.ContinuousFund{
		Recipient:  msg.Recipient,
		Percentage: msg.Percentage,
		Expiry:     msg.Expiry,
	}

	// Set continuous fund to the state
	err = k.ContinuousFund.Set(ctx, recipient, cf)
	if err != nil {
		return nil, err
	}

	// Set recipient fund percentage & distribution
	err = k.RecipientFundPercentage.Set(ctx, recipient, percentage.TruncateInt())
	if err != nil {
		return nil, err
	}
	err = k.RecipientFundDistribution.Set(ctx, recipient, math.ZeroInt())
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateContinuousFundResponse{}, nil
}

func (k MsgServer) WithdrawContinuousFund(ctx context.Context, msg *types.MsgWithdrawContinuousFund) (*types.MsgWithdrawContinuousFundResponse, error) {
	amount, err := k.withdrawContinuousFund(ctx, msg.RecipientAddress)
	if err != nil {
		return nil, err
	}
	if amount.IsNil() {
		k.Logger(ctx).Info(fmt.Sprintf("no distribution amount found for recipient %s", msg.RecipientAddress))
	}

	return &types.MsgWithdrawContinuousFundResponse{Amount: amount}, nil
}

func (k MsgServer) CancelContinuousFund(ctx context.Context, msg *types.MsgCancelContinuousFund) (*types.MsgCancelContinuousFundResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	recipient, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(msg.RecipientAddress)
	if err != nil {
		return nil, err
	}

	canceledHeight := k.environment.HeaderService.GetHeaderInfo(ctx).Height
	canceledTime := k.environment.HeaderService.GetHeaderInfo(ctx).Time

	found, err := k.ContinuousFund.Has(ctx, recipient)
	if !found {
		return nil, fmt.Errorf("no recipient found to cancel continuous fund: %s", msg.RecipientAddress)
	}
	if err != nil {
		return nil, err
	}

	// withdraw funds if any are allocated
	withdrawnFunds, err := k.withdrawRecipientFunds(ctx, msg.RecipientAddress)
	if err != nil && !errorspkg.Is(err, types.ErrNoRecipientFund) {
		return nil, fmt.Errorf("error while withdrawing already allocated funds for recipient %s: %v", msg.RecipientAddress, err)
	}

	if err := k.ContinuousFund.Remove(ctx, recipient); err != nil {
		return nil, fmt.Errorf("failed to remove continuous fund for recipient %s: %w", msg.RecipientAddress, err)
	}

	return &types.MsgCancelContinuousFundResponse{
		CanceledTime:           canceledTime,
		CanceledHeight:         uint64(canceledHeight),
		RecipientAddress:       msg.RecipientAddress,
		WithdrawnAllocatedFund: withdrawnFunds,
	}, nil
}

func (k *Keeper) validateAuthority(authority string) error {
	if _, err := k.authKeeper.AddressCodec().StringToBytes(authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if k.authority != authority {
		return errors.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return nil
}

func validateAmount(amount sdk.Coins) error {
	if amount == nil {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, "amount cannot be nil")
	}

	if err := amount.Validate(); err != nil {
		return errors.Wrap(sdkerrors.ErrInvalidCoins, amount.String())
	}

	return nil
}
