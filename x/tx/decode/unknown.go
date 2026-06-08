package decode

import (
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/x/tx/signing"
)

// RejectUnknownFieldsStrict rejects any bytes bz that contain unknown proto fields.
// It delegates to x/tx/signing where the implementation now lives.
func RejectUnknownFieldsStrict(bz []byte, msg protoreflect.MessageDescriptor, resolver protodesc.Resolver) error {
	return signing.RejectUnknownFieldsStrict(bz, msg, resolver)
}

// RejectUnknownFields is like RejectUnknownFieldsStrict but allows non-critical unknown fields.
func RejectUnknownFields(bz []byte, desc protoreflect.MessageDescriptor, allowUnknownNonCriticals bool, resolver protodesc.Resolver) (bool, error) {
	return signing.RejectUnknownFields(bz, desc, allowUnknownNonCriticals, resolver)
}

// WireTypeToString returns a human-readable string for a protobuf wire type.
// Re-exported from x/tx/signing for backwards compatibility.
func WireTypeToString(wt protowire.Type) string {
	return signing.WireTypeToString(wt)
}
