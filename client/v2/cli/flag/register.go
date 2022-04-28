package flag

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// AddMessageFlags adds flags for each field in the message to the flag set.
func (b *Builder) AddMessageFlags(ctx context.Context, set *pflag.FlagSet, messageType protoreflect.MessageType, options Options) *MessageBinder {
	fields := messageType.Descriptor().Fields()
	numFields := fields.Len()
	handler := &MessageBinder{
		messageType: messageType,
	}
	for i := 0; i < numFields; i++ {
		field := fields.Get(i)
		binder := b.AddFieldFlag(ctx, set, field, options)
		if binder == nil {
			fmt.Printf("unable to bind field %s to a flag, support will be added soon\n", field)
			continue
		}
		handler.flagFieldPairs = append(handler.flagFieldPairs, struct {
			binder FieldValueBinder
			field  protoreflect.FieldDescriptor
		}{binder: binder, field: field})
	}
	return handler
}

// MessageBinder binds multiple flags in a flag set to a protobuf message.
type MessageBinder struct {
	flagFieldPairs []struct {
		binder FieldValueBinder
		field  protoreflect.FieldDescriptor
	}
	messageType protoreflect.MessageType
}

// BuildMessage builds and returns a new message for the bound flags.
func (m MessageBinder) BuildMessage() protoreflect.Message {
	msg := m.messageType.New()
	m.Bind(msg)
	return msg
}

// Bind binds the flag values to an existing protobuf message.
func (m MessageBinder) Bind(msg protoreflect.Message) {
	for _, pair := range m.flagFieldPairs {
		pair.binder.Bind(msg, pair.field)
	}
}

// Get calls BuildMessage and wraps the result in a protoreflect.Value.
func (m MessageBinder) Get() protoreflect.Value {
	return protoreflect.ValueOfMessage(m.BuildMessage())
}
