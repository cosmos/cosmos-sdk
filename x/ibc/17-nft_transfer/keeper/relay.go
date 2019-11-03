package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channelexported "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/17-nft_transfer/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// SendTransfer handles nft_transfer sending logic
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel,
	id,
	denom string,
	sender,
	receiver sdk.AccAddress,
	isSourceChain bool,
) error {
	// get the port and channel of the counterparty
	channel, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrChannelNotFound(k.codespace, sourcePort, sourceChannel)
	}

	destinationPort := channel.Counterparty.PortID
	destinationChannel := channel.Counterparty.ChannelID

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrSequenceNotFound(k.codespace, "send")
	}

	prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
	if isSourceChain {
		// build the receiving denomination prefix
		denom = prefix + denom
	}

	return k.createOutgoingPacket(ctx, sequence, sourcePort, sourceChannel, destinationPort, destinationChannel, id, denom, sender, receiver, isSourceChain)
}

// ReceivePacket handles receiving packet
func (k Keeper) ReceivePacket(ctx sdk.Context, packet channelexported.PacketI, proof commitment.ProofI, height uint64) error {
	_, err := k.channelKeeper.RecvPacket(ctx, packet, proof, height, nil, k.storeKey)
	if err != nil {
		return err
	}

	var data types.PacketData
	err = data.UnmarshalJSON(packet.GetData())
	if err != nil {
		return sdk.NewError(types.DefaultCodespace, types.CodeInvalidPacketData, "invalid packet data")
	}

	return k.ReceiveTransfer(ctx, packet.GetSourcePort(), packet.GetSourceChannel(), packet.GetDestPort(), packet.GetDestChannel(), data)
}

// ReceiveTransfer handles nft_transfer receiving logic
func (k Keeper) ReceiveTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketData,
) error {
	if data.Source {
		prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		if !strings.HasPrefix(data.Denom, prefix) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidDenom, "incorrect denomination")
		}

		nft := NewBaseNFT(data.ID, data.Receiver, data.TokenURI)

		// mint new non-fungible token if the source of the nft_transfer is the same chain
		return k.nftKeeper.MintNFT(ctx, data.Denom, &nft)
	}

	// unescrow tokens

	// check the denom prefix
	prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
	if !strings.HasPrefix(data.Denom, prefix) {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidDenom, "incorrect denomination")
	}

	escrowAddress := types.GetEscrowAddress(destinationPort, destinationChannel)
	denom := data.Denom[len(prefix):]

	nft, err := k.nftKeeper.GetNFT(ctx, denom, data.ID)
	if err != nil {
		return err
	}

	// NFT needs to be in escrow to continue
	if !nft.GetOwner().Equals(escrowAddress) {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidNFT, "cant nft_transfer un-owned NFT")
	}

	// update NFT owner
	nft.SetOwner(data.Receiver)
	return k.nftKeeper.UpdateNFT(ctx, denom, nft)
}

func (k Keeper) createOutgoingPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel,
	id,
	denom string,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
	isSourceChain bool,
) error {

	var tokenURI string

	if isSourceChain {
		// escrow tokens if the destination chain is the same as the sender's
		escrowAddress := types.GetEscrowAddress(sourcePort, sourceChannel)

		prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		if !strings.HasPrefix(denom, prefix) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidDenom, "incorrect denomination")
		}

		denomination := denom[len(prefix):]

		nft, err := k.nftKeeper.GetNFT(ctx, denomination, id)
		if err != nil {
			return err
		}

		// NFT needs to be owned by sender to continue
		if !nft.GetOwner().Equals(sender) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidNFT, "cant nft_transfer un-owned NFT")
		}

		tokenURI = nft.GetTokenURI()

		// update NFT owner
		nft.SetOwner(escrowAddress)
		err = k.nftKeeper.UpdateNFT(ctx, denomination, nft)
		if err != nil {
			return err
		}

	} else {
		// burn vouchers from the sender's balance if the source is from another chain
		prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		if !strings.HasPrefix(denom, prefix) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidDenom, "incorrect denomination")
		}

		nft, err := k.nftKeeper.GetNFT(ctx, denom, id)
		if err != nil {
			return err
		}

		// NFT needs to be owned by sender to continue
		if !nft.GetOwner().Equals(sender) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidNFT, "cant nft_transfer un-owned NFT")
		}

		tokenURI = nft.GetTokenURI()

		err = k.nftKeeper.DeleteNFT(ctx, denom, id)
		if err != nil {
			return err
		}
	}

	packetData := types.PacketData{
		ID:       id,
		Denom:    denom,
		TokenURI: tokenURI,
		Sender:   sender,
		Receiver: receiver,
		Source:   isSourceChain,
	}

	packetDataBz, err := packetData.MarshalJSON()
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidPacketData, "invalid packet data")
	}

	packet := channeltypes.NewPacket(
		seq,
		uint64(ctx.BlockHeight())+DefaultPacketTimeout,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		packetDataBz,
	)

	// generate the capability key
	key := sdk.NewKVStoreKey(types.BoundPortID)
	return k.channelKeeper.SendPacket(ctx, packet, key)
}
