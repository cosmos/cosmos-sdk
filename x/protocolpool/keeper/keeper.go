package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"

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
	// ensure stream account is set
	if addr := ak.GetModuleAddress(types.StreamAccount); addr == nil {
		panic(fmt.Sprintf(errModuleAccountNotSet, types.StreamAccount))
	}
	// ensure protocol pool distribution account is set
	if addr := ak.GetModuleAddress(types.ProtocolPoolDistrAccount); addr == nil {
		panic(fmt.Sprintf(errModuleAccountNotSet, types.ProtocolPoolDistrAccount))
	}

	sb := collections.NewSchemaBuilder(storeService)

	keeper := Keeper{
		storeService:    storeService,
		authKeeper:      ak,
		bankKeeper:      bk,
		cdc:             cdc,
		authority:       authority,
		ContinuousFunds: collections.NewMap(sb, types.ContinuousFundKey, "continuous_fund", sdk.AccAddressKey, codec.CollValue[types.ContinuousFund](cdc)),
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
	return types.ProtocolPoolDistrAccount
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

// DistributeFromStreamFunds distributes funds from the protocolpool's stream module account to
// a receiver address.
func (k Keeper) DistributeFromStreamFunds(ctx sdk.Context, amount sdk.Coins, receiveAddr []byte) error {
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.StreamAccount, receiveAddr, amount)
}

// GetCommunityPool gets the community pool balance.
func (k Keeper) GetCommunityPool(ctx sdk.Context) (sdk.Coins, error) {
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ModuleName)
	if moduleAccount == nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", types.ModuleName)
	}
	return k.bankKeeper.GetAllBalances(ctx, moduleAccount.GetAddress()), nil
}

// SetToDistribute sets the amount to be distributed among recipients.
func (k Keeper) SetToDistribute(ctx sdk.Context) error {
	// Get current balance of the intermediary module account
	moduleAccount := k.authKeeper.GetModuleAccount(ctx, types.ProtocolPoolDistrAccount)
	if moduleAccount == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", types.ProtocolPoolDistrAccount)
	}
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// only take into account the balances of denoms whitelisted in EnabledDistributionDenoms
	amountToDistribute := sdk.NewCoins()
	for _, denom := range params.EnabledDistributionDenoms {
		bal := k.bankKeeper.GetBalance(ctx, moduleAccount.GetAddress(), denom)
		amountToDistribute = amountToDistribute.Add(bal)
	}

	// if the balance is zero, return early
	if amountToDistribute.IsZero() {
		return nil
	}

	// TODO
	// still need to iter thru continuous funds
	// ....

	if err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ProtocolPoolDistrAccount, types.ModuleName, amountToDistribute); err != nil {
		return err
	}

	return nil
}

// TODO: test
func (k Keeper) GetAllContinuousFunds(ctx sdk.Context) ([]types.ContinuousFund, error) {
	var cf []types.ContinuousFund
	err := k.ContinuousFunds.Walk(ctx, nil, func(key sdk.AccAddress, value types.ContinuousFund) (stop bool, err error) {
		recipient, err := k.authKeeper.AddressCodec().BytesToString(key)
		if err != nil {
			return true, err
		}
		cf = append(cf, types.ContinuousFund{
			Recipient:  recipient,
			Percentage: value.Percentage,
			Expiry:     value.Expiry,
		})
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return cf, nil
}

func (k *Keeper) validateAuthority(authority string) error {
	if _, err := k.authKeeper.AddressCodec().StringToBytes(authority); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	if k.authority != authority {
		return errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return nil
}
