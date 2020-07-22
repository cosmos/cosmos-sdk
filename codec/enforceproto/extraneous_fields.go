package enforceproto

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"sync"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
	"google.golang.org/protobuf/encoding/protowire"

	"github.com/cosmos/cosmos-sdk/codec/types"
)

const bit11NonCritical = 1 << 10

type descriptorIface interface {
	Descriptor() ([]byte, []int)
}

type protoMessageWithDescriptor struct {
	descriptorIface
	gogoproto.Message
}

var _ descriptor.Message = (*protoMessageWithDescriptor)(nil)

type descriptorMatch struct {
	cache map[int32]*descriptor.FieldDescriptorProto
	desc  *descriptor.DescriptorProto
}

// CheckMismatchedProtoFields walks through the protobuf serialized bytes in b, and tries to
// compare field numbers and wireTypes against what msg expects. The error returned if non-nil will contain
// the listing of the extraneous fields by tagNumber and wireType, as well as mismatched
// wireTypes.
func CheckMismatchedProtoFields(b []byte, msg gogoproto.Message) error {
	if len(b) == 0 {
		return nil
	}

	desc, ok := msg.(descriptorIface)
	if !ok {
		return fmt.Errorf("%T does not have a Descriptor() method", msg)
	}

	fieldDescProtoFromTagNum, _, err := descProtoCache(desc, msg)
	if err != nil {
		return err
	}

	for len(b) > 0 {
		tagNum, wireType, n := protowire.ConsumeField(b)
		if n < 0 {
			return errors.New("invalid length")
		}

		fieldDescProto, ok := fieldDescProtoFromTagNum[int32(tagNum)]
		switch {
		case ok:
			// Assert that the wireTypes match.
			if !canEncodeType(wireType, fieldDescProto.GetType()) {
				return &ErrMismatchedWireType{
					Type:         reflect.ValueOf(msg).Type().String(),
					TagNum:       tagNum,
					GotWireType:  wireType,
					WantWireType: protowire.Type(fieldDescProto.WireType()),
				}
			}

		default:
			if tagNum&bit11NonCritical == 0 {
				// The tag is non-critical, so report it.
				return &UnexpectedField{
					Type:     reflect.ValueOf(msg).Type().String(),
					TagNum:   tagNum,
					WireType: wireType,
				}
			}
		}

		// Skip over the 2 bytes that store fieldNumber and wireType bytes.
		fieldBytes := b[2:n]
		b = b[n:]

		// An unknown but non-critical field or just a scalar type (aka *INT and BYTES like).
		if fieldDescProto == nil || fieldDescProto.IsScalar() {
			continue
		}

		protoMessageName := fieldDescProto.GetTypeName()
		if protoMessageName == "" {
			// At this point only TYPE_STRING is expected to be unregistered, since FieldDescriptorProto.IsScalar() returns false for TYPE_STRING
			// per https://github.com/gogo/protobuf/blob/5628607bb4c51c3157aacc3a50f0ab707582b805/protoc-gen-gogo/descriptor/descriptor.go#L95-L118
			if typ := fieldDescProto.GetType(); typ != descriptor.FieldDescriptorProto_TYPE_STRING {
				return fmt.Errorf("failed to get typename for message of type %d, can only be TYPE_STRING", typ)
			}
			continue
		}

		// Let's recursively traverse and typecheck the field.

		if protoMessageName == ".google.protobuf.Any" {
			// We'll need to extract the TypeURL which will contain the protoMessageName.
			any := new(types.Any)
			if err := gogoproto.Unmarshal(fieldBytes, any); err != nil {
				return err
			}
			protoMessageName = any.TypeUrl
		}

		msg, err := protoMessageForTypeName(protoMessageName[1:])
		if err != nil {
			return err
		}
		if err := CheckMismatchedProtoFields(fieldBytes, msg); err != nil {
			return err
		}
	}

	return nil
}

var protoMessageForTypeNameMu sync.RWMutex
var protoMessageForTypeNameCache = make(map[string]gogoproto.Message)

