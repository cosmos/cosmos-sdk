package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

// SubModuleCdc defines the IBC tendermint client codec.
var SubModuleCdc *codec.Codec

func init() {
	SubModuleCdc = codec.New()
	cryptocodec.RegisterCrypto(SubModuleCdc)
	RegisterCodec(SubModuleCdc)
}

// RegisterCodec registers the Tendermint types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/tendermint/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/tendermint/Evidence", nil)
	cdc.RegisterConcrete(&MsgCreateClient{}, "ibc/client/tendermint/MsgCreateClient", nil)
	cdc.RegisterConcrete(&MsgUpdateClient{}, "ibc/client/tendermint/MsgUpdateClient", nil)
	cdc.RegisterConcrete(&MsgSubmitClientMisbehaviour{}, "ibc/client/tendermint/MsgSubmitClientMisbehaviour", nil)
}
