package abciv1beta1

import (
	fmt "fmt"
	runtime "github.com/cosmos/cosmos-proto/runtime"
	crypto "github.com/cosmos/cosmos-sdk/api/tendermint/crypto"
	_ "github.com/gogo/protobuf/gogoproto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoiface "google.golang.org/protobuf/runtime/protoiface"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/anypb"
	io "io"
	reflect "reflect"
	sync "sync"
)

var (
	md_QueryRequest        protoreflect.MessageDescriptor
	fd_QueryRequest_data   protoreflect.FieldDescriptor
	fd_QueryRequest_path   protoreflect.FieldDescriptor
	fd_QueryRequest_height protoreflect.FieldDescriptor
	fd_QueryRequest_prove  protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_base_abci_v1beta1_query_proto_init()
	md_QueryRequest = File_cosmos_base_abci_v1beta1_query_proto.Messages().ByName("QueryRequest")
	fd_QueryRequest_data = md_QueryRequest.Fields().ByName("data")
	fd_QueryRequest_path = md_QueryRequest.Fields().ByName("path")
	fd_QueryRequest_height = md_QueryRequest.Fields().ByName("height")
	fd_QueryRequest_prove = md_QueryRequest.Fields().ByName("prove")
}

var _ protoreflect.Message = (*fastReflection_QueryRequest)(nil)

type fastReflection_QueryRequest QueryRequest

func (x *QueryRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_QueryRequest)(x)
}

func (x *QueryRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_base_abci_v1beta1_query_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_QueryRequest_messageType fastReflection_QueryRequest_messageType
var _ protoreflect.MessageType = fastReflection_QueryRequest_messageType{}

type fastReflection_QueryRequest_messageType struct{}

func (x fastReflection_QueryRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_QueryRequest)(nil)
}
func (x fastReflection_QueryRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_QueryRequest)
}
func (x fastReflection_QueryRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_QueryRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_QueryRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_QueryRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_QueryRequest) Type() protoreflect.MessageType {
	return _fastReflection_QueryRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_QueryRequest) New() protoreflect.Message {
	return new(fastReflection_QueryRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_QueryRequest) Interface() protoreflect.ProtoMessage {
	return (*QueryRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_QueryRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if len(x.Data) != 0 {
		value := protoreflect.ValueOfBytes(x.Data)
		if !f(fd_QueryRequest_data, value) {
			return
		}
	}
	if x.Path != "" {
		value := protoreflect.ValueOfString(x.Path)
		if !f(fd_QueryRequest_path, value) {
			return
		}
	}
	if x.Height != int64(0) {
		value := protoreflect.ValueOfInt64(x.Height)
		if !f(fd_QueryRequest_height, value) {
			return
		}
	}
	if x.Prove != false {
		value := protoreflect.ValueOfBool(x.Prove)
		if !f(fd_QueryRequest_prove, value) {
			return
		}
	}
}

// Has reports whether a field is populated.
//
// Some fields have the property of nullability where it is possible to
// distinguish between the default value of a field and whether the field
// was explicitly populated with the default value. Singular message fields,
// member fields of a oneof, and proto2 scalar fields are nullable. Such
// fields are populated only if explicitly set.
//
// In other cases (aside from the nullable cases above),
// a proto3 scalar field is populated if it contains a non-zero value, and
// a repeated field is populated if it is non-empty.
func (x *fastReflection_QueryRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryRequest.data":
		return len(x.Data) != 0
	case "cosmos.base.abci.v1beta1.QueryRequest.path":
		return x.Path != ""
	case "cosmos.base.abci.v1beta1.QueryRequest.height":
		return x.Height != int64(0)
	case "cosmos.base.abci.v1beta1.QueryRequest.prove":
		return x.Prove != false
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryRequest"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryRequest.data":
		x.Data = nil
	case "cosmos.base.abci.v1beta1.QueryRequest.path":
		x.Path = ""
	case "cosmos.base.abci.v1beta1.QueryRequest.height":
		x.Height = int64(0)
	case "cosmos.base.abci.v1beta1.QueryRequest.prove":
		x.Prove = false
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryRequest"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_QueryRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.base.abci.v1beta1.QueryRequest.data":
		value := x.Data
		return protoreflect.ValueOfBytes(value)
	case "cosmos.base.abci.v1beta1.QueryRequest.path":
		value := x.Path
		return protoreflect.ValueOfString(value)
	case "cosmos.base.abci.v1beta1.QueryRequest.height":
		value := x.Height
		return protoreflect.ValueOfInt64(value)
	case "cosmos.base.abci.v1beta1.QueryRequest.prove":
		value := x.Prove
		return protoreflect.ValueOfBool(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryRequest"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryRequest does not contain field %s", descriptor.FullName()))
	}
}

