package client

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ConsensusState)(nil), nil)
	cdc.RegisterInterface((*Header)(nil), nil)
}
