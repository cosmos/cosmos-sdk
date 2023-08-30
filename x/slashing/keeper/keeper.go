package keeper

import (
	"context"
	"fmt"

	st "cosmossdk.io/api/cosmos/staking/v1beta1"
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// Keeper of the slashing store
type Keeper struct {
	storeService storetypes.KVStoreService
	cdc          codec.BinaryCodec
	legacyAmino  *codec.LegacyAmino
	sk           types.StakingKeeper

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority                  string
	Schema                     collections.Schema
	Params                     collections.Item[types.Params]
	ValidatorSigningInfo       collections.Map[sdk.ConsAddress, types.ValidatorSigningInfo]
	AddrPubkeyRelation         collections.Map[[]byte, cryptotypes.PubKey]
	ValidatorMissedBlockBitmap collections.Map[collections.Pair[[]byte, uint64], []byte]
}

// NewKeeper creates a slashing keeper
func NewKeeper(cdc codec.BinaryCodec, legacyAmino *codec.LegacyAmino, storeService storetypes.KVStoreService, sk types.StakingKeeper, authority string) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		legacyAmino:  legacyAmino,
		sk:           sk,
		authority:    authority,
		Params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		ValidatorSigningInfo: collections.NewMap(
			sb,
			types.ValidatorSigningInfoKeyPrefix,
			"validator_signing_info",
			sdk.LengthPrefixedAddressKey(sdk.ConsAddressKey), // nolint: staticcheck // sdk.LengthPrefixedAddressKey is needed to retain state compatibility
			codec.CollValue[types.ValidatorSigningInfo](cdc),
		),
		AddrPubkeyRelation: collections.NewMap(
			sb,
			types.AddrPubkeyRelationKeyPrefix,
			"addr_pubkey_relation",
			sdk.LengthPrefixedBytesKey, // sdk.LengthPrefixedBytesKey is needed to retain state compatibility
			codec.CollInterfaceValue[cryptotypes.PubKey](cdc),
		),
		ValidatorMissedBlockBitmap: collections.NewMap(
			sb,
			types.ValidatorMissedBlockBitmapKeyPrefix,
			"validator_missed_block_bitmap",
			collections.PairKeyCodec(sdk.LengthPrefixedBytesKey, collections.Uint64Key),
			collections.BytesValue,
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// GetAuthority returns the x/slashing module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// GetPubkey returns the pubkey from the adddress-pubkey relation
func (k Keeper) GetPubkey(ctx context.Context, a cryptotypes.Address) (cryptotypes.PubKey, error) {
	return k.AddrPubkeyRelation.Get(ctx, a)
}

// Slash attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes. It specifies no intraction reason.
func (k Keeper) Slash(ctx context.Context, consAddr sdk.ConsAddress, fraction sdkmath.LegacyDec, power, distributionHeight int64) error {
	return k.SlashWithInfractionReason(ctx, consAddr, fraction, power, distributionHeight, st.Infraction_INFRACTION_UNSPECIFIED)
}

// SlashWithInfractionReason attempts to slash a validator. The slash is delegated to the staking
// module to make the necessary validator changes. It specifies an intraction reason.
func (k Keeper) SlashWithInfractionReason(ctx context.Context, consAddr sdk.ConsAddress, fraction sdkmath.LegacyDec, power, distributionHeight int64, infraction st.Infraction) error {
	coinsBurned, err := k.sk.SlashWithInfractionReason(ctx, consAddr, distributionHeight, power, fraction, infraction)
	if err != nil {
		return err
	}

	reasonAttr := sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueUnspecified)
	switch infraction {
	case st.Infraction_INFRACTION_DOUBLE_SIGN:
		reasonAttr = sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueDoubleSign)
	case st.Infraction_INFRACTION_DOWNTIME:
		reasonAttr = sdk.NewAttribute(types.AttributeKeyReason, types.AttributeValueMissingSignature)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyAddress, consAddr.String()),
			sdk.NewAttribute(types.AttributeKeyPower, fmt.Sprintf("%d", power)),
			reasonAttr,
			sdk.NewAttribute(types.AttributeKeyBurnedCoins, coinsBurned.String()),
		),
	)
	return nil
}

// Jail attempts to jail a validator. The slash is delegated to the staking module
// to make the necessary validator changes.
func (k Keeper) Jail(ctx context.Context, consAddr sdk.ConsAddress) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err := k.sk.Jail(sdkCtx, consAddr)
	if err != nil {
		return err
	}
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlash,
			sdk.NewAttribute(types.AttributeKeyJailed, consAddr.String()),
		),
	)
	return nil
}
