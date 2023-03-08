package textual

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-proto/any"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/anypb"
)

// Any messages are rendered as a one-screen header of "Object: <type_url>"
// followed by an indented rendering of the contained message.

// anyValueRenderer is a ValueRenderer for google.protobuf.Any messages.
type anyValueRenderer struct {
	tr *SignModeHandler
}

// NewAnyValueRenderer returns a ValueRenderer for google.protobuf.Any messages.
func NewAnyValueRenderer(t *SignModeHandler) ValueRenderer {
	return anyValueRenderer{tr: t}
}

// Format implements the ValueRenderer interface.
func (ar anyValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	msg := v.Message().Interface()
	anymsg, ok := msg.(*anypb.Any)
	if !ok {
		return nil, fmt.Errorf("expected Any, got %T", msg)
	}

	packedMsg, err := anymsg.UnmarshalNew()
	if err == protoregistry.NotFound {
		// If the proto v2 registry doesn't have this message, then we use
		// protoFiles (which can e.g. be initialized to gogo's MergedRegistry)
		// to retrieve the message descriptor, and then use dynamicpb on that
		// message descriptor to create a proto.Message
		typeURL := strings.TrimPrefix(anymsg.TypeUrl, "/")

		msgDesc, err := ar.tr.protoFiles.FindDescriptorByName(protoreflect.FullName(typeURL))
		if err != nil {
			return nil, fmt.Errorf("textual protoFiles does not have descriptor %s: %w", anymsg.TypeUrl, err)
		}

		typ := dynamicpb.NewMessageType(msgDesc.(protoreflect.MessageDescriptor))
		msg := typ.New().Interface()
		err = anymsg.UnmarshalTo(msg)
		if err != nil {
			return nil, fmt.Errorf("cannot unmarshal msg %s into dynamicpb generated type: %w", anymsg.TypeUrl, err)
		}

		packedMsg = msg
	} else if err != nil {
		return nil, fmt.Errorf("error unmarshalling any %s: %w", anymsg.TypeUrl, err)
	}

	vr, err := ar.tr.GetMessageValueRenderer(packedMsg.ProtoReflect().Descriptor())
	if err != nil {
		return nil, err
	}

	subscreens, err := vr.Format(ctx, protoreflect.ValueOfMessage(packedMsg.ProtoReflect()))
	if err != nil {
		return nil, err
	}

	screens := make([]Screen, 1+len(subscreens))
	screens[0].Content = anymsg.GetTypeUrl()
	for i, subscreen := range subscreens {
		subscreen.Indent++
		screens[i+1] = subscreen
	}

	return screens, nil
}

// Parse implements the ValueRenderer interface.
func (ar anyValueRenderer) Parse(ctx context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) == 0 {
		return nilValue, fmt.Errorf("expect at least one screen")
	}
	if screens[0].Indent != 0 {
		return nilValue, fmt.Errorf("bad indentation: want 0, got %d", screens[0].Indent)
	}

	typeURL := screens[0].Content
	msgType, err := ar.tr.protoTypes.FindMessageByURL(typeURL)
	if err == protoregistry.NotFound {
		// If the proto v2 registry doesn't have this message, then we use
		// protoFiles (which can e.g. be initialized to gogo's MergedRegistry)
		// to retrieve the message descriptor, and then use dynamicpb on that
		// message descriptor to create a proto.Message
		typeURL := strings.TrimPrefix(typeURL, "/")

		msgDesc, err := ar.tr.protoFiles.FindDescriptorByName(protoreflect.FullName(typeURL))
		if err != nil {
			return nilValue, fmt.Errorf("textual protoFiles does not have descriptor %s: %w", typeURL, err)
		}

		msgType = dynamicpb.NewMessageType(msgDesc.(protoreflect.MessageDescriptor))
	} else if err != nil {
		return nilValue, err
	}
	vr, err := ar.tr.GetMessageValueRenderer(msgType.Descriptor())
	if err != nil {
		return nilValue, err
	}

	subscreens := make([]Screen, len(screens)-1)
	for i := 1; i < len(screens); i++ {
		if screens[i].Indent < 1 {
			return nilValue, fmt.Errorf("bad indent for subscreen %d: %d", i, screens[i].Indent)
		}
		subscreens[i-1] = screens[i]
		subscreens[i-1].Indent--
	}

	packedMsg, err := vr.Parse(ctx, subscreens)
	if err != nil {
		return nilValue, err
	}

	anyMsg, err := any.New(packedMsg.Message().Interface())
	if err != nil {
		return nilValue, err
	}

	return protoreflect.ValueOfMessage(anyMsg.ProtoReflect()), nil
}
