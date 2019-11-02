package keeper

import (
	"encoding/binary"
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

	clientKeeper     types.ClientKeeper
	connectionKeeper types.ConnectionKeeper
	portKeeper       types.PortKeeper
}

// NewKeeper creates a new IBC channel Keeper instance
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType,
	clientKeeper types.ClientKeeper, connectionKeeper types.ConnectionKeeper,
	portKeeper types.PortKeeper,
) Keeper {
	return Keeper{
		storeKey:  key,
		cdc:       cdc,
		codespace: sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/channel",
		prefix:    []byte{},
		// prefix:           []byte(types.SubModuleName + "/"),                                          // "channel/"
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
		portKeeper:       portKeeper,
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
	fmt.Printf("setting channel %s at port %s\n", channelID, portID)
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
func (k Keeper) SetChannelCapability(ctx sdk.Context, portID, channelID string, key string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Set(types.KeyChannelCapabilityPath(portID, channelID), []byte(key))
}

// GetNextSequenceSend gets a channel's next send sequence from the store
func (k Keeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyNextSequenceSend(portID, channelID))
	if bz == nil {
		return 0, false
	}

	return binary.BigEndian.Uint64(bz), true
}

// SetNextSequenceSend sets a channel's next send sequence to the store
func (k Keeper) SetNextSequenceSend(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set(types.KeyNextSequenceSend(portID, channelID), bz)
}

// GetNextSequenceRecv gets a channel's next receive sequence from the store
func (k Keeper) GetNextSequenceRecv(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	bz := store.Get(types.KeyNextSequenceRecv(portID, channelID))
	if bz == nil {
		return 0, false
	}

	return binary.BigEndian.Uint64(bz), true
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

// SetPacketCommitment sets the packet commitment hash to the store
func (k Keeper) SetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64, commitmentHash []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Set(types.KeyPacketCommitment(portID, channelID, sequence), commitmentHash)
}

func (k Keeper) deletePacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Delete(types.KeyPacketCommitment(portID, channelID, sequence))
}

// SetPacketAcknowledgement sets the packet ack hash to the store
func (k Keeper) SetPacketAcknowledgement(ctx sdk.Context, portID, channelID string, sequence uint64, ackHash []byte) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), k.prefix)
	store.Set(types.KeyPacketAcknowledgement(portID, channelID, sequence), ackHash)
}
