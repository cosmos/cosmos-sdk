package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochingkeeper "github.com/cosmos/cosmos-sdk/x/epoching/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Keeper of the slashing store
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        codec.BinaryMarshaler
	sk         types.StakingKeeper
	ek         epochingkeeper.Keeper
	paramspace types.ParamSubspace
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc codec.BinaryMarshaler, key sdk.StoreKey, sk types.StakingKeeper, paramspace types.ParamSubspace) Keeper {
	// set KeyTable if it has not already been set
	if !paramspace.HasKeyTable() {
		paramspace = paramspace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   key,
		cdc:        cdc,
		sk:         sk,
		ek:         epochingkeeper.NewKeeper(cdc, key),
		paramspace: paramspace,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// AddPubkey sets a address-pubkey relation
func (k Keeper) AddPubkey(ctx sdk.Context, pubkey cryptotypes.PubKey) {
	addr := pubkey.Address()

	pkStr, err := sdk.Bech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, pubkey)
	if err != nil {
		panic(fmt.Errorf("error while setting address-pubkey relation: %s", addr))
	}

	k.setAddrPubkeyRelation(ctx, addr, pkStr)
}

// GetPubkey returns the pubkey from the adddress-pubkey relation
func (k Keeper) GetPubkey(ctx sdk.Context, address cryptotypes.Address) (cryptotypes.PubKey, error) {
	store := ctx.KVStore(k.storeKey)

	var pubkey gogotypes.StringValue
	err := k.cdc.UnmarshalBinaryBare(store.Get(types.AddrPubkeyRelationKey(address)), &pubkey)
	if err != nil {
		return nil, fmt.Errorf("address %s not found", sdk.ConsAddress(address))
	}

	pkStr, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pubkey.Value)
	if err != nil {
		return pkStr, err
	}

	return pkStr, nil
}

// Slash attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes.
func (k Keeper) Slash(ctx sdk.Context, consAddr sdk.ConsAddress, fraction sdk.Dec, power, distributionHeight int64) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
			sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueDoubleSign),
		),
	)
	// epoch slashing action for next epoch
	validator := k.sk.ValidatorByConsAddr(ctx, consAddr)
	if validator != nil {
		k.ek.QueueMsgForEpoch(ctx, k.sk.GetEpochNumber(ctx), types.SlashEvent{
			Address:                validator.GetOperator(),
			ValidatorVotingPercent: sdk.NewDec(power),
			SlashedSoFar:           fraction,
		})
	}
}

// Jail attempts to jail a validator. The slash is delegated to the staking module
// to make the necessary validator changes.
func (k Keeper) Jail(ctx sdk.Context, consAddr sdk.ConsAddress) {
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
		),
	)
	// delegate action to staking keeper for Jail
	k.sk.Jail(ctx, consAddr)
}

func (k Keeper) setAddrPubkeyRelation(ctx sdk.Context, addr cryptotypes.Address, pubkey string) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshalBinaryBare(&gogotypes.StringValue{Value: pubkey})
	store.Set(types.AddrPubkeyRelationKey(addr), bz)
}

func (k Keeper) deleteAddrPubkeyRelation(ctx sdk.Context, addr cryptotypes.Address) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.AddrPubkeyRelationKey(addr))
}