// Set stores the value for a field.
//
// For a field belonging to a oneof, it implicitly clears any other field
// that may be currently set within the same oneof.
// For extension fields, it implicitly stores the provided ExtensionType.
// When setting a composite type, it is unspecified whether the stored value
// aliases the source's memory in any way. If the composite value is an
// empty, read-only value, then it panics.
//
// Set is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryRequest.data":
		x.Data = value.Bytes()
	case "cosmos.base.abci.v1beta1.QueryRequest.path":
		x.Path = value.Interface().(string)
	case "cosmos.base.abci.v1beta1.QueryRequest.height":
		x.Height = value.Int()
	case "cosmos.base.abci.v1beta1.QueryRequest.prove":
		x.Prove = value.Bool()
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryRequest"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryRequest does not contain field %s", fd.FullName()))
	}
}

// Mutable returns a mutable reference to a composite type.
//
// If the field is unpopulated, it may allocate a composite value.
// For a field belonging to a oneof, it implicitly clears any other field
// that may be currently set within the same oneof.
// For extension fields, it implicitly stores the provided ExtensionType
// if not already stored.
// It panics if the field does not contain a composite type.
//
// Mutable is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryRequest.data":
		panic(fmt.Errorf("field data of message cosmos.base.abci.v1beta1.QueryRequest is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryRequest.path":
		panic(fmt.Errorf("field path of message cosmos.base.abci.v1beta1.QueryRequest is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryRequest.height":
		panic(fmt.Errorf("field height of message cosmos.base.abci.v1beta1.QueryRequest is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryRequest.prove":
		panic(fmt.Errorf("field prove of message cosmos.base.abci.v1beta1.QueryRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryRequest"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_QueryRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryRequest.data":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.base.abci.v1beta1.QueryRequest.path":
		return protoreflect.ValueOfString("")
	case "cosmos.base.abci.v1beta1.QueryRequest.height":
		return protoreflect.ValueOfInt64(int64(0))
	case "cosmos.base.abci.v1beta1.QueryRequest.prove":
		return protoreflect.ValueOfBool(false)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryRequest"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_QueryRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.base.abci.v1beta1.QueryRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_QueryRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryRequest) SetUnknown(fields protoreflect.RawFields) {
	x.unknownFields = fields
}

// IsValid reports whether the message is valid.
//
// An invalid message is an empty, read-only value.
//
// An invalid message often corresponds to a nil pointer of the concrete
// message type, but the details are implementation dependent.
// Validity is not part of the protobuf data model, and may not
// be preserved in marshaling or other operations.
func (x *fastReflection_QueryRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_QueryRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*QueryRequest)
		if x == nil {
			return protoiface.SizeOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Size:              0,
			}
		}
		options := runtime.SizeInputToOptions(input)
		_ = options
		var n int
		var l int
		_ = l
		l = len(x.Data)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Path)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Height != 0 {
			n += 1 + runtime.Sov(uint64(x.Height))
		}
		if x.Prove {
			n += 2
		}
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*QueryRequest)
		if x == nil {
			return protoiface.MarshalOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Buf:               input.Buf,
			}, nil
		}
		options := runtime.MarshalInputToOptions(input)
		_ = options
		size := options.Size(x)
		dAtA := make([]byte, size)
		i := len(dAtA)
		_ = i
		var l int
		_ = l
		if x.unknownFields != nil {
			i -= len(x.unknownFields)
			copy(dAtA[i:], x.unknownFields)
		}
		if x.Prove {
			i--
			if x.Prove {
				dAtA[i] = 1
			} else {
				dAtA[i] = 0
			}
			i--
			dAtA[i] = 0x20
		}
		if x.Height != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Height))
			i--
			dAtA[i] = 0x18
		}
		if len(x.Path) > 0 {
			i -= len(x.Path)
			copy(dAtA[i:], x.Path)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Path)))
			i--
			dAtA[i] = 0x12
		}
		if len(x.Data) > 0 {
			i -= len(x.Data)
			copy(dAtA[i:], x.Data)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Data)))
			i--
			dAtA[i] = 0xa
		}
		if input.Buf != nil {
			input.Buf = append(input.Buf, dAtA...)
		} else {
			input.Buf = dAtA
		}
		return protoiface.MarshalOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Buf:               input.Buf,
		}, nil
	}
	unmarshal := func(input protoiface.UnmarshalInput) (protoiface.UnmarshalOutput, error) {
		x := input.Message.Interface().(*QueryRequest)
		if x == nil {
			return protoiface.UnmarshalOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Flags:             input.Flags,
			}, nil
		}
		options := runtime.UnmarshalInputToOptions(input)
		_ = options
		dAtA := input.Buf
		l := len(dAtA)
		iNdEx := 0
		for iNdEx < l {
			preIndex := iNdEx
			var wire uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
				}
				if iNdEx >= l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				wire |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			fieldNum := int32(wire >> 3)
			wireType := int(wire & 0x7)
			if wireType == 4 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: QueryRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: QueryRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Data", wireType)
				}
				var byteLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					byteLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if byteLen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + byteLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Data = append(x.Data[:0], dAtA[iNdEx:postIndex]...)
				if x.Data == nil {
					x.Data = []byte{}
				}
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Path", wireType)
				}
				var stringLen uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					stringLen |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLen := int(stringLen)
				if intStringLen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + intStringLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Path = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 3:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
				}
				x.Height = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Height |= int64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 4:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Prove", wireType)
				}
				var v int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				x.Prove = bool(v != 0)
			default:
				iNdEx = preIndex
				skippy, err := runtime.Skip(dAtA[iNdEx:])
				if err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				if (skippy < 0) || (iNdEx+skippy) < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if (iNdEx + skippy) > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				if !options.DiscardUnknown {
					x.unknownFields = append(x.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
				}
				iNdEx += skippy
			}
		}

		if iNdEx > l {
			return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
		}
		return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, nil
	}
	return &protoiface.Methods{
		NoUnkeyedLiterals: struct{}{},
		Flags:             protoiface.SupportMarshalDeterministic | protoiface.SupportUnmarshalDiscardUnknown,
		Size:              size,
		Marshal:           marshal,
		Unmarshal:         unmarshal,
		Merge:             nil,
		CheckInitialized:  nil,
	}
}

