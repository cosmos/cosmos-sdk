package keeper

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-metrics"

	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	pooltypes "cosmossdk.io/api/cosmos/protocolpool/v1"
	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
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

func (k msgServer) SetWithdrawAddress(ctx context.Context, msg *types.MsgSetWithdrawAddress) (*types.MsgSetWithdrawAddressResponse, error) {
	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
	}

	withdrawAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.WithdrawAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid withdraw address: %s", err)
	}

	err = k.SetWithdrawAddr(ctx, delegatorAddress, withdrawAddress)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetWithdrawAddressResponse{}, nil
}

func (k msgServer) WithdrawDelegatorReward(ctx context.Context, msg *types.MsgWithdrawDelegatorReward) (*types.MsgWithdrawDelegatorRewardResponse, error) {
	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
	}

	delegatorAddress, err := k.authKeeper.AddressCodec().StringToBytes(msg.DelegatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid delegator address: %s", err)
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

func (k msgServer) WithdrawValidatorCommission(ctx context.Context, msg *types.MsgWithdrawValidatorCommission) (*types.MsgWithdrawValidatorCommissionResponse, error) {
	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid validator address: %s", err)
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

// NOTE: This method uses deprecated message request. Use FundCommunityPool from x/protocolpool module instead.
func (k msgServer) FundCommunityPool(ctx context.Context, msg *types.MsgFundCommunityPool) (*types.MsgFundCommunityPoolResponse, error) { //nolint:staticcheck // we're using a deprecated call for compatibility
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	amount := make([]*basev1beta1.Coin, len(msg.Amount))
	for i, coin := range msg.Amount {
		amount[i] = &basev1beta1.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		}
	}
	poolMsg := pooltypes.MsgFundCommunityPool{
		Amount:    amount,
		Depositor: msg.Depositor,
	}
	// Pass the msg to the MessageRouter
	handler := k.router.Handler(&poolMsg)
	if handler == nil {
		return nil, fmt.Errorf("message not recognized by router: %s", sdk.MsgTypeURL(&poolMsg))
	}
	msgResp, err := handler(sdkCtx, &poolMsg)
	if err != nil {
		return nil, err
	}

	events := msgResp.Events
	sdkEvents := make([]sdk.Event, 0, len(events))
	for _, event := range events {
		sdkEvents = append(sdkEvents, sdk.Event(event))
	}
	sdkCtx.EventManager().EmitEvents(sdkEvents)

	return &types.MsgFundCommunityPoolResponse{}, nil //nolint:staticcheck // we're using a deprecated call for compatibility
}

func (k msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := k.validateAuthority(msg.Authority); err != nil {
		return nil, err
	}

	if (!msg.Params.BaseProposerReward.IsNil() && !msg.Params.BaseProposerReward.IsZero()) || //nolint:staticcheck // deprecated but kept for backwards compatibility
		(!msg.Params.BonusProposerReward.IsNil() && !msg.Params.BonusProposerReward.IsZero()) { //nolint:staticcheck // deprecated but kept for backwards compatibility
		return nil, errors.Wrapf(sdkerrors.ErrInvalidRequest, "cannot update base or bonus proposer reward because these are deprecated fields")
	}

	if err := msg.Params.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := k.Params.Set(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// NOTE: This method uses deprecated message request. Use CommunityPoolSpend from x/protocolpool module instead.
func (k msgServer) CommunityPoolSpend(ctx context.Context, msg *types.MsgCommunityPoolSpend) (*types.MsgCommunityPoolSpendResponse, error) { //nolint:staticcheck // we're using a deprecated call for compatibility
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	amount := make([]*basev1beta1.Coin, len(msg.Amount))
	for i, coin := range msg.Amount {
		amount[i] = &basev1beta1.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.String(),
		}
	}

	poolMsg := pooltypes.MsgCommunityPoolSpend{
		Authority: msg.Authority,
		Recipient: msg.Recipient,
		Amount:    amount,
	}

	// Pass the msg to the MessageRouter
	handler := k.router.Handler(&poolMsg)
	if handler == nil {
		return nil, fmt.Errorf("message not recognized by router: %s", sdk.MsgTypeURL(&poolMsg))
	}
	msgResp, err := handler(sdkCtx, &poolMsg)
	if err != nil {
		return nil, err
	}

	events := msgResp.Events
	sdkEvents := make([]sdk.Event, 0, len(events))
	for _, event := range events {
		sdkEvents = append(sdkEvents, sdk.Event(event))
	}
	sdkCtx.EventManager().EmitEvents(sdkEvents)

	return &types.MsgCommunityPoolSpendResponse{}, nil //nolint:staticcheck // we're using a deprecated call for compatibility
}

func (k msgServer) DepositValidatorRewardsPool(ctx context.Context, msg *types.MsgDepositValidatorRewardsPool) (*types.MsgDepositValidatorRewardsPoolResponse, error) {
	depositor, err := k.authKeeper.AddressCodec().StringToBytes(msg.Depositor)
	if err != nil {
		return nil, err
	}

	// deposit coins from depositor's account to the distribution module
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositor, types.ModuleName, msg.Amount); err != nil {
		return nil, err
	}

	valAddr, err := k.stakingKeeper.ValidatorAddressCodec().StringToBytes(msg.ValidatorAddress)
	if err != nil {
		return nil, err
	}

	validator, err := k.stakingKeeper.Validator(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	if validator == nil {
		return nil, errors.Wrapf(types.ErrNoValidatorExists, msg.ValidatorAddress)
	}

	// Allocate tokens from the distribution module to the validator, which are
	// then distributed to the validator's delegators.
	reward := sdk.NewDecCoinsFromCoins(msg.Amount...)
	if err = k.AllocateTokensToValidator(ctx, validator, reward); err != nil {
		return nil, err
	}

	logger := k.Logger(ctx)
	logger.Info(
		"transferred from rewards to validator rewards pool",
		"depositor", msg.Depositor,
		"amount", msg.Amount.String(),
		"validator", msg.ValidatorAddress,
	)

	return &types.MsgDepositValidatorRewardsPoolResponse{}, nil
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
