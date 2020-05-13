package keeper

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
	db "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Keeper defines the IBC channel keeper
type Keeper struct {
	storeKey         sdk.StoreKey
	cdc              codec.Marshaler
	clientKeeper     types.ClientKeeper
	connectionKeeper types.ConnectionKeeper
	portKeeper       types.PortKeeper
	scopedKeeper     capability.ScopedKeeper
}

// NewKeeper creates a new IBC channel Keeper instance
func NewKeeper(
	cdc codec.Marshaler, key sdk.StoreKey,
	clientKeeper types.ClientKeeper, connectionKeeper types.ConnectionKeeper,
	portKeeper types.PortKeeper, scopedKeeper capability.ScopedKeeper,
) Keeper {
	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		clientKeeper:     clientKeeper,
		connectionKeeper: connectionKeeper,
		portKeeper:       portKeeper,
		scopedKeeper:     scopedKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", host.ModuleName, types.SubModuleName))
}

// GetChannel returns a channel with a particular identifier binded to a specific port
func (k Keeper) GetChannel(ctx sdk.Context, portID, channelID string) (types.Channel, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.KeyChannel(portID, channelID))
	if bz == nil {
		return types.Channel{}, false
	}

	var channel types.Channel
	k.cdc.MustUnmarshalBinaryBare(bz, &channel)
	return channel, true
}

// SetChannel sets a channel to the store
func (k Keeper) SetChannel(ctx sdk.Context, portID, channelID string, channel types.Channel) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(&channel)
	store.Set(host.KeyChannel(portID, channelID), bz)
}

// GetNextSequenceSend gets a channel's next send sequence from the store
func (k Keeper) GetNextSequenceSend(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.KeyNextSequenceSend(portID, channelID))
	if bz == nil {
		return 0, false
	}

	return binary.BigEndian.Uint64(bz), true
}

// SetNextSequenceSend sets a channel's next send sequence to the store
func (k Keeper) SetNextSequenceSend(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set(host.KeyNextSequenceSend(portID, channelID), bz)
}

// GetNextSequenceRecv gets a channel's next receive sequence from the store
func (k Keeper) GetNextSequenceRecv(ctx sdk.Context, portID, channelID string) (uint64, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.KeyNextSequenceRecv(portID, channelID))
	if bz == nil {
		return 0, false
	}

	return binary.BigEndian.Uint64(bz), true
}

// SetNextSequenceRecv sets a channel's next receive sequence to the store
func (k Keeper) SetNextSequenceRecv(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set(host.KeyNextSequenceRecv(portID, channelID), bz)
}

// GetPacketCommitment gets the packet commitment hash from the store
func (k Keeper) GetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) []byte {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.KeyPacketCommitment(portID, channelID, sequence))
	return bz
}

// SetPacketCommitment sets the packet commitment hash to the store
func (k Keeper) SetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64, commitmentHash []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(host.KeyPacketCommitment(portID, channelID, sequence), commitmentHash)
}

func (k Keeper) deletePacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(host.KeyPacketCommitment(portID, channelID, sequence))
}

// SetPacketAcknowledgement sets the packet ack hash to the store
func (k Keeper) SetPacketAcknowledgement(ctx sdk.Context, portID, channelID string, sequence uint64, ackHash []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(host.KeyPacketAcknowledgement(portID, channelID, sequence), ackHash)
}

// GetPacketAcknowledgement gets the packet ack hash from the store
func (k Keeper) GetPacketAcknowledgement(ctx sdk.Context, portID, channelID string, sequence uint64) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(host.KeyPacketAcknowledgement(portID, channelID, sequence))
	if bz == nil {
		return nil, false
	}
	return bz, true
}

// IteratePacketSequence provides an iterator over all send and receive sequences. For each
// sequence, cb will be called. If the cb returns true, the iterator will close
// and stop.
func (k Keeper) IteratePacketSequence(ctx sdk.Context, send bool, cb func(portID, channelID string, sequence uint64) bool) {
	store := ctx.KVStore(k.storeKey)
	var iterator db.Iterator
	if send {
		iterator = sdk.KVStorePrefixIterator(store, []byte(host.KeyNextSeqSendPrefix))
	} else {
		iterator = sdk.KVStorePrefixIterator(store, []byte(host.KeyNextSeqRecvPrefix))
	}

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")
		portID := keySplit[2]
		channelID := keySplit[4]

		sequence := sdk.BigEndianToUint64(iterator.Value())

		if cb(portID, channelID, sequence) {
			break
		}
	}
}

