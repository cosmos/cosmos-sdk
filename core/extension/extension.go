package extension

import "github.com/gogo/protobuf/proto"

// Resolver is a type used to resolve interface instances for protobuf
// types which are registered as extension handlers rather than interface
// methods to avoid tight coupling of API and implementation.
type Resolver interface {
	// Register must take either:
	// * a function which takes a message and returns a handler, ex:
	//   r.Register(func (msg *MsgSend) validate.BasicValidator { return msgSendValidator{msg} })
	// * or a struct which implements the handler and has the message type embedded, ex:
	//   type msgSendValidator { *MsgSend }
	//   r.Register(msgSendValidator{})
	Register(handlerFactory interface{})

	// Resolve takes a messages and a pointer to a handler type and returns
	// an error if a handler wasn't found. Ex:
	//   var validator validate.BasicValidator
	//   err := r.Resolve(msgSend, &validator)
	Resolve(message proto.Message, handler interface{}) error
}
