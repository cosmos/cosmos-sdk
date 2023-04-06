package keeper

import (
	"context"

	"github.com/armon/go-metrics"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the distribution MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) SetWithdrawAddress(goCtx context.Context, msg *types.MsgSetWithdrawAddress) (*types.MsgSetWithdrawAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	delegatorAddress, err := k.authKeeper.StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	withdrawAddress, err := k.authKeeper.StringToBytes(msg.WithdrawAddress)
	if err != nil {
		return nil, err
	}
	err = k.SetWithdrawAddr(ctx, delegatorAddress, withdrawAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetWithdrawAddressResponse{}, nil
}

func (k msgServer) WithdrawDelegatorReward(goCtx context.Context, msg *types.MsgWithdrawDelegatorReward) (*types.MsgWithdrawDelegatorRewardResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	delegatorAddress, err := k.authKeeper.StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}
	amount, err := k.WithdrawDelegationRewards(ctx, delegatorAddress, valAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, a := range amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "withdraw_reward"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	return &types.MsgWithdrawDelegatorRewardResponse{Amount: amount}, nil
}

func (k msgServer) WithdrawValidatorCommission(goCtx context.Context, msg *types.MsgWithdrawValidatorCommission) (*types.MsgWithdrawValidatorCommissionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	amount, err := k.Keeper.WithdrawValidatorCommission(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	defer func() {
		for _, a := range amount {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "withdraw_commission"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	return &types.MsgWithdrawValidatorCommissionResponse{Amount: amount}, nil
}

func (k msgServer) FundCommunityPool(goCtx context.Context, msg *types.MsgFundCommunityPool) (*types.MsgFundCommunityPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	depositer, err := k.authKeeper.StringToBytes(msg.Depositor)
	if err != nil {
		return nil, err
	}
	if err := k.Keeper.FundCommunityPool(ctx, msg.Amount, depositer); err != nil {
		return nil, err
	}

	return &types.MsgFundCommunityPoolResponse{}, nil
}

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	if (!req.Params.BaseProposerReward.IsNil() && !req.Params.BaseProposerReward.IsZero()) || //nolint:staticcheck // deprecated but kept for backwards compatibility
		(!req.Params.BonusProposerReward.IsNil() && !req.Params.BonusProposerReward.IsZero()) { //nolint:staticcheck // deprecated but kept for backwards compatibility
		return nil, errorsmod.Wrapf(errors.ErrInvalidRequest, "cannot update base or bonus proposer reward because these are deprecated fields")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) CommunityPoolSpend(goCtx context.Context, req *types.MsgCommunityPoolSpend) (*types.MsgCommunityPoolSpendResponse, error) {
	if k.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	recipient, err := k.authKeeper.StringToBytes(req.Recipient)
	if err != nil {
		return nil, err
	}

	if k.bankKeeper.BlockedAddr(recipient) {
		return nil, errorsmod.Wrapf(errors.ErrUnauthorized, "%s is not allowed to receive external funds", req.Recipient)
	}

	if err := k.DistributeFromFeePool(ctx, req.Amount, recipient); err != nil {
		return nil, err
	}

	logger := k.Logger(ctx)
	logger.Info("transferred from the community pool to recipient", "amount", req.Amount.String(), "recipient", req.Recipient)

	return &types.MsgCommunityPoolSpendResponse{}, nil
}

func (k msgServer) DepositValidatorRewardsPool(goCtx context.Context, req *types.MsgDepositValidatorRewardsPool) (*types.MsgDepositValidatorRewardsPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	authority, err := k.authKeeper.StringToBytes(req.Authority)
	if err != nil {
		return nil, err
	}

	// deposit coins from sender's account to the distribution module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, authority, types.ModuleName, req.Amount); err != nil {
		return nil, err
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	validator := k.stakingKeeper.Validator(ctx, valAddr)
	if validator == nil {
		return nil, errorsmod.Wrapf(types.ErrNoValidatorExists, valAddr.String())
	}

	// Allocate tokens from the distribution module to the validator, which are
	// then distributed to the validator's delegators.
	reward := sdk.NewDecCoinsFromCoins(req.Amount...)
	k.AllocateTokensToValidator(ctx, validator, reward)

	logger := k.Logger(ctx)
	logger.Info(
		"transferred from rewards to validator rewards pool",
		"authority", req.Authority,
		"amount", req.Amount.String(),
		"validator", req.ValidatorAddress,
	)

	return &types.MsgDepositValidatorRewardsPoolResponse{}, nil
}
