package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
)

// SubModuleCdc defines the IBC client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the IBC client interfaces and types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.ConsensusState)(nil), nil)
	cdc.RegisterInterface((*exported.Committer)(nil), nil)
	cdc.RegisterInterface((*exported.Header)(nil), nil)
	cdc.RegisterInterface((*exported.Misbehaviour)(nil), nil)

	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/MsgUpdateClient", nil)

	cdc.RegisterConcrete(tendermint.ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(tendermint.Committer{}, "ibc/client/tendermint/Committer", nil)
	cdc.RegisterConcrete(tendermint.Header{}, "ibc/client/tendermint/Header", nil)
	cdc.RegisterConcrete(tendermint.Evidence{}, "ibc/client/tendermint/Evidence", nil)

	SetSubModuleCodec(cdc)
}

func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
