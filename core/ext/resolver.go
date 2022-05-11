package ext

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Resolver[I any] struct {
	handlerMap map[protoreflect.FullName]*handler[I]
}

func NewResolver[I any](handlers map[string]*HandlerSet[I]) (*Resolver[I], error) {
	handlerMap := map[protoreflect.FullName]*handler[I]{}
	var i I
	for module, handlerSet := range handlers {
		for _, h := range handlerSet.handlers {
			fullName := h.msg.ProtoReflect().Descriptor().FullName()
			if existing, ok := handlerMap[fullName]; ok {
				return nil, fmt.Errorf("duplicate registration handler of %s for interface %T in modules %s and %s",
					fullName, i, module, existing.module)
			}
			h.module = module
			handlerMap[fullName] = h
		}
	}

	return &Resolver[I]{handlerMap: handlerMap}, nil
}

func (r Resolver[I]) Resolve(a anypb.Any) (I, error) {
	url := a.TypeUrl
	fullName := protoreflect.FullName(url)
	if i := strings.LastIndexByte(url, '/'); i >= 0 {
		fullName = fullName[i+len("/"):]
	}

	h, ok := r.handlerMap[fullName]
	var i I
	if !ok {
		return nil, fmt.Errorf("handler %T not found for %s", i, fullName)
	}

	msg := h.msg.ProtoReflect().New().Interface()
	err := proto.Unmarshal(a.Value, msg)
	if err != nil {
		return nil, err
	}

	return h.f(msg), nil
}
