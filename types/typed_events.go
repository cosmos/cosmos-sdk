package types

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/codec"
)

// TypedEventToEvent takes typed event and converts to Event object
func TypedEventToEvent(tev proto.Message) (Event, error) {
	if msgv2, ok := tev.(protov2.Message); ok {
		return protoreflectMessageToEvent(msgv2.ProtoReflect())
	}

	evtType := proto.MessageName(tev)
	evtJSON, err := codec.ProtoMarshalJSON(tev, nil)
	if err != nil {
		return Event{}, err
	}

	var attrMap map[string]json.RawMessage
	err = json.Unmarshal(evtJSON, &attrMap)
	if err != nil {
		return Event{}, err
	}

	// sort the keys to ensure the order is always the same
	keys := maps.Keys(attrMap)
	slices.Sort(keys)

	attrs := make([]abci.EventAttribute, 0, len(attrMap))
	for _, k := range keys {
		v := attrMap[k]
		attrs = append(attrs, abci.EventAttribute{
			Key:   k,
			Value: string(v),
		})
	}

	return Event{
		Type:       evtType,
		Attributes: attrs,
	}, nil
}

func protoreflectMessageToEvent(msg protoreflect.Message) (Event, error) {
	messageDescriptor := msg.Descriptor()
	fields := messageDescriptor.Fields()
	numFields := fields.Len()
	allFields := make([]protoreflect.FieldDescriptor, numFields)
	for i := 0; i < numFields; i++ {
		allFields[i] = fields.Get(i)
	}

	sort.Slice(allFields, func(i, j int) bool {
		return allFields[i].Number() < allFields[j].Number()
	})

	attrs := make([]abci.EventAttribute, 0, numFields)
	for _, field := range allFields {
		var err error
		value, hasValue, err := protoreflectValueToEventAttrValue(msg, field)
		if err != nil {
			return Event{}, err
		}

		if !hasValue {
			continue
		}

		attrs = append(attrs, abci.EventAttribute{
			Key:   string(field.Name()),
			Value: value,
		})
	}

	return Event{
		Type:       string(messageDescriptor.FullName()),
		Attributes: attrs,
	}, nil
}

func protoreflectValueToEventAttrValue(message protoreflect.Message, field protoreflect.FieldDescriptor) (attrValue string, hasValue bool, err error) {
	if !message.Has(field) {
		return "", false, nil
	}

	value := message.Get(field)

	if field.IsList() {
		list := value.List()
		n := list.Len()
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(json.Delim('['))
		if err != nil {
			return "", true, err
		}

		for i := 0; i < n; i++ {
			elem := list.Get(i)
			elemStr, err := protoreflectValueToEventAttrString(field, elem)
			if err != nil {
				return "", true, err
			}

			err = enc.Encode(elemStr)
			if err != nil {
				return "", true, err
			}
		}

		err = enc.Encode(json.Delim(']'))
		if err != nil {
			return "", true, err
		}

		return
	}

	attrValue, err = protoreflectValueToEventAttrString(field, value)
	return attrValue, true, err
}

func protoreflectValueToEventAttrString(field protoreflect.FieldDescriptor, value protoreflect.Value) (string, error) {
	switch field.Kind() {
	case protoreflect.StringKind:
		return value.String(), nil
	case protoreflect.BytesKind:
		return base64.URLEncoding.EncodeToString(value.Bytes()), nil
	case protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind:
		return strconv.FormatUint(value.Uint(), 10), nil
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.Int64Kind,
		protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return strconv.FormatInt(value.Int(), 10), nil
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return fmt.Sprintf("%f", value.Float()), nil
	case protoreflect.BoolKind:
		return strconv.FormatBool(value.Bool()), nil
	case protoreflect.EnumKind:
		return string(field.Enum().Values().ByNumber(value.Enum()).Name()), nil
	case protoreflect.MessageKind:
		// NOTE: this is not deterministic and will need to be replaced with a determinstic JSON serializer
		bz, err := protojson.MarshalOptions{UseProtoNames: true}.Marshal(value.Message().Interface())
		return string(bz), err
	default:
		return "", fmt.Errorf("unexpected field kind %s", field.Kind())
	}
}

// ParseTypedEvent converts abci.Event back to typed event
func ParseTypedEvent(event abci.Event) (proto.Message, error) {
	concreteGoType := proto.MessageType(event.Type)
	if concreteGoType == nil {
		return nil, fmt.Errorf("failed to retrieve the message of type %q", event.Type)
	}

	var value reflect.Value
	if concreteGoType.Kind() == reflect.Ptr {
		value = reflect.New(concreteGoType.Elem())
	} else {
		value = reflect.Zero(concreteGoType)
	}

	protoMsg, ok := value.Interface().(proto.Message)
	if !ok {
		return nil, fmt.Errorf("%q does not implement proto.Message", event.Type)
	}

	attrMap := make(map[string]json.RawMessage)
	for _, attr := range event.Attributes {
		attrMap[attr.Key] = json.RawMessage(attr.Value)
	}

	attrBytes, err := json.Marshal(attrMap)
	if err != nil {
		return nil, err
	}

	err = jsonpb.Unmarshal(strings.NewReader(string(attrBytes)), protoMsg)
	if err != nil {
		return nil, err
	}

	return protoMsg, nil
}
