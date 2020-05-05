package types

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
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

	// nolint
	TypeUrl string `protobuf:"bytes,1,opt,name=type_url,json=typeUrl,proto3" json:"type_url,omitempty"`
	// Must be a valid serialized protocol buffer of the above specified type.
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`

	// nolint
	XXX_NoUnkeyedLiteral struct{} `json:"-"`

	// nolint
	XXX_unrecognized []byte `json:"-"`

	// nolint
	XXX_sizecache int32 `json:"-"`

	cachedValue interface{}
}

// NewAnyWithValue constructs a new Any packed with the value provided or
// returns an error if that value couldn't be packed. This also caches
// the packed value so that it can be retrieved from GetCachedValue without
// unmarshaling
func NewAnyWithValue(value proto.Message) (*Any, error) {
	any := &Any{}

	err := any.Pack(value)
	if err != nil {
		return nil, err
	}

	return any, nil
}

// Pack packs the value x in the Any or returns an error. This also caches
// the packed value so that it can be retrieved from GetCachedValue without
// unmarshaling
func (any *Any) Pack(x proto.Message) error {
	any.TypeUrl = "/" + proto.MessageName(x)
	bz, err := proto.Marshal(x)
	if err != nil {
		return err
	}

	any.Value = bz
	any.cachedValue = x

	return nil
}

// GetCachedValue returns the cached value from the Any if present
func (any *Any) GetCachedValue() interface{} {
	return any.cachedValue
}

// ClearCachedValue clears the cached value from the Any
func (any *Any) ClearCachedValue() {
	any.cachedValue = nil
}

// MarshalAny is a convenience function for packing the provided value in an
// Any and then proto marshaling it to bytes
func MarshalAny(x interface{}) ([]byte, error) {
	msg, ok := x.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("can't proto marshal %T", x)
	}

	any := &Any{}
	err := any.Pack(msg)
	if err != nil {
		return nil, err
	}

	return any.Marshal()
}

// UnmarshalAny is a convenience function for proto unmarshaling an Any from
// bz and then unpacking it to the interface pointer passed in as iface using
// the provided AnyUnpacker or returning an error
//
// Ex:
//		var x MyInterface
//		err := UnmarshalAny(unpacker, &x, bz)
func UnmarshalAny(unpacker AnyUnpacker, iface interface{}, bz []byte) error {
	any := &Any{}

	err := any.Unmarshal(bz)
	if err != nil {
		return err
	}

	return unpacker.UnpackAny(any, iface)
}
