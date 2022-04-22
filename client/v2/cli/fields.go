package cli

import (
	"context"

	"github.com/spf13/pflag"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func (b *Builder) registerMessageFlagSet(ctx context.Context, set *pflag.FlagSet, messageType protoreflect.MessageType) *messageFlagHandler {
	fields := messageType.Descriptor().Fields()
	numFields := fields.Len()
	handler := &messageFlagHandler{
		messageType: messageType,
	}
	for i := 0; i < numFields; i++ {
		field := fields.Get(i)
		typ := b.getFlagType(field)
		if typ == nil {
			// TODO get rid of this
			continue
		}
		val := typ.AddFlag(ctx, b, set, field)
		handler.flagFieldPairs = append(handler.flagFieldPairs, struct {
			value FlagValue
			field protoreflect.FieldDescriptor
		}{value: val, field: field})
	}
	return handler
}

type messageFlagHandler struct {
	flagFieldPairs []struct {
		value FlagValue
		field protoreflect.FieldDescriptor
	}
	messageType protoreflect.MessageType
}

func (m messageFlagHandler) buildMessage() protoreflect.Message {
	msg := m.messageType.New()
	for _, pair := range m.flagFieldPairs {
		msg.Set(pair.field, pair.value.Get())
	}
	return msg
}
