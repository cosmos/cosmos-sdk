package ext

import (
	"google.golang.org/protobuf/proto"
)

type HandlerSet[I any] struct {
	handlers []*handler[I]
}

func (h *HandlerSet[I]) IsOnePerModuleType() {}

type handler[I any] struct {
	msg    proto.Message
	f      func(message proto.Message) I
	module string
}

func Register[I any, M proto.Message](handlerSet *HandlerSet[I], msg M, mkHandler func(M) I) {
	handlerSet.handlers = append(handlerSet.handlers, &handler[I]{
		msg: msg,
		f: func(message proto.Message) I {
			return mkHandler(message.(M))
		},
	})
}
