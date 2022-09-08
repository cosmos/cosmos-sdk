package types

import (
	gogoproto "github.com/cosmos/gogoproto/types/any"
)

type Any struct {
	// A URL/resource name that uniquely identifies the type of the serialized
	// protocol buffer message. This string must contain at least
	// one "/" character. The last segment of the URL's path must represent
	// the fully qualified name of the type (as in
	// `path/google.protobuf.Duration`). The name should be in a canonical form
	// (e.g., leading "." is not accepted).
	//
	// In practice, teams usually precompile into the binary all types that they
	// expect it to use in the context of Any. However, for URLs which use the
	// scheme `http`, `https`, or no scheme, one can optionally set up a type
	// server that maps type URLs to message definitions as follows:
	//
	// * If no scheme is provided, `https` is assumed.
	// * An HTTP GET on the URL must yield a [google.protobuf.Type][]
	//   value in binary format, or produce an error.
	// * Applications are allowed to cache lookup results based on the
	//   URL, or have them precompiled into a binary to avoid any
	//   lookup. Therefore, binary compatibility needs to be preserved
	//   on changes to types. (Use versioned type names to manage
	//   breaking changes.)
	//
	// Note: this functionality is not currently available in the official
	// protobuf release, and it is not used for type URLs beginning with
	// type.googleapis.com.
	//
	// Schemes other than `http`, `https` (or the empty scheme) might be
	// used with implementation specific semantics.

	TypeUrl string `protobuf:"bytes,1,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	// Must be a valid serialized protocol buffer of the above specified type.
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`

	XXX_NoUnkeyedLiteral struct{} `json:"-"`

	XXX_unrecognized []byte `json:"-"`

	XXX_sizecache int32 `json:"-"`

	cachedValue interface{}

	compat *anyCompat
}

// NewAnyWithValue constructs a new Any packed with the value provided or
// returns an error if that value couldn't be packed. This also caches
// the packed value so that it can be retrieved from GetCachedValue without
// unmarshaling
func NewAnyWithValue(v proto.Message) (*Any, error) {
	if v == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrPackAny, "Expecting non nil value to create a new Any")
	}

	bz, err := proto.Marshal(v)
	if err != nil {
		return nil, err
	}

	return &Any{
		TypeUrl:     "/" + proto.MessageName(v),
		Value:       bz,
		cachedValue: v,
	}, nil
}

// AminoPacker is an alias for gogoproto.AminoPacker. It provides functionality
// for packing and unpacking data using the Amino encoding.
type AminoPacker = gogoproto.AminoPacker

// AminoUnpacker is an alias for gogoproto.AminoUnpacker. It is used for
// unpacking Amino-encoded data into Go types.
type AminoUnpacker = gogoproto.AminoUnpacker

// AminoJSONPacker is an alias for gogoproto.AminoJSONPacker. It allows for
// packing data into a JSON format using Amino encoding rules.
type AminoJSONPacker = gogoproto.AminoJSONPacker

// AminoJSONUnpacker is an alias for gogoproto.AminoJSONUnpacker. It provides
// the ability to unpack JSON data encoded with Amino encoding rules.
type AminoJSONUnpacker = gogoproto.AminoJSONUnpacker

// ProtoJSONPacker is an alias for gogoproto.ProtoJSONPacker. This is used for
// packing protocol buffer messages into a JSON format.
type ProtoJSONPacker = gogoproto.ProtoJSONPacker

// NewAnyWithValue is an alias for gogoproto.NewAnyWithCacheWithValue. This function
// creates a new Any instance containing the provided value, with caching
// mechanisms to improve performance.
var NewAnyWithValue = gogoproto.NewAnyWithCacheWithValue

// UnsafePackAny is an alias for gogoproto.UnsafePackAnyWithCache. This function
// packs a given message into an Any type without performing safety checks.
var UnsafePackAny = gogoproto.UnsafePackAnyWithCache