var (
	md_QueryResponse           protoreflect.MessageDescriptor
	fd_QueryResponse_code      protoreflect.FieldDescriptor
	fd_QueryResponse_log       protoreflect.FieldDescriptor
	fd_QueryResponse_info      protoreflect.FieldDescriptor
	fd_QueryResponse_index     protoreflect.FieldDescriptor
	fd_QueryResponse_key       protoreflect.FieldDescriptor
	fd_QueryResponse_value     protoreflect.FieldDescriptor
	fd_QueryResponse_proof_ops protoreflect.FieldDescriptor
	fd_QueryResponse_height    protoreflect.FieldDescriptor
	fd_QueryResponse_codespace protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_base_abci_v1beta1_query_proto_init()
	md_QueryResponse = File_cosmos_base_abci_v1beta1_query_proto.Messages().ByName("QueryResponse")
	fd_QueryResponse_code = md_QueryResponse.Fields().ByName("code")
	fd_QueryResponse_log = md_QueryResponse.Fields().ByName("log")
	fd_QueryResponse_info = md_QueryResponse.Fields().ByName("info")
	fd_QueryResponse_index = md_QueryResponse.Fields().ByName("index")
	fd_QueryResponse_key = md_QueryResponse.Fields().ByName("key")
	fd_QueryResponse_value = md_QueryResponse.Fields().ByName("value")
	fd_QueryResponse_proof_ops = md_QueryResponse.Fields().ByName("proof_ops")
	fd_QueryResponse_height = md_QueryResponse.Fields().ByName("height")
	fd_QueryResponse_codespace = md_QueryResponse.Fields().ByName("codespace")
}

var _ protoreflect.Message = (*fastReflection_QueryResponse)(nil)

type fastReflection_QueryResponse QueryResponse

func (x *QueryResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_QueryResponse)(x)
}

