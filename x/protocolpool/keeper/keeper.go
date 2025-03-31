package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

// assert that this keeper can be used by x/distribution
var _ types.ExternalCommunityPoolKeeper = &Keeper{}

type Keeper struct {
	storeService store.KVStoreService

	authKeeper types.AccountKeeper
	bankKeeper types.BankKeeper

	cdc codec.BinaryCodec

	authority string

	// State
	Schema          collections.Schema
	ContinuousFunds collections.Map[sdk.AccAddress, types.ContinuousFund]
	Params          collections.Item[types.Params]
}

const (
	errModuleAccountNotSet = "%s module account has not been set"
)

func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, ak types.AccountKeeper, bk types.BankKeeper, authority string,
) Keeper {
	// ensure pool module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf(errModuleAccountNotSet, types.ModuleName))
	}
	// ensure protocol pool distribution account is set
	if addr := ak.GetModuleAddress(types.ProtocolPoolEscrowAccount); addr == nil {
		panic(fmt.Sprintf(errModuleAccountNotSet, types.ProtocolPoolEscrowAccount))
	}

	sb := collections.NewSchemaBuilder(storeService)

	keeper := Keeper{
		storeService:    storeService,
		authKeeper:      ak,
		bankKeeper:      bk,
		cdc:             cdc,
		authority:       authority,
		ContinuousFunds: collections.NewMap(sb, types.ContinuousFundsKey, "continuous_funds", sdk.AccAddressKey, codec.CollValue[types.ContinuousFund](cdc)),
		Params:          collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	keeper.Schema = schema

	return keeper
}

// GetAuthority returns the x/protocolpool module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// GetCommunityPoolModule gets the module name that funds should be sent to for the community pool.
// This is the address that x/distribution will send funds to for external management.
func (k Keeper) GetCommunityPoolModule() string {
	return types.ProtocolPoolEscrowAccount
}

// FundCommunityPool allows an account to directly fund the community fund pool.
func (k Keeper) FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleName, amount)
}

// DistributeFromCommunityPool distributes funds from the protocolpool module account to
// a receiver address.
func (k Keeper) DistributeFromCommunityPool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, receiveAddr, amount)
}

// GetCommunityPool gets the community pool balance.
func (k Keeper) GetCommunityPool(ctx sdk.Context) (sdk.Coins, error) {
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", types.ModuleName)
	}
	return k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress()), nil
}

// DistributeFunds sets the amount to be distributed among recipients.
// Get all valid continuous funds:
// - for each continuous fund, check if expired and remove if so
// - for each continuous fund, distribute funds according to percentage
// - distribute remaining funds to the community pool
//
// This function is run at the BeginBlocker method and therefore must be very safe.
func (k Keeper) DistributeFunds(ctx sdk.Context) error {
	// Get current balance of the intermediary module account
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ProtocolPoolEscrowAccount)
	if moduleAccount == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", types.ProtocolPoolEscrowAccount)
	}
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// only take into account the balances of denoms whitelisted in EnabledDistributionDenoms
	amountToDistribute := sdk.NewCoins()
	for _, denom := range params.EnabledDistributionDenoms {
		bal := k.bankKeeper.GetBalance(ctx, moduleAccount.GetAddress(), denom)
		amountToDistribute = append(amountToDistribute, bal)
	}

	// if the balance is zero, return early
	if amountToDistribute.IsZero() {
		return nil
	}

	remainingCoins := sdk.NewCoins(amountToDistribute...)
	iter, err := k.ContinuousFunds.Iterate(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create iterator for continuous funds: %w", err)
	}

	kValues, err := iter.KeyValues()
	if err != nil {
		return fmt.Errorf("failed to iterate continuous funds: %w", err)
	}

	blockTime := ctx.BlockTime()
	anyNegative := false
	for _, kv := range kValues {
		recipient := kv.Key
		fund := kv.Value

		// remove newly expired funds
		if fund.Expiry != nil && fund.Expiry.Before(blockTime) {
			err := k.ContinuousFunds.Remove(ctx, recipient)
			if err != nil {
				return fmt.Errorf("failed to remove fund for %s from ContinuousFunds: %w", recipient, err)
			}
			continue
		}

		amountToStream := PercentageCoinMul(fund.Percentage, amountToDistribute)
		remainingCoins, anyNegative = remainingCoins.SafeSub(amountToStream...)
		if anyNegative {
			return fmt.Errorf("negative funds for distribution from ContinuousFunds: %v", remainingCoins)
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ProtocolPoolEscrowAccount, recipient, amountToStream)
		if err != nil {
			// in the case where the address is unauthorized (blocked)
			// - remove the continuous fund
			// - add back the coins
			if errors.Is(err, sdkerrors.ErrUnauthorized) {
				ctx.Logger().Debug("recipient is unauthorized - removing the continuous fund", "error", err)
				remainingCoins = remainingCoins.Add(amountToStream...)
				err := k.ContinuousFunds.Remove(ctx, recipient)
				if err != nil {
					return fmt.Errorf("failed to remove fund for %s from ContinuousFunds: %w", recipient, err)
				}
				continue
			}

			return fmt.Errorf("failed to distribute fund for %s from ContinuousFunds: %w", recipient, err)
		}
	}

	// send all remaining funds to the community pool
	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ProtocolPoolEscrowAccount, types.ModuleName, remainingCoins); err != nil {
		return fmt.Errorf("failed to send coins to community pool: %w", err)
	}

	return nil
}

// GetAllContinuousFunds gets all continuous funds in the store.
func (k Keeper) GetAllContinuousFunds(ctx sdk.Context) ([]types.ContinuousFund, error) {
	it, err := k.ContinuousFunds.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	cf, err := it.Values()
	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (k Keeper) validateAuthority(authority string) error {
	if _, err := k.authKeeper.AddressCodec().StringToBytes(authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if k.authority != authority {
		return errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return nil
}

// PercentageCoinMul multiplies each coin in an sdk.Coins struct by the given percentage and returns the new
// value.
//
// When performing multiplication, the resulting values are truncated to an sdk.Int.
func PercentageCoinMul(percentage math.LegacyDec, coins sdk.Coins) sdk.Coins {
	ret := sdk.NewCoins()

	for _, denom := range coins.Denoms() {
		am := sdk.NewCoin(denom, percentage.MulInt(coins.AmountOf(denom)).TruncateInt())
		ret = ret.Add(am)
	}

	return ret
}
