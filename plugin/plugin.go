package plugin

import "github.com/gogo/protobuf/proto"

type Host interface {
	// Expect tells the host to expect that every registered protobuf type which
	// implements the protobuf interface name (cosmos_proto.implements_interface)
	// is expected to register a golang interface implementation of type golangInterfaceType.
	Expect(protoInterfaceName string, golangInterfaceType interface{})
	Register(handlers ...interface{})
	Resolve(target proto.Message, plugin interface{}) error
}