func (x *QueryResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_base_abci_v1beta1_query_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_QueryResponse_messageType fastReflection_QueryResponse_messageType
var _ protoreflect.MessageType = fastReflection_QueryResponse_messageType{}

type fastReflection_QueryResponse_messageType struct{}

func (x fastReflection_QueryResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_QueryResponse)(nil)
}
func (x fastReflection_QueryResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_QueryResponse)
}
func (x fastReflection_QueryResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_QueryResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_QueryResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_QueryResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_QueryResponse) Type() protoreflect.MessageType {
	return _fastReflection_QueryResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_QueryResponse) New() protoreflect.Message {
	return new(fastReflection_QueryResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_QueryResponse) Interface() protoreflect.ProtoMessage {
	return (*QueryResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_QueryResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Code != uint32(0) {
		value := protoreflect.ValueOfUint32(x.Code)
		if !f(fd_QueryResponse_code, value) {
			return
		}
	}
	if x.Log != "" {
		value := protoreflect.ValueOfString(x.Log)
		if !f(fd_QueryResponse_log, value) {
			return
		}
	}
	if x.Info != "" {
		value := protoreflect.ValueOfString(x.Info)
		if !f(fd_QueryResponse_info, value) {
			return
		}
	}
	if x.Index != int64(0) {
		value := protoreflect.ValueOfInt64(x.Index)
		if !f(fd_QueryResponse_index, value) {
			return
		}
	}
	if len(x.Key) != 0 {
		value := protoreflect.ValueOfBytes(x.Key)
		if !f(fd_QueryResponse_key, value) {
			return
		}
	}
	if len(x.Value) != 0 {
		value := protoreflect.ValueOfBytes(x.Value)
		if !f(fd_QueryResponse_value, value) {
			return
		}
	}
	if x.ProofOps != nil {
		value := protoreflect.ValueOfMessage(x.ProofOps.ProtoReflect())
		if !f(fd_QueryResponse_proof_ops, value) {
			return
		}
	}
	if x.Height != int64(0) {
		value := protoreflect.ValueOfInt64(x.Height)
		if !f(fd_QueryResponse_height, value) {
			return
		}
	}
	if x.Codespace != "" {
		value := protoreflect.ValueOfString(x.Codespace)
		if !f(fd_QueryResponse_codespace, value) {
			return
		}
	}
}

// Has reports whether a field is populated.
//
// Some fields have the property of nullability where it is possible to
// distinguish between the default value of a field and whether the field
// was explicitly populated with the default value. Singular message fields,
// member fields of a oneof, and proto2 scalar fields are nullable. Such
// fields are populated only if explicitly set.
//
// In other cases (aside from the nullable cases above),
// a proto3 scalar field is populated if it contains a non-zero value, and
// a repeated field is populated if it is non-empty.
func (x *fastReflection_QueryResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryResponse.code":
		return x.Code != uint32(0)
	case "cosmos.base.abci.v1beta1.QueryResponse.log":
		return x.Log != ""
	case "cosmos.base.abci.v1beta1.QueryResponse.info":
		return x.Info != ""
	case "cosmos.base.abci.v1beta1.QueryResponse.index":
		return x.Index != int64(0)
	case "cosmos.base.abci.v1beta1.QueryResponse.key":
		return len(x.Key) != 0
	case "cosmos.base.abci.v1beta1.QueryResponse.value":
		return len(x.Value) != 0
	case "cosmos.base.abci.v1beta1.QueryResponse.proof_ops":
		return x.ProofOps != nil
	case "cosmos.base.abci.v1beta1.QueryResponse.height":
		return x.Height != int64(0)
	case "cosmos.base.abci.v1beta1.QueryResponse.codespace":
		return x.Codespace != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryResponse"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryResponse.code":
		x.Code = uint32(0)
	case "cosmos.base.abci.v1beta1.QueryResponse.log":
		x.Log = ""
	case "cosmos.base.abci.v1beta1.QueryResponse.info":
		x.Info = ""
	case "cosmos.base.abci.v1beta1.QueryResponse.index":
		x.Index = int64(0)
	case "cosmos.base.abci.v1beta1.QueryResponse.key":
		x.Key = nil
	case "cosmos.base.abci.v1beta1.QueryResponse.value":
		x.Value = nil
	case "cosmos.base.abci.v1beta1.QueryResponse.proof_ops":
		x.ProofOps = nil
	case "cosmos.base.abci.v1beta1.QueryResponse.height":
		x.Height = int64(0)
	case "cosmos.base.abci.v1beta1.QueryResponse.codespace":
		x.Codespace = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryResponse"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_QueryResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.base.abci.v1beta1.QueryResponse.code":
		value := x.Code
		return protoreflect.ValueOfUint32(value)
	case "cosmos.base.abci.v1beta1.QueryResponse.log":
		value := x.Log
		return protoreflect.ValueOfString(value)
	case "cosmos.base.abci.v1beta1.QueryResponse.info":
		value := x.Info
		return protoreflect.ValueOfString(value)
	case "cosmos.base.abci.v1beta1.QueryResponse.index":
		value := x.Index
		return protoreflect.ValueOfInt64(value)
	case "cosmos.base.abci.v1beta1.QueryResponse.key":
		value := x.Key
		return protoreflect.ValueOfBytes(value)
	case "cosmos.base.abci.v1beta1.QueryResponse.value":
		value := x.Value
		return protoreflect.ValueOfBytes(value)
	case "cosmos.base.abci.v1beta1.QueryResponse.proof_ops":
		value := x.ProofOps
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	case "cosmos.base.abci.v1beta1.QueryResponse.height":
		value := x.Height
		return protoreflect.ValueOfInt64(value)
	case "cosmos.base.abci.v1beta1.QueryResponse.codespace":
		value := x.Codespace
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryResponse"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryResponse does not contain field %s", descriptor.FullName()))
	}
}

