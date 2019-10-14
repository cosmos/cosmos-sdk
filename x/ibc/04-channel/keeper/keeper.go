package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Keeper defines the IBC channel keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store

	connectionKeeper types.ConnectionKeeper
	// TODO: portKeeper
}

// NewKeeper creates a new IBC channel Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType, ck types.ConnectionKeeper) Keeper {
	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		codespace:        sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/channel",
		prefix:           []byte(types.SubModuleName + "/"),                                          // "channel/"
		connectionKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// GetChannel returns a channel with a particular identifier binded to a specific port
func (k Keeper) GetChannel(ctx sdk.Context, portID, channelID string) (types.Channel, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyChannel(portID, channelID))
	if bz == nil {
		return types.Channel{}, false
	}

	var channel types.Channel
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &channel)
	return channel, true
}

// SetChannel sets a channel to the store
func (k Keeper) SetChannel(ctx sdk.Context, portID, channelID string, channel types.Channel) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(channel)
	store.Set(types.KeyChannel(portID, channelID), bz)
}

// GetChannelCapability gets a channel's capability key from the store
func (k Keeper) GetChannelCapability(ctx sdk.Context, portID, channelID string) (string, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyChannelCapabilityPath(portID, channelID))
	if bz == nil {
		return "", false
	}

	return string(bz), true
}

// SetChannelCapability sets a channel's capability key to the store
// TODO: is the key a string ?
func (k Keeper) SetChannelCapability(ctx sdk.Context, portID, channelID string, key string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Set(types.KeyChannelCapabilityPath(portID, channelID), []byte(key))
}

// SetNextSequenceSend sets a channel's next send sequence to the store
func (k Keeper) SetNextSequenceSend(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set(types.KeyNextSequenceSend(portID, channelID), bz)
}

// SetNextSequenceRecv sets a channel's next receive sequence to the store
func (k Keeper) SetNextSequenceRecv(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set(types.KeyNextSequenceRecv(portID, channelID), bz)
}

// GetPacketCommitment gets the packet commitment hash from the store
func (k Keeper) GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyPacketCommitment(portID, channelID, sequence))
	return bz
}

func (k Keeper) deletePacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Delete(types.KeyPacketCommitment(portID, channelID, sequence))
}