// GetAllPacketSendSeqs returns all stored next send sequences.
func (k Keeper) GetAllPacketSendSeqs(ctx sdk.Context) (seqs []types.PacketSequence) {
	k.IteratePacketSequence(ctx, true, func(portID, channelID string, nextSendSeq uint64) bool {
		ps := types.NewPacketSequence(portID, channelID, nextSendSeq)
		seqs = append(seqs, ps)
		return false
	})
	return seqs
}

// GetAllPacketRecvSeqs returns all stored next recv sequences.
func (k Keeper) GetAllPacketRecvSeqs(ctx sdk.Context) (seqs []types.PacketSequence) {
	k.IteratePacketSequence(ctx, false, func(portID, channelID string, nextRecvSeq uint64) bool {
		ps := types.NewPacketSequence(portID, channelID, nextRecvSeq)
		seqs = append(seqs, ps)
		return false
	})
	return seqs
}

// IteratePacketCommitment provides an iterator over all PacketCommitment objects. For each
// aknowledgement, cb will be called. If the cb returns true, the iterator will close
// and stop.
func (k Keeper) IteratePacketCommitment(ctx sdk.Context, cb func(portID, channelID string, sequence uint64, hash []byte) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(host.KeyPacketCommitmentPrefix))
	k.iterateHashes(ctx, iterator, cb)
}

// GetAllPacketCommitments returns all stored PacketCommitments objects.
func (k Keeper) GetAllPacketCommitments(ctx sdk.Context) (commitments []types.PacketAckCommitment) {
	k.IteratePacketCommitment(ctx, func(portID, channelID string, sequence uint64, hash []byte) bool {
		pc := types.NewPacketAckCommitment(portID, channelID, sequence, hash)
		commitments = append(commitments, pc)
		return false
	})
	return commitments
}

// IteratePacketAcknowledgement provides an iterator over all PacketAcknowledgement objects. For each
// aknowledgement, cb will be called. If the cb returns true, the iterator will close
// and stop.
func (k Keeper) IteratePacketAcknowledgement(ctx sdk.Context, cb func(portID, channelID string, sequence uint64, hash []byte) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(host.KeyPacketAckPrefix))
	k.iterateHashes(ctx, iterator, cb)
}

// GetAllPacketAcks returns all stored PacketAcknowledgements objects.
func (k Keeper) GetAllPacketAcks(ctx sdk.Context) (acks []types.PacketAckCommitment) {
	k.IteratePacketAcknowledgement(ctx, func(portID, channelID string, sequence uint64, ack []byte) bool {
		packetAck := types.NewPacketAckCommitment(portID, channelID, sequence, ack)
		acks = append(acks, packetAck)
		return false
	})
	return acks
}

// IterateChannels provides an iterator over all Channel objects. For each
// Channel, cb will be called. If the cb returns true, the iterator will close
// and stop.
func (k Keeper) IterateChannels(ctx sdk.Context, cb func(types.IdentifiedChannel) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(host.KeyChannelPrefix))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var channel types.Channel
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &channel)

		portID, channelID := host.MustParseChannelPath(string(iterator.Key()))
		identifiedChannel := types.NewIdentifiedChannel(portID, channelID, channel)
		if cb(identifiedChannel) {
			break
		}
	}
}

// GetAllChannels returns all stored Channel objects.
func (k Keeper) GetAllChannels(ctx sdk.Context) (channels []types.IdentifiedChannel) {
	k.IterateChannels(ctx, func(channel types.IdentifiedChannel) bool {
		channels = append(channels, channel)
		return false
	})
	return channels
}

// LookupModuleByChannel will return the IBCModule along with the capability associated with a given channel defined by its portID and channelID
func (k Keeper) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capability.Capability, error) {
	modules, cap, err := k.scopedKeeper.LookupModules(ctx, host.ChannelCapabilityPath(portID, channelID))
	if err != nil {
		return "", nil, err
	}

	return porttypes.GetModuleOwner(modules), cap, nil
}

// common functionality for IteratePacketCommitment and IteratePacketAcknowledgemen
func (k Keeper) iterateHashes(_ sdk.Context, iterator db.Iterator, cb func(portID, channelID string, sequence uint64, hash []byte) bool) {
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")
		portID := keySplit[2]
		channelID := keySplit[4]

		sequence, err := strconv.ParseUint(keySplit[len(keySplit)-1], 10, 64)
		if err != nil {
			panic(err)
		}

		if cb(portID, channelID, sequence, iterator.Value()) {
			break
		}
	}
}
