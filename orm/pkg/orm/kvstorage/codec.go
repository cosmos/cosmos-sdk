package kvstorage

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/orm/apis/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/pkg/protoutils/kindencoder"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	SingletonKey = []byte{0x0}
)

// Codec represents a state object encoder/decoder
type Codec struct {
	messageType      protoreflect.MessageType
	primaryKeyEncode func(object proto.Message) ([]byte, error)
}

// EncodePrimaryKey provides the object's primary key
func (c *Codec) EncodePrimaryKey(object proto.Message) ([]byte, error) {
	return c.primaryKeyEncode(object)
}

// EncodeObject encodes the object to bytes
// TODO(fdymylja): this exists to provide primary key fields clearance
func (c *Codec) EncodeObject(object proto.Message) ([]byte, error) {
	// TODO(fdymylja): maybe this wants a custom marshaler
	return proto.MarshalOptions{Deterministic: true}.Marshal(object)
}

// DecodeObject decodes the provided bytes to target
// TODO(fdymylja): this exists to decode the object and then decode the primary key and fill the missing fields
func (c *Codec) DecodeObject(primaryKey, objectBytes []byte, target proto.Message) error {
	// TODO(fdymylja): maybe this wants a custom unmarshaler
	return proto.UnmarshalOptions{}.Unmarshal(objectBytes, target)
}

func NewCodec(td *v1alpha1.TableDescriptor, messageType protoreflect.MessageType) (*Codec, error) {
	md := messageType.Descriptor()

	if td.Singleton {
		return &Codec{
			messageType: messageType,
			primaryKeyEncode: func(object proto.Message) ([]byte, error) {
				return SingletonKey, nil
			},
		}, nil
	}

	pkd := td.PrimaryKey

	if len(pkd.FieldNames) == 0 {
		return nil, fmt.Errorf("no protobuf field names are defined")
	}

	type fieldEncodeFunc func(object proto.Message) ([]byte, error)

	var fieldEncodeFuncs []fieldEncodeFunc

	for _, field := range pkd.FieldNames {
		fd := md.Fields().ByName(protoreflect.Name(field))
		if fd == nil {
			return nil, fmt.Errorf("field %s does not belong to %s", field, md.FullName())
		}
		// TODO(fdymylja): check field is valid kind
		kindEncoder, err := kindencoder.NewKindEncoder(fd.Kind())
		if err != nil {
			return nil, err
		}

		fieldEncodeFuncs = append(fieldEncodeFuncs, func(object proto.Message) ([]byte, error) {
			v := object.ProtoReflect().Get(fd)
			if !v.IsValid() {
				return nil, fmt.Errorf("field %s is invalid", fd.FullName())
			}
			return kindEncoder.EncodeValueToBytes(v), nil
		})
	}

	return &Codec{
		messageType: messageType,
		primaryKeyEncode: func(object proto.Message) ([]byte, error) {
			// TODO(fdymylja): maybe we can efficiently pre-compute primary key size
			var key []byte
			for _, fieldEnc := range fieldEncodeFuncs {
				keyPart, err := fieldEnc(object)
				if err != nil {
					return nil, err
				}
				key = append(key, keyPart...)
			}

			if len(key) == 0 {
				return nil, fmt.Errorf("object computed an empty key")
			}

			return key, nil
		},
	}, nil
}
