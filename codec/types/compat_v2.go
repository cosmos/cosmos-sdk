package types

import (
	"bytes"
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
	msg, err := u.AnyResolver.Resolve(any.TypeUrl)
	if err != nil {
		return err
	}

	switch msg := msg.(type) {
	case protov2.Message:
		unmarshalv2 := protojson.UnmarshalOptions{
			AllowPartial:   false,
			DiscardUnknown: !u.AllowUnknownFields, // TODO(tyler): this ok?
			Resolver:       u.AnyResolver.(InterfaceRegistry),
		}
		err := unmarshalv2.Unmarshal(bz, msg)
		if err != nil {
			return err
		}
		any.Value = bz
		return nil
	case gogoproto.Message:
		buf := bytes.NewReader(bz)
		err := u.Unmarshal(buf, msg)
		if err != nil {
			return err
		}
		any.Value = bz
		return nil
	default:
		return fmt.Errorf("the message resolved from the Any was not a gogoproto nor a protov2 message, got: %T", msg)
	}
}
