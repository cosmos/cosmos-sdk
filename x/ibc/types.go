package ibc

import (
	"fmt"

	"encoding/json"

	"github.com/tendermint/iavl"
	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/lib"
)

// ----------------------------------
// PacketProof

type PacketProof struct {
	Proof    *iavl.KeyExistsProof
	Height   int64
	Sequence int64
}

func (proof PacketProof) Verify(ctx sdk.Context, keeper Keeper, space string, packet Packet) sdk.Error {
	chainID := packet.SrcChain

	expected := keeper.getIngressSequence(ctx, chainID)
	if proof.Sequence != expected {
		return ErrInvalidSequence()
	}

	keeper.setIngressSequence(ctx, chainID, proof.Sequence+1)
	/*
		commit, ok := keeper.getChannelCommit(ctx, chainID, proof.Height)
		if !ok {
			return ErrNoCommitFound()
		}

		key := []byte(fmt.Sprintf("ibc/%s", EgressKey(ctx.ChainID(), proof.Sequence)))
		value, rawerr := keeper.cdc.MarshalBinary(packet) // better way to do this?
		if rawerr != nil {
			return ErrInvalidPacket(rawerr)
		}

		if rawerr = proof.Proof.Verify(key, value, commit.Commit.Header.AppHash); rawerr != nil {
			return ErrInvalidPacket(rawerr)
		}
	*/
	return nil
}

// ---------------------------------
// CleanupProof

type CleanupProof struct {
	Proof  *iavl.KeyExistsProof
	Height int64
}

func (proof CleanupProof) Verify(ctx sdk.Context, q lib.QueueMapper, seq uint) sdk.Error {
	chainID := packet.SrcChain

	info := q.Info(ctx)
	if info.End <= seq || seq < info.Begin {
		return ErrInvalidSequence()
	}

	/*

	 */

	return nil
}

// ---------------------------------
// ReceiveMsg

// ReceiveMsg defines the message that a relayer uses to post an IBCPacket
// to the destination chain.

type ReceiveMsg struct {
	Packet
	PacketProof
	Relayer sdk.Address
}

func (msg ReceiveMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiveMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}

func (msg ReceiveMsg) Verify(ctx sdk.Context, keeper Keeper) sdk.Error {
	return msg.PacketProof.Verify(ctx, keeper, msg.Packet)
}

// --------------------------------
// ReceiveCleanupMsg

type ReceiveCleanupMsg struct {
	Sequence int64
	CleanupProof
	Cleaner sdk.Address
}

func (msg ReceiveCleanupMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiveCleanupMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiveCleanupMsg) GetSigners() []byte {
	return sdk.Address
}

// --------------------------------
// ReceiptMsg

type ReceiptMsg struct {
	Packet
	PacketProof
	Relayer sdk.Address
}

func (msg ReceiptMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg ReceiptMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg ReceiptMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
}

func (msg ReceiptMsg) Verify(ctx sdk.Context, keeper Keeper) sdk.Error {
	return msg.PacketProof.Verify(ctx, keeper, msg.Packet)
}

//-------------------------------------
// OpenChannelMsg

// OpenChannelMsg defines the message that is used for open a channel
// that receives msg from another chain
type OpenChannelMsg struct {
	ROT      lite.FullCommit
	SrcChain string
	Signer   sdk.Address
}

func (msg OpenChannelMsg) Type() string {
	return "ibc"
}

func (msg OpenChannelMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg OpenChannelMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg OpenChannelMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg OpenChannelMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer}
}

//------------------------------------
// UpdateChannelMsg

type UpdateChannelMsg struct {
	SrcChain string
	Commit   lite.FullCommit
	//PacketProof
	Signer sdk.Address
}

func (msg UpdateChannelMsg) Type() string {
	return "ibc"
}

func (msg UpdateChannelMsg) Get(key interface{}) interface{} {
	return nil
}

func (msg UpdateChannelMsg) GetSignBytes() []byte {
	bz, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return bz
}

func (msg UpdateChannelMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg UpdateChannelMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer}
}

// ------------------------------
// Payload
// Payload defines inter-blockchain message
// that can be proved by light-client protocol

type Payload interface {
	Type() string
	ValidateBasic() sdk.Error
}

// ------------------------------
// Packet

// Packet defines a piece of data that can be send between two separate
// blockchains.
type Packet struct {
	Payload
	SrcChain  string
	DestChain string
}