// Set stores the value for a field.
//
// For a field belonging to a oneof, it implicitly clears any other field
// that may be currently set within the same oneof.
// For extension fields, it implicitly stores the provided ExtensionType.
// When setting a composite type, it is unspecified whether the stored value
// aliases the source's memory in any way. If the composite value is an
// empty, read-only value, then it panics.
//
// Set is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryResponse.code":
		x.Code = uint32(value.Uint())
	case "cosmos.base.abci.v1beta1.QueryResponse.log":
		x.Log = value.Interface().(string)
	case "cosmos.base.abci.v1beta1.QueryResponse.info":
		x.Info = value.Interface().(string)
	case "cosmos.base.abci.v1beta1.QueryResponse.index":
		x.Index = value.Int()
	case "cosmos.base.abci.v1beta1.QueryResponse.key":
		x.Key = value.Bytes()
	case "cosmos.base.abci.v1beta1.QueryResponse.value":
		x.Value = value.Bytes()
	case "cosmos.base.abci.v1beta1.QueryResponse.proof_ops":
		x.ProofOps = value.Message().Interface().(*crypto.ProofOps)
	case "cosmos.base.abci.v1beta1.QueryResponse.height":
		x.Height = value.Int()
	case "cosmos.base.abci.v1beta1.QueryResponse.codespace":
		x.Codespace = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryResponse"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryResponse does not contain field %s", fd.FullName()))
	}
}

