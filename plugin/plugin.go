package plugin

import "github.com/gogo/protobuf/proto"

type Host interface {
	Expect(protoInterfaceName string, golangInterfaceType interface{})
	Register(handlers ...interface{})
	Resolve(target proto.Message, plugin interface{}) error
}
