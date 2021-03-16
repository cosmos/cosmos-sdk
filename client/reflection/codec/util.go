package codec

import "google.golang.org/protobuf/proto"

// MessageName is a utility function to return the message name
// similar to the deprecated proto.MessageName function.
// If proto.Message is nil then an empty string is returned.
func MessageName(pb proto.Message) string {
	if pb == nil {
		return ""
	}
	return (string)(pb.ProtoReflect().Descriptor().FullName())
}