// Mutable returns a mutable reference to a composite type.
//
// If the field is unpopulated, it may allocate a composite value.
// For a field belonging to a oneof, it implicitly clears any other field
// that may be currently set within the same oneof.
// For extension fields, it implicitly stores the provided ExtensionType
// if not already stored.
// It panics if the field does not contain a composite type.
//
// Mutable is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryResponse.proof_ops":
		if x.ProofOps == nil {
			x.ProofOps = new(crypto.ProofOps)
		}
		return protoreflect.ValueOfMessage(x.ProofOps.ProtoReflect())
	case "cosmos.base.abci.v1beta1.QueryResponse.code":
		panic(fmt.Errorf("field code of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryResponse.log":
		panic(fmt.Errorf("field log of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryResponse.info":
		panic(fmt.Errorf("field info of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryResponse.index":
		panic(fmt.Errorf("field index of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryResponse.key":
		panic(fmt.Errorf("field key of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryResponse.value":
		panic(fmt.Errorf("field value of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryResponse.height":
		panic(fmt.Errorf("field height of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	case "cosmos.base.abci.v1beta1.QueryResponse.codespace":
		panic(fmt.Errorf("field codespace of message cosmos.base.abci.v1beta1.QueryResponse is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryResponse"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_QueryResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.abci.v1beta1.QueryResponse.code":
		return protoreflect.ValueOfUint32(uint32(0))
	case "cosmos.base.abci.v1beta1.QueryResponse.log":
		return protoreflect.ValueOfString("")
	case "cosmos.base.abci.v1beta1.QueryResponse.info":
		return protoreflect.ValueOfString("")
	case "cosmos.base.abci.v1beta1.QueryResponse.index":
		return protoreflect.ValueOfInt64(int64(0))
	case "cosmos.base.abci.v1beta1.QueryResponse.key":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.base.abci.v1beta1.QueryResponse.value":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.base.abci.v1beta1.QueryResponse.proof_ops":
		m := new(crypto.ProofOps)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	case "cosmos.base.abci.v1beta1.QueryResponse.height":
		return protoreflect.ValueOfInt64(int64(0))
	case "cosmos.base.abci.v1beta1.QueryResponse.codespace":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.abci.v1beta1.QueryResponse"))
		}
		panic(fmt.Errorf("message cosmos.base.abci.v1beta1.QueryResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_QueryResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.base.abci.v1beta1.QueryResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_QueryResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_QueryResponse) SetUnknown(fields protoreflect.RawFields) {
	x.unknownFields = fields
}

// IsValid reports whether the message is valid.
//
// An invalid message is an empty, read-only value.
//
// An invalid message often corresponds to a nil pointer of the concrete
// message type, but the details are implementation dependent.
// Validity is not part of the protobuf data model, and may not
// be preserved in marshaling or other operations.
func (x *fastReflection_QueryResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_QueryResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*QueryResponse)
		if x == nil {
			return protoiface.SizeOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Size:              0,
			}
		}
		options := runtime.SizeInputToOptions(input)
		_ = options
		var n int
		var l int
		_ = l
		if x.Code != 0 {
			n += 1 + runtime.Sov(uint64(x.Code))
		}
		l = len(x.Log)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Info)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Index != 0 {
			n += 1 + runtime.Sov(uint64(x.Index))
		}
		l = len(x.Key)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Value)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.ProofOps != nil {
			l = options.Size(x.ProofOps)
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Height != 0 {
			n += 1 + runtime.Sov(uint64(x.Height))
		}
		l = len(x.Codespace)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*QueryResponse)
		if x == nil {
			return protoiface.MarshalOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Buf:               input.Buf,
			}, nil
		}
		options := runtime.MarshalInputToOptions(input)
		_ = options
		size := options.Size(x)
		dAtA := make([]byte, size)
		i := len(dAtA)
		_ = i
		var l int
		_ = l
		if x.unknownFields != nil {
			i -= len(x.unknownFields)
			copy(dAtA[i:], x.unknownFields)
		}
		if len(x.Codespace) > 0 {
			i -= len(x.Codespace)
			copy(dAtA[i:], x.Codespace)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Codespace)))
			i--
			dAtA[i] = 0x52
		}
		if x.Height != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Height))
			i--
			dAtA[i] = 0x48
		}
		if x.ProofOps != nil {
			encoded, err := options.Marshal(x.ProofOps)
			if err != nil {
				return protoiface.MarshalOutput{
					NoUnkeyedLiterals: input.NoUnkeyedLiterals,
					Buf:               input.Buf,
				}, err
			}
			i -= len(encoded)
			copy(dAtA[i:], encoded)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(encoded)))
			i--
			dAtA[i] = 0x42
		}
		if len(x.Value) > 0 {
			i -= len(x.Value)
			copy(dAtA[i:], x.Value)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Value)))
			i--
			dAtA[i] = 0x3a
		}
		if len(x.Key) > 0 {
			i -= len(x.Key)
			copy(dAtA[i:], x.Key)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Key)))
			i--
			dAtA[i] = 0x32
		}
		if x.Index != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Index))
			i--
			dAtA[i] = 0x28
		}
		if len(x.Info) > 0 {
			i -= len(x.Info)
			copy(dAtA[i:], x.Info)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Info)))
			i--
			dAtA[i] = 0x22
		}
		if len(x.Log) > 0 {
			i -= len(x.Log)
			copy(dAtA[i:], x.Log)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Log)))
			i--
			dAtA[i] = 0x1a
		}
		if x.Code != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Code))
			i--
			dAtA[i] = 0x8
		}
		if input.Buf != nil {
			input.Buf = append(input.Buf, dAtA...)
		} else {
			input.Buf = dAtA
		}
		return protoiface.MarshalOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Buf:               input.Buf,
		}, nil
	}
	unmarshal := func(input protoiface.UnmarshalInput) (protoiface.UnmarshalOutput, error) {
		x := input.Message.Interface().(*QueryResponse)
		if x == nil {
			return protoiface.UnmarshalOutput{
				NoUnkeyedLiterals: input.NoUnkeyedLiterals,
				Flags:             input.Flags,
			}, nil
		}
		options := runtime.UnmarshalInputToOptions(input)
		_ = options
		dAtA := input.Buf
		l := len(dAtA)
		iNdEx := 0
		for iNdEx < l {
			preIndex := iNdEx
			var wire uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
				}
				if iNdEx >= l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				wire |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			fieldNum := int32(wire >> 3)
			wireType := int(wire & 0x7)
			if wireType == 4 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: QueryResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: QueryResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Code", wireType)
				}
				x.Code = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Code |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Log", wireType)
				}
				var stringLen uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					stringLen |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLen := int(stringLen)
				if intStringLen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + intStringLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Log = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 4:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Info", wireType)
				}
				var stringLen uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					stringLen |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLen := int(stringLen)
				if intStringLen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + intStringLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Info = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 5:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
				}
				x.Index = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Index |= int64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 6:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
				}
				var byteLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					byteLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if byteLen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + byteLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Key = append(x.Key[:0], dAtA[iNdEx:postIndex]...)
				if x.Key == nil {
					x.Key = []byte{}
				}
				iNdEx = postIndex
			case 7:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
				}
				var byteLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					byteLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if byteLen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + byteLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Value = append(x.Value[:0], dAtA[iNdEx:postIndex]...)
				if x.Value == nil {
					x.Value = []byte{}
				}
				iNdEx = postIndex
			case 8:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field ProofOps", wireType)
				}
				var msglen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					msglen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if msglen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + msglen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				if x.ProofOps == nil {
					x.ProofOps = &crypto.ProofOps{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.ProofOps); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 9:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Height", wireType)
				}
				x.Height = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Height |= int64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 10:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Codespace", wireType)
				}
				var stringLen uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					stringLen |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLen := int(stringLen)
				if intStringLen < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				postIndex := iNdEx + intStringLen
				if postIndex < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if postIndex > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Codespace = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			default:
				iNdEx = preIndex
				skippy, err := runtime.Skip(dAtA[iNdEx:])
				if err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				if (skippy < 0) || (iNdEx+skippy) < 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
				}
				if (iNdEx + skippy) > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				if !options.DiscardUnknown {
					x.unknownFields = append(x.unknownFields, dAtA[iNdEx:iNdEx+skippy]...)
				}
				iNdEx += skippy
			}
		}

		if iNdEx > l {
			return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
		}
		return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, nil
	}
	return &protoiface.Methods{
		NoUnkeyedLiterals: struct{}{},
		Flags:             protoiface.SupportMarshalDeterministic | protoiface.SupportUnmarshalDiscardUnknown,
		Size:              size,
		Marshal:           marshal,
		Unmarshal:         unmarshal,
		Merge:             nil,
		CheckInitialized:  nil,
	}
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.0
// 	protoc        (unknown)
// source: cosmos/base/abci/v1beta1/query.proto

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// The request format for an ABCI Query
type QueryRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data   []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Path   string `protobuf:"bytes,2,opt,name=path,proto3" json:"path,omitempty"`
	Height int64  `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Prove  bool   `protobuf:"varint,4,opt,name=prove,proto3" json:"prove,omitempty"`
}

func (x *QueryRequest) Reset() {
	*x = QueryRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_base_abci_v1beta1_query_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryRequest) ProtoMessage() {}

// Deprecated: Use QueryRequest.ProtoReflect.Descriptor instead.
func (*QueryRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_base_abci_v1beta1_query_proto_rawDescGZIP(), []int{0}
}

func (x *QueryRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *QueryRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *QueryRequest) GetHeight() int64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *QueryRequest) GetProve() bool {
	if x != nil {
		return x.Prove
	}
	return false
}

// The response format for an ABCI Query
type QueryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code uint32 `protobuf:"varint,1,opt,name=code,proto3" json:"code,omitempty"`
	// bytes data = 2; // use "value" instead.
	Log       string           `protobuf:"bytes,3,opt,name=log,proto3" json:"log,omitempty"`   // nondeterministic
	Info      string           `protobuf:"bytes,4,opt,name=info,proto3" json:"info,omitempty"` // nondeterministic
	Index     int64            `protobuf:"varint,5,opt,name=index,proto3" json:"index,omitempty"`
	Key       []byte           `protobuf:"bytes,6,opt,name=key,proto3" json:"key,omitempty"`
	Value     []byte           `protobuf:"bytes,7,opt,name=value,proto3" json:"value,omitempty"`
	ProofOps  *crypto.ProofOps `protobuf:"bytes,8,opt,name=proof_ops,json=proofOps,proto3" json:"proof_ops,omitempty"`
	Height    int64            `protobuf:"varint,9,opt,name=height,proto3" json:"height,omitempty"`
	Codespace string           `protobuf:"bytes,10,opt,name=codespace,proto3" json:"codespace,omitempty"`
}

func (x *QueryResponse) Reset() {
	*x = QueryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_base_abci_v1beta1_query_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryResponse) ProtoMessage() {}

// Deprecated: Use QueryResponse.ProtoReflect.Descriptor instead.
func (*QueryResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_base_abci_v1beta1_query_proto_rawDescGZIP(), []int{1}
}

func (x *QueryResponse) GetCode() uint32 {
	if x != nil {
		return x.Code
	}
	return 0
}

func (x *QueryResponse) GetLog() string {
	if x != nil {
		return x.Log
	}
	return ""
}

func (x *QueryResponse) GetInfo() string {
	if x != nil {
		return x.Info
	}
	return ""
}

func (x *QueryResponse) GetIndex() int64 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *QueryResponse) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *QueryResponse) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *QueryResponse) GetProofOps() *crypto.ProofOps {
	if x != nil {
		return x.ProofOps
	}
	return nil
}

func (x *QueryResponse) GetHeight() int64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *QueryResponse) GetCodespace() string {
	if x != nil {
		return x.Codespace
	}
	return ""
}

var File_cosmos_base_abci_v1beta1_query_proto protoreflect.FileDescriptor

var file_cosmos_base_abci_v1beta1_query_proto_rawDesc = []byte{
	0x0a, 0x24, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x2f, 0x61, 0x62,
	0x63, 0x69, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x18, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62,
	0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31,
	0x1a, 0x14, 0x67, 0x6f, 0x67, 0x6f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x67, 0x6f,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1d, 0x74, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x74, 0x2f, 0x63, 0x72,
	0x79, 0x70, 0x74, 0x6f, 0x2f, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0x64, 0x0a, 0x0c, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x05, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x22, 0xf7, 0x01, 0x0a, 0x0d, 0x51, 0x75, 0x65, 0x72, 0x79,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x6c, 0x6f, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6c, 0x6f, 0x67, 0x12, 0x12,
	0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x69, 0x6e,
	0x66, 0x6f, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x03, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61,
	0x6c, 0x75, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x12, 0x38, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x5f, 0x6f, 0x70, 0x73, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x74, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x6d, 0x69, 0x6e, 0x74,
	0x2e, 0x63, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x2e, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x4f, 0x70, 0x73,
	0x52, 0x08, 0x70, 0x72, 0x6f, 0x6f, 0x66, 0x4f, 0x70, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x68, 0x65,
	0x69, 0x67, 0x68, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x68, 0x65, 0x69, 0x67,
	0x68, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x64, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18,
	0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63, 0x6f, 0x64, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x32, 0x63, 0x0a, 0x07, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x58, 0x0a, 0x05, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x12, 0x26, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61,
	0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e,
	0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x27, 0x2e, 0x63,
	0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xf4, 0x01, 0x0a, 0x1c, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x61, 0x62, 0x63, 0x69, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x42, 0x0a, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2d, 0x73,
	0x64, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x62, 0x61,
	0x73, 0x65, 0x2f, 0x61, 0x62, 0x63, 0x69, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x3b,
	0x61, 0x62, 0x63, 0x69, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x42,
	0x41, 0xaa, 0x02, 0x18, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x42, 0x61, 0x73, 0x65, 0x2e,
	0x41, 0x62, 0x63, 0x69, 0x2e, 0x56, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0xca, 0x02, 0x18, 0x43,
	0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x42, 0x61, 0x73, 0x65, 0x5c, 0x41, 0x62, 0x63, 0x69, 0x5c,
	0x56, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0xe2, 0x02, 0x24, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x5c, 0x42, 0x61, 0x73, 0x65, 0x5c, 0x41, 0x62, 0x63, 0x69, 0x5c, 0x56, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x1b, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x3a, 0x3a, 0x42, 0x61, 0x73, 0x65, 0x3a, 0x3a, 0x41,
	0x62, 0x63, 0x69, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cosmos_base_abci_v1beta1_query_proto_rawDescOnce sync.Once
	file_cosmos_base_abci_v1beta1_query_proto_rawDescData = file_cosmos_base_abci_v1beta1_query_proto_rawDesc
)

func file_cosmos_base_abci_v1beta1_query_proto_rawDescGZIP() []byte {
	file_cosmos_base_abci_v1beta1_query_proto_rawDescOnce.Do(func() {
		file_cosmos_base_abci_v1beta1_query_proto_rawDescData = protoimpl.X.CompressGZIP(file_cosmos_base_abci_v1beta1_query_proto_rawDescData)
	})
	return file_cosmos_base_abci_v1beta1_query_proto_rawDescData
}

var file_cosmos_base_abci_v1beta1_query_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_cosmos_base_abci_v1beta1_query_proto_goTypes = []interface{}{
	(*QueryRequest)(nil),    // 0: cosmos.base.abci.v1beta1.QueryRequest
	(*QueryResponse)(nil),   // 1: cosmos.base.abci.v1beta1.QueryResponse
	(*crypto.ProofOps)(nil), // 2: tendermint.crypto.ProofOps
}
var file_cosmos_base_abci_v1beta1_query_proto_depIdxs = []int32{
	2, // 0: cosmos.base.abci.v1beta1.QueryResponse.proof_ops:type_name -> tendermint.crypto.ProofOps
	0, // 1: cosmos.base.abci.v1beta1.Service.Query:input_type -> cosmos.base.abci.v1beta1.QueryRequest
	1, // 2: cosmos.base.abci.v1beta1.Service.Query:output_type -> cosmos.base.abci.v1beta1.QueryResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_cosmos_base_abci_v1beta1_query_proto_init() }
func file_cosmos_base_abci_v1beta1_query_proto_init() {
	if File_cosmos_base_abci_v1beta1_query_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cosmos_base_abci_v1beta1_query_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_cosmos_base_abci_v1beta1_query_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cosmos_base_abci_v1beta1_query_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cosmos_base_abci_v1beta1_query_proto_goTypes,
		DependencyIndexes: file_cosmos_base_abci_v1beta1_query_proto_depIdxs,
		MessageInfos:      file_cosmos_base_abci_v1beta1_query_proto_msgTypes,
	}.Build()
	File_cosmos_base_abci_v1beta1_query_proto = out.File
	file_cosmos_base_abci_v1beta1_query_proto_rawDesc = nil
	file_cosmos_base_abci_v1beta1_query_proto_goTypes = nil
	file_cosmos_base_abci_v1beta1_query_proto_depIdxs = nil
}
