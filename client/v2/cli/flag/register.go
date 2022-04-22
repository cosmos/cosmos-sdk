package flag

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (b *Builder) RegisterMessageFlags(ctx context.Context, set *pflag.FlagSet, messageType protoreflect.MessageType, options Options) *MessageBinder {
	fields := messageType.Descriptor().Fields()
	numFields := fields.Len()
	handler := &MessageBinder{
		messageType: messageType,
	}
	for i := 0; i < numFields; i++ {
		field := fields.Get(i)
		binder := b.BindFieldFlag(ctx, set, field, options)
		if binder == nil {
			fmt.Printf("unable to bind field %s to a flag, support will be added soon\n", field)
		}
		handler.flagFieldPairs = append(handler.flagFieldPairs, struct {
			binder FieldBinder
			field  protoreflect.FieldDescriptor
		}{binder: binder, field: field})
	}
	return handler
}

type MessageBinder struct {
	flagFieldPairs []struct {
		binder FieldBinder
		field  protoreflect.FieldDescriptor
	}
	messageType protoreflect.MessageType
}

func (m MessageBinder) BuildMessage() protoreflect.Message {
	msg := m.messageType.New()
	m.Bind(msg)
	return msg
}

func (m MessageBinder) Bind(msg protoreflect.Message) {
	for _, pair := range m.flagFieldPairs {
		pair.binder.Bind(msg, pair.field)
	}
}

func (m MessageBinder) Get() protoreflect.Value {
	return protoreflect.ValueOfMessage(m.BuildMessage())
}