// protoMessageForTypeName takes in a fully qualified name e.g. testdata.TestVersionFD1
// and returns a corresponding empty protobuf message that serves the prototype for typechecking.
func protoMessageForTypeName(protoMessageName string) (gogoproto.Message, error) {
	protoMessageForTypeNameMu.RLock()
	msg, ok := protoMessageForTypeNameCache[protoMessageName]
	protoMessageForTypeNameMu.RUnlock()
	if ok {
		return msg, nil
	}

	concreteGoType := gogoproto.MessageType(protoMessageName)
	if concreteGoType == nil {
		return nil, fmt.Errorf("failed to retrieve the message of type %q", protoMessageName)
	}

	value := reflect.New(concreteGoType).Elem()
	msg, ok = value.Interface().(gogoproto.Message)
	if !ok {
		return nil, fmt.Errorf("%q does not implement proto.Message", protoMessageName)
	}

	// Now cache it.
	protoMessageForTypeNameMu.Lock()
	protoMessageForTypeNameCache[protoMessageName] = msg
	protoMessageForTypeNameMu.Unlock()

	return msg, nil
}

// canEncodeType returns true if the wireType is suitable for encoding the descriptor type.
// See https://developers.google.com/protocol-buffers/docs/encoding#structure.
func canEncodeType(wireType protowire.Type, descType descriptor.FieldDescriptorProto_Type) bool {
	switch descType {
	// "0	Varint: int32, int64, uint32, uint64, sint32, sint64, bool, enum"
	case descriptor.FieldDescriptorProto_TYPE_INT32, descriptor.FieldDescriptorProto_TYPE_INT64,
		descriptor.FieldDescriptorProto_TYPE_UINT32, descriptor.FieldDescriptorProto_TYPE_UINT64,
		descriptor.FieldDescriptorProto_TYPE_SINT32, descriptor.FieldDescriptorProto_TYPE_SINT64,
		descriptor.FieldDescriptorProto_TYPE_BOOL, descriptor.FieldDescriptorProto_TYPE_ENUM:
		return wireType == protowire.VarintType

	// "1	64-bit:	fixed64, sfixed64, double"
	case descriptor.FieldDescriptorProto_TYPE_FIXED64, descriptor.FieldDescriptorProto_TYPE_SFIXED64,
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		return wireType == protowire.Fixed64Type

	// "2	Length-delimited: string, bytes, embedded messages, packed repeated fields"
	case descriptor.FieldDescriptorProto_TYPE_STRING, descriptor.FieldDescriptorProto_TYPE_BYTES,
		descriptor.FieldDescriptorProto_TYPE_MESSAGE:
		return wireType == protowire.BytesType

	// "3	Start group:	groups (deprecated)"
	// "4	End group:	groups (deprecated)"
	case descriptor.FieldDescriptorProto_TYPE_GROUP:
		return wireType == protowire.StartGroupType || wireType == protowire.EndGroupType

	// "5	32-bit:	fixed32, sfixed32, float"
	case descriptor.FieldDescriptorProto_TYPE_FIXED32, descriptor.FieldDescriptorProto_TYPE_SFIXED32,
		descriptor.FieldDescriptorProto_TYPE_FLOAT:
		return wireType == protowire.Fixed32Type

	default:
		panic(fmt.Sprintf("Should not happen but wireType: %s cannot handle %s", wireTypeToString(wireType), descType))
	}
}

// ErrMismatchedWireType describes a mismatch between
// expected and got wireTypes for a specific tag number.
type ErrMismatchedWireType struct {
	Type         string
	GotWireType  protowire.Type
	WantWireType protowire.Type
	TagNum       protowire.Number
}

func (mwt *ErrMismatchedWireType) String() string {
	return fmt.Sprintf("Mismatched %q: {TagNum: %d, GotWireType: %q != WantWireType: %q}",
		mwt.Type, mwt.TagNum, wireTypeToString(mwt.GotWireType), wireTypeToString(mwt.WantWireType))
}

