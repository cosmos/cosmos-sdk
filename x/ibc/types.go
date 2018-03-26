package ibc

import (
	"encoding/json"

	"github.com/tendermint/iavl"

	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// ----------------------------------
// ReceiveMsg

// ReceiveMsg defines the message that a relayer uses to post an IBCPacket
// to the destination chain.
type ReceiveMsg struct {
	types.Packet
	Proof    *iavl.KeyExistsProof
	Relayer  sdk.Address
	Sequence int64
}

func (msg ReceiveMsg) Type() string {
	return "ibc"
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

func (msg ReceiveMsg) ValidateBasic() sdk.Error {
	return msg.Packet.ValidateBasic()
}

// x/bank/tx.go SendMsg.GetSigners()
func (msg ReceiveMsg) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Relayer}
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
	Signer   sdk.Address
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
