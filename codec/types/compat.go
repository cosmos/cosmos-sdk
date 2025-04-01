package types

import (
	gogoany "github.com/cosmos/gogoproto/types/any"
)

var Debug = true

type AminoUnpacker = gogoany.AminoUnpacker

type AminoPacker = gogoany.AminoPacker

type AminoJSONUnpacker = gogoany.AminoJSONUnpacker

type AminoJSONPacker = gogoany.AminoJSONPacker

type ProtoJSONPacker = gogoany.ProtoJSONPacker
