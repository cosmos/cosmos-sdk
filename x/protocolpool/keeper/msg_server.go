package keeper

import (
	"context"

	"cosmossdk.io/errors"
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
	recipient, err := k.Keeper.authKeeper.AddressCodec().StringToBytes(msg.RecipientAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid recipient address: %s", err)
	}

	amount, err := k.claimFunds(ctx, recipient)
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
	return &types.MsgCreateContinuousFundResponse{}, nil
}

func (k MsgServer) CancelContinuousFund(ctx context.Context, msg *types.MsgCancelContinuousFund) (*types.MsgCancelContinuousFundResponse, error) {
	return &types.MsgCancelContinuousFundResponse{}, nil
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