func (mwt *ErrMismatchedWireType) Error() string {
	return mwt.String()
}

var _ error = (*ErrMismatchedWireType)(nil)

func wireTypeToString(wt protowire.Type) string {
	switch wt {
	case 0:
		return "varint"
	case 1:
		return "fixed64"
	case 2:
		return "bytes"
	case 3:
		return "start_group"
	case 4:
		return "end_group"
	case 5:
		return "fixed32"
	default:
		return fmt.Sprintf("unknown type: %d", wt)
	}
}

type UnexpectedField struct {
	Type     string
	TagNum   protowire.Number
	WireType protowire.Type
}

func (twt *UnexpectedField) String() string {
	return fmt.Sprintf("UnexpectedField %q: {TagNum: %d, WireType:%q}",
		twt.Type, twt.TagNum, wireTypeToString(twt.WireType))
}

func (twt *UnexpectedField) Error() string {
	return twt.String()
}

var _ error = (*UnexpectedField)(nil)

var (
	protoFileToDesc   = make(map[string]*descriptor.FileDescriptorProto)
	protoFileToDescMu sync.RWMutex
)

// Invoking descriptor.ForMessage(gogoproto.Message.(Descriptor).Descriptor()) is incredibly slow
// for every single message, thus the need for a hand-rolled custom version that's performant and caches.
func extractFileDesMessageDesc(desc descriptorIface) (*descriptor.FileDescriptorProto, *descriptor.DescriptorProto, error) {
	gzippedPb, indices := desc.Descriptor()

	protoFileToDescMu.RLock()
	cached, ok := protoFileToDesc[string(gzippedPb)]
	protoFileToDescMu.RUnlock()

	if ok {
		mdesc := cached.MessageType[indices[0]]
		for _, index := range indices[1:] {
			mdesc = mdesc.NestedType[index]
		}
		return cached, mdesc, nil
	}

	// Time to gunzip the content of the FileDescriptor and then proto unmarshal them.
	gzr, err := gzip.NewReader(bytes.NewReader(gzippedPb))
	if err != nil {
		return nil, nil, err
	}
	protoBlob, err := ioutil.ReadAll(gzr)
	if err != nil {
		return nil, nil, err
	}

	fdesc := new(descriptor.FileDescriptorProto)
	if err := gogoproto.Unmarshal(protoBlob, fdesc); err != nil {
		return nil, nil, err
	}

	// Now cache the FileDescriptor.
	protoFileToDescMu.Lock()
	protoFileToDesc[string(gzippedPb)] = fdesc
	protoFileToDescMu.Unlock()

	// Unnest the type if necessary.
	mdesc := fdesc.MessageType[indices[0]]
	for _, index := range indices[1:] {
		mdesc = mdesc.NestedType[index]
	}
	return fdesc, mdesc, nil
}

var descprotoCacheMu sync.RWMutex
var descprotoCache = make(map[reflect.Type]*descriptorMatch)

func descProtoCache(desc descriptorIface, msg gogoproto.Message) (map[int32]*descriptor.FieldDescriptorProto, *descriptor.DescriptorProto, error) {
	key := reflect.ValueOf(msg).Type()

	descprotoCacheMu.RLock()
	got, ok := descprotoCache[key]
	descprotoCacheMu.RUnlock()

	if ok {
		return got.cache, got.desc, nil
	}

	// Now compute and cache the index.
	_, md, err := extractFileDesMessageDesc(desc)
	if err != nil {
		return nil, nil, err
	}

	tagNumToTypeIndex := make(map[int32]*descriptor.FieldDescriptorProto)
	for _, field := range md.Field {
		tagNumToTypeIndex[field.GetNumber()] = field
	}

	descprotoCacheMu.Lock()
	descprotoCache[key] = &descriptorMatch{
		cache: tagNumToTypeIndex,
		desc:  md,
	}
	descprotoCacheMu.Unlock()

	return tagNumToTypeIndex, md, nil
}
