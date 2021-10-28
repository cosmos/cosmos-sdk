package plugin

import "github.com/gogo/protobuf/proto"

type Host interface {
	// Expect tells the host to expect that every registered protobuf type which
	// implements the protobuf interface name (cosmos_proto.implements_interface)
	// is expected to register a golang interface implementation of type golangInterfaceType.
	Expect(protoInterfaceName string, golangInterfaceType interface{})

	// Register registers plugin handler constructors which are functions taking a plugin target
	// and returning a handler type. Ex: func(secp256k1.PubKey) crypto.PubKey.
	Register(handlerConstructors ...interface{})

	// Resolve resolves the handler for the provided target. Ex:
	//  var pubKey crypto.PubKey
	//  host.Resolve(&secp256k1.PubKey{}, &pubKey)
	Resolve(target proto.Message, handler interface{}) error
}
