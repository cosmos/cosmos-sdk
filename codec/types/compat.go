package types

import (
	gogoany "github.com/cosmos/gogoproto/types/any"
)

// Deprecated: this is no longer used for anything.
var Debug = true

// AminoUnpacker is an alias for github.com/cosmos/gogoproto/types/any.AminoUnpacker.
type AminoUnpacker = gogoany.AminoUnpacker

// AminoPacker is an alias for github.com/cosmos/gogoproto/types/any.AminoPacker.
type AminoPacker = gogoany.AminoPacker

// AminoJSONUnpacker is an alias for github.com/cosmos/gogoproto/types/any.AminoJSONUnpacker.
type AminoJSONUnpacker = gogoany.AminoJSONUnpacker

// AminoJSONPacker is an alias for github.com/cosmos/gogoproto/types/any.AminoJSONPacker.
type AminoJSONPacker = gogoany.AminoJSONPacker

// ProtoUnpacker is an alias for github.com/cosmos/gogoproto/types/any.ProtoJSONPacker.
type ProtoJSONPacker = gogoany.ProtoJSONPacker
