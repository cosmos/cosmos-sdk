package kvindexer

import (
	"encoding/binary"
	"fmt"
	"github.com/cosmos/cosmos-sdk/orm/apis/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/pkg/protoutils/kindencoder"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"log"
)

// FieldKeyEncoder encodes the field key following this schema:
// KV namespace: [1]byte
// Type prefix: [2]byte
// Type index identifier (protoreflect.FieldNumber): [4]byte
// Field value length: [8]byte
// Field value bytes: [x]byte
// Primary key value: [x]byte
type FieldKeyEncoder struct {
	typePrefix       []byte
	fieldNumberBytes []byte
	fd               protoreflect.FieldDescriptor
	kindEncoder      kindencoder.KindEncoder
}

// IndexPrefix provides the bytes prefix to match all the primary keys
// associated with the provided protoreflect.Value
// NOTE: it assumes toMatch value is valid.
func (f *FieldKeyEncoder) IndexPrefix(toMatch protoreflect.Value) []byte {
	encodedField := f.kindEncoder.EncodeValueToBytes(toMatch)
	lengthBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(lengthBytes, uint64(len(encodedField)))

	indexPrefix := []byte{
		objectIndexesPrefix[0],
		f.typePrefix[0], f.typePrefix[1],
		f.fieldNumberBytes[0], f.fieldNumberBytes[1], f.fieldNumberBytes[2], f.fieldNumberBytes[3]}
	indexPrefix = append(indexPrefix, lengthBytes...)
	indexPrefix = append(indexPrefix, encodedField...)
	log.Printf("index(%s) value(%s): %v", f.fd.Name(), toMatch.String(), indexPrefix)
	return indexPrefix
}

// EncodePrimaryKey provides the mapping key for the given primary key associated
// with the provided field.
func (f *FieldKeyEncoder) EncodePrimaryKey(primaryKey []byte, o proto.Message) ([]byte, error) {
	value := o.ProtoReflect().Get(f.fd)
	if !value.IsValid() {
		return nil, fmt.Errorf("invalid value in %s", f.fd.FullName())
	}

	return append(f.IndexPrefix(value), primaryKey...), nil
}

func NewFieldIndexer(typePrefix []byte, sd *v1alpha1.SecondaryKeyDescriptor, messageType protoreflect.MessageType) (*FieldKeyEncoder, error) {
	md := messageType.Descriptor()
	fd := md.Fields().ByName(protoreflect.Name(sd.ProtobufFieldName))
	if fd == nil {
		return nil, fmt.Errorf("field %s does not belogn to message %s", sd.ProtobufFieldName, md.FullName())
	}
	kindEncoder, err := kindencoder.NewKindEncoder(fd.Kind())
	if err != nil {
		return nil, err
	}

	fieldNumberBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(fieldNumberBytes, uint32(fd.Number()))

	return &FieldKeyEncoder{
		fieldNumberBytes: fieldNumberBytes,
		fd:               fd,
		kindEncoder:      kindEncoder,
		typePrefix:       typePrefix,
	}, nil
}
