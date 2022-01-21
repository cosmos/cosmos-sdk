package types

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/jsonpb"
	gogoproto "github.com/gogo/protobuf/proto"
	gogotypes "github.com/gogo/protobuf/types"
	"google.golang.org/protobuf/encoding/protojson"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func (any *Any) MarshalJSONPB(m *jsonpb.Marshaler) ([]byte, error) {
	msg, err := m.AnyResolver.Resolve(any.TypeUrl)
	if err != nil {
		return nil, err
	}

	switch msg.(type) {
	case protov2.Message:
		any2 := anypb.Any{
			TypeUrl: any.TypeUrl,
			Value:   any.Value,
		}
		marshalv2 := protojson.MarshalOptions{
			Multiline:       false,
			Indent:          "",
			AllowPartial:    false,
			UseProtoNames:   false,
			UseEnumNumbers:  false,
			EmitUnpopulated: false,
			Resolver:        m.AnyResolver.(InterfaceRegistry),
		}
		return marshalv2.Marshal(&any2)
	case gogoproto.Message:
		gogoany := gogotypes.Any{
			TypeUrl:              any.TypeUrl,
			Value:                any.Value,
			XXX_NoUnkeyedLiteral: any.XXX_NoUnkeyedLiteral,
			XXX_unrecognized:     any.XXX_unrecognized,
			XXX_sizecache:        any.XXX_sizecache,
		}
		buf := new(bytes.Buffer)
		err := m.Marshal(buf, &gogoany)
		return buf.Bytes(), err
	default:
		return nil, fmt.Errorf("the message resolved from the Any was not a gogoproto nor a protov2 message, got: %T", msg)
	}
}

func (any *Any) UnmarshalJSONPB(u *jsonpb.Unmarshaler, bz []byte) error {
	typeURL, jsonbz, err := typeUrlFromBytes(bz)
	if err != nil {
		return err
	}

	msg, err := u.AnyResolver.Resolve(typeURL)
	if err != nil {
		return err
	}

	switch msg := msg.(type) {
	case protov2.Message:
		unmarshalv2 := protojson.UnmarshalOptions{
			AllowPartial:   false,
			DiscardUnknown: !u.AllowUnknownFields, // TODO(tyler): is this ok?
			Resolver:       u.AnyResolver.(InterfaceRegistry),
		}
		err := unmarshalv2.Unmarshal(jsonbz, msg)
		if err != nil {
			return err
		}
		v2bz, err := protov2.Marshal(msg)
		if err != nil {
			return err
		}
		any.TypeUrl = typeURL
		any.Value = v2bz
		any.cachedValue = msg
		return nil
	case gogoproto.Message:
		buf := bytes.NewReader(jsonbz)
		err := u.Unmarshal(buf, msg)
		if err != nil {
			return err
		}
		gogobz, err := gogoproto.Marshal(msg)
		if err != nil {
			return err
		}
		any.TypeUrl = typeURL
		any.Value = gogobz
		any.cachedValue = msg
		return nil
	default:
		return fmt.Errorf("the message resolved from the Any was not a gogoproto nor a protov2 message, got: %T", msg)
	}
}

func typeUrlFromBytes(bz []byte) (typeUrl string, jsonbz []byte, err error) {
	// we need to extract the typeURL from the bytes in order to correctly decide
	// if this is a gogo message or a proto v2 message
	var objmap map[string]json.RawMessage
	err = json.Unmarshal(bz, &objmap)
	if err != nil {
		return "", nil, err
	}

	raw, ok := objmap["@type"]
	if !ok {
		return "", nil, errors.New("field @type not found in message bytes")
	}

	delete(objmap, "@type")
	jsonbz, err = json.Marshal(objmap)

	var typeURL string
	err = json.Unmarshal(raw, &typeURL)
	return typeURL, jsonbz, err
}
