package groupv1beta1

import (
	context "context"
	fmt "fmt"
	io "io"
	reflect "reflect"
	sync "sync"

	_ "github.com/cosmos/cosmos-proto"
	runtime "github.com/cosmos/cosmos-proto/runtime"
	_ "github.com/gogo/protobuf/gogoproto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoiface "google.golang.org/protobuf/runtime/protoiface"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

var _ protoreflect.List = (*_MsgCreateGroupRequest_2_list)(nil)

type _MsgCreateGroupRequest_2_list struct {
	list *[]*Member
}

func (x *_MsgCreateGroupRequest_2_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_MsgCreateGroupRequest_2_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfMessage((*x.list)[i].ProtoReflect())
}

func (x *_MsgCreateGroupRequest_2_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*Member)
	(*x.list)[i] = concreteValue
}

func (x *_MsgCreateGroupRequest_2_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*Member)
	*x.list = append(*x.list, concreteValue)
}

func (x *_MsgCreateGroupRequest_2_list) AppendMutable() protoreflect.Value {
	v := new(Member)
	*x.list = append(*x.list, v)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_MsgCreateGroupRequest_2_list) Truncate(n int) {
	for i := n; i < len(*x.list); i++ {
		(*x.list)[i] = nil
	}
	*x.list = (*x.list)[:n]
}

func (x *_MsgCreateGroupRequest_2_list) NewElement() protoreflect.Value {
	v := new(Member)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_MsgCreateGroupRequest_2_list) IsValid() bool {
	return x.list != nil
}

var (
	md_MsgCreateGroupRequest          protoreflect.MessageDescriptor
	fd_MsgCreateGroupRequest_admin    protoreflect.FieldDescriptor
	fd_MsgCreateGroupRequest_members  protoreflect.FieldDescriptor
	fd_MsgCreateGroupRequest_metadata protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgCreateGroupRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgCreateGroupRequest")
	fd_MsgCreateGroupRequest_admin = md_MsgCreateGroupRequest.Fields().ByName("admin")
	fd_MsgCreateGroupRequest_members = md_MsgCreateGroupRequest.Fields().ByName("members")
	fd_MsgCreateGroupRequest_metadata = md_MsgCreateGroupRequest.Fields().ByName("metadata")
}

var _ protoreflect.Message = (*fastReflection_MsgCreateGroupRequest)(nil)

type fastReflection_MsgCreateGroupRequest MsgCreateGroupRequest

func (x *MsgCreateGroupRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupRequest)(x)
}

func (x *MsgCreateGroupRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgCreateGroupRequest_messageType fastReflection_MsgCreateGroupRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgCreateGroupRequest_messageType{}

type fastReflection_MsgCreateGroupRequest_messageType struct{}

func (x fastReflection_MsgCreateGroupRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupRequest)(nil)
}
func (x fastReflection_MsgCreateGroupRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupRequest)
}
func (x fastReflection_MsgCreateGroupRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgCreateGroupRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgCreateGroupRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgCreateGroupRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgCreateGroupRequest) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgCreateGroupRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgCreateGroupRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgCreateGroupRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgCreateGroupRequest_admin, value) {
			return
		}
	}
	if len(x.Members) != 0 {
		value := protoreflect.ValueOfList(&_MsgCreateGroupRequest_2_list{list: &x.Members})
		if !f(fd_MsgCreateGroupRequest_members, value) {
			return
		}
	}
	if len(x.Metadata) != 0 {
		value := protoreflect.ValueOfBytes(x.Metadata)
		if !f(fd_MsgCreateGroupRequest_metadata, value) {
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
func (x *fastReflection_MsgCreateGroupRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.members":
		return len(x.Members) != 0
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.metadata":
		return len(x.Metadata) != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.members":
		x.Members = nil
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.metadata":
		x.Metadata = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgCreateGroupRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.members":
		if len(x.Members) == 0 {
			return protoreflect.ValueOfList(&_MsgCreateGroupRequest_2_list{})
		}
		listValue := &_MsgCreateGroupRequest_2_list{list: &x.Members}
		return protoreflect.ValueOfList(listValue)
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.metadata":
		value := x.Metadata
		return protoreflect.ValueOfBytes(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgCreateGroupRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.members":
		lv := value.List()
		clv := lv.(*_MsgCreateGroupRequest_2_list)
		x.Members = *clv.list
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.metadata":
		x.Metadata = value.Bytes()
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgCreateGroupRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.members":
		if x.Members == nil {
			x.Members = []*Member{}
		}
		value := &_MsgCreateGroupRequest_2_list{list: &x.Members}
		return protoreflect.ValueOfList(value)
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgCreateGroupRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.metadata":
		panic(fmt.Errorf("field metadata of message cosmos.group.v1beta1.MsgCreateGroupRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgCreateGroupRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.members":
		list := []*Member{}
		return protoreflect.ValueOfList(&_MsgCreateGroupRequest_2_list{list: &list})
	case "cosmos.group.v1beta1.MsgCreateGroupRequest.metadata":
		return protoreflect.ValueOfBytes(nil)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgCreateGroupRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgCreateGroupRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgCreateGroupRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgCreateGroupRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgCreateGroupRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgCreateGroupRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if len(x.Members) > 0 {
			for _, e := range x.Members {
				l = options.Size(e)
				n += 1 + l + runtime.Sov(uint64(l))
			}
		}
		l = len(x.Metadata)
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
		x := input.Message.Interface().(*MsgCreateGroupRequest)
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
		if len(x.Metadata) > 0 {
			i -= len(x.Metadata)
			copy(dAtA[i:], x.Metadata)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Metadata)))
			i--
			dAtA[i] = 0x1a
		}
		if len(x.Members) > 0 {
			for iNdEx := len(x.Members) - 1; iNdEx >= 0; iNdEx-- {
				encoded, err := options.Marshal(x.Members[iNdEx])
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
				dAtA[i] = 0x12
			}
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgCreateGroupRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Members", wireType)
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
				x.Members = append(x.Members, &Member{})
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Members[len(x.Members)-1]); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
				x.Metadata = append(x.Metadata[:0], dAtA[iNdEx:postIndex]...)
				if x.Metadata == nil {
					x.Metadata = []byte{}
				}
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

var (
	md_MsgCreateGroupResponse          protoreflect.MessageDescriptor
	fd_MsgCreateGroupResponse_group_id protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgCreateGroupResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgCreateGroupResponse")
	fd_MsgCreateGroupResponse_group_id = md_MsgCreateGroupResponse.Fields().ByName("group_id")
}

var _ protoreflect.Message = (*fastReflection_MsgCreateGroupResponse)(nil)

type fastReflection_MsgCreateGroupResponse MsgCreateGroupResponse

func (x *MsgCreateGroupResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupResponse)(x)
}

func (x *MsgCreateGroupResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgCreateGroupResponse_messageType fastReflection_MsgCreateGroupResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgCreateGroupResponse_messageType{}

type fastReflection_MsgCreateGroupResponse_messageType struct{}

func (x fastReflection_MsgCreateGroupResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupResponse)(nil)
}
func (x fastReflection_MsgCreateGroupResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupResponse)
}
func (x fastReflection_MsgCreateGroupResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgCreateGroupResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgCreateGroupResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgCreateGroupResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgCreateGroupResponse) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgCreateGroupResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgCreateGroupResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgCreateGroupResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.GroupId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.GroupId)
		if !f(fd_MsgCreateGroupResponse_group_id, value) {
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
func (x *fastReflection_MsgCreateGroupResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupResponse.group_id":
		return x.GroupId != uint64(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupResponse.group_id":
		x.GroupId = uint64(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgCreateGroupResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupResponse.group_id":
		value := x.GroupId
		return protoreflect.ValueOfUint64(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgCreateGroupResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupResponse.group_id":
		x.GroupId = value.Uint()
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgCreateGroupResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupResponse.group_id":
		panic(fmt.Errorf("field group_id of message cosmos.group.v1beta1.MsgCreateGroupResponse is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgCreateGroupResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupResponse.group_id":
		return protoreflect.ValueOfUint64(uint64(0))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgCreateGroupResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgCreateGroupResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgCreateGroupResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgCreateGroupResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgCreateGroupResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgCreateGroupResponse)
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
		if x.GroupId != 0 {
			n += 1 + runtime.Sov(uint64(x.GroupId))
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
		x := input.Message.Interface().(*MsgCreateGroupResponse)
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
		if x.GroupId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.GroupId))
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
		x := input.Message.Interface().(*MsgCreateGroupResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
				}
				x.GroupId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.GroupId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
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

var _ protoreflect.List = (*_MsgUpdateGroupMembersRequest_3_list)(nil)

type _MsgUpdateGroupMembersRequest_3_list struct {
	list *[]*Member
}

func (x *_MsgUpdateGroupMembersRequest_3_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_MsgUpdateGroupMembersRequest_3_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfMessage((*x.list)[i].ProtoReflect())
}

func (x *_MsgUpdateGroupMembersRequest_3_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*Member)
	(*x.list)[i] = concreteValue
}

func (x *_MsgUpdateGroupMembersRequest_3_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*Member)
	*x.list = append(*x.list, concreteValue)
}

func (x *_MsgUpdateGroupMembersRequest_3_list) AppendMutable() protoreflect.Value {
	v := new(Member)
	*x.list = append(*x.list, v)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_MsgUpdateGroupMembersRequest_3_list) Truncate(n int) {
	for i := n; i < len(*x.list); i++ {
		(*x.list)[i] = nil
	}
	*x.list = (*x.list)[:n]
}

func (x *_MsgUpdateGroupMembersRequest_3_list) NewElement() protoreflect.Value {
	v := new(Member)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_MsgUpdateGroupMembersRequest_3_list) IsValid() bool {
	return x.list != nil
}

var (
	md_MsgUpdateGroupMembersRequest                protoreflect.MessageDescriptor
	fd_MsgUpdateGroupMembersRequest_admin          protoreflect.FieldDescriptor
	fd_MsgUpdateGroupMembersRequest_group_id       protoreflect.FieldDescriptor
	fd_MsgUpdateGroupMembersRequest_member_updates protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupMembersRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupMembersRequest")
	fd_MsgUpdateGroupMembersRequest_admin = md_MsgUpdateGroupMembersRequest.Fields().ByName("admin")
	fd_MsgUpdateGroupMembersRequest_group_id = md_MsgUpdateGroupMembersRequest.Fields().ByName("group_id")
	fd_MsgUpdateGroupMembersRequest_member_updates = md_MsgUpdateGroupMembersRequest.Fields().ByName("member_updates")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupMembersRequest)(nil)

type fastReflection_MsgUpdateGroupMembersRequest MsgUpdateGroupMembersRequest

func (x *MsgUpdateGroupMembersRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMembersRequest)(x)
}

func (x *MsgUpdateGroupMembersRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupMembersRequest_messageType fastReflection_MsgUpdateGroupMembersRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupMembersRequest_messageType{}

type fastReflection_MsgUpdateGroupMembersRequest_messageType struct{}

func (x fastReflection_MsgUpdateGroupMembersRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMembersRequest)(nil)
}
func (x fastReflection_MsgUpdateGroupMembersRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMembersRequest)
}
func (x fastReflection_MsgUpdateGroupMembersRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMembersRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupMembersRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMembersRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupMembersRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupMembersRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupMembersRequest) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMembersRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupMembersRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupMembersRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupMembersRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgUpdateGroupMembersRequest_admin, value) {
			return
		}
	}
	if x.GroupId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.GroupId)
		if !f(fd_MsgUpdateGroupMembersRequest_group_id, value) {
			return
		}
	}
	if len(x.MemberUpdates) != 0 {
		value := protoreflect.ValueOfList(&_MsgUpdateGroupMembersRequest_3_list{list: &x.MemberUpdates})
		if !f(fd_MsgUpdateGroupMembersRequest_member_updates, value) {
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
func (x *fastReflection_MsgUpdateGroupMembersRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.group_id":
		return x.GroupId != uint64(0)
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.member_updates":
		return len(x.MemberUpdates) != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMembersRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.group_id":
		x.GroupId = uint64(0)
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.member_updates":
		x.MemberUpdates = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupMembersRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.group_id":
		value := x.GroupId
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.member_updates":
		if len(x.MemberUpdates) == 0 {
			return protoreflect.ValueOfList(&_MsgUpdateGroupMembersRequest_3_list{})
		}
		listValue := &_MsgUpdateGroupMembersRequest_3_list{list: &x.MemberUpdates}
		return protoreflect.ValueOfList(listValue)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMembersRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.group_id":
		x.GroupId = value.Uint()
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.member_updates":
		lv := value.List()
		clv := lv.(*_MsgUpdateGroupMembersRequest_3_list)
		x.MemberUpdates = *clv.list
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMembersRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.member_updates":
		if x.MemberUpdates == nil {
			x.MemberUpdates = []*Member{}
		}
		value := &_MsgUpdateGroupMembersRequest_3_list{list: &x.MemberUpdates}
		return protoreflect.ValueOfList(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.group_id":
		panic(fmt.Errorf("field group_id of message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupMembersRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.group_id":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.member_updates":
		list := []*Member{}
		return protoreflect.ValueOfList(&_MsgUpdateGroupMembersRequest_3_list{list: &list})
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupMembersRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupMembersRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupMembersRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMembersRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupMembersRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupMembersRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupMembersRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.GroupId != 0 {
			n += 1 + runtime.Sov(uint64(x.GroupId))
		}
		if len(x.MemberUpdates) > 0 {
			for _, e := range x.MemberUpdates {
				l = options.Size(e)
				n += 1 + l + runtime.Sov(uint64(l))
			}
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
		x := input.Message.Interface().(*MsgUpdateGroupMembersRequest)
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
		if len(x.MemberUpdates) > 0 {
			for iNdEx := len(x.MemberUpdates) - 1; iNdEx >= 0; iNdEx-- {
				encoded, err := options.Marshal(x.MemberUpdates[iNdEx])
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
				dAtA[i] = 0x1a
			}
		}
		if x.GroupId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.GroupId))
			i--
			dAtA[i] = 0x10
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgUpdateGroupMembersRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMembersRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMembersRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
				}
				x.GroupId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.GroupId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field MemberUpdates", wireType)
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
				x.MemberUpdates = append(x.MemberUpdates, &Member{})
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.MemberUpdates[len(x.MemberUpdates)-1]); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
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

var (
	md_MsgUpdateGroupMembersResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupMembersResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupMembersResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupMembersResponse)(nil)

type fastReflection_MsgUpdateGroupMembersResponse MsgUpdateGroupMembersResponse

func (x *MsgUpdateGroupMembersResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMembersResponse)(x)
}

func (x *MsgUpdateGroupMembersResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupMembersResponse_messageType fastReflection_MsgUpdateGroupMembersResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupMembersResponse_messageType{}

type fastReflection_MsgUpdateGroupMembersResponse_messageType struct{}

func (x fastReflection_MsgUpdateGroupMembersResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMembersResponse)(nil)
}
func (x fastReflection_MsgUpdateGroupMembersResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMembersResponse)
}
func (x fastReflection_MsgUpdateGroupMembersResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMembersResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupMembersResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMembersResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupMembersResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupMembersResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupMembersResponse) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMembersResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupMembersResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupMembersResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupMembersResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgUpdateGroupMembersResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMembersResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupMembersResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMembersResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMembersResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupMembersResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMembersResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMembersResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupMembersResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupMembersResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupMembersResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMembersResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupMembersResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupMembersResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupMembersResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgUpdateGroupMembersResponse)
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
		x := input.Message.Interface().(*MsgUpdateGroupMembersResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMembersResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMembersResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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
	md_MsgUpdateGroupAdminRequest           protoreflect.MessageDescriptor
	fd_MsgUpdateGroupAdminRequest_admin     protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAdminRequest_group_id  protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAdminRequest_new_admin protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAdminRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAdminRequest")
	fd_MsgUpdateGroupAdminRequest_admin = md_MsgUpdateGroupAdminRequest.Fields().ByName("admin")
	fd_MsgUpdateGroupAdminRequest_group_id = md_MsgUpdateGroupAdminRequest.Fields().ByName("group_id")
	fd_MsgUpdateGroupAdminRequest_new_admin = md_MsgUpdateGroupAdminRequest.Fields().ByName("new_admin")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAdminRequest)(nil)

type fastReflection_MsgUpdateGroupAdminRequest MsgUpdateGroupAdminRequest

func (x *MsgUpdateGroupAdminRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAdminRequest)(x)
}

func (x *MsgUpdateGroupAdminRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAdminRequest_messageType fastReflection_MsgUpdateGroupAdminRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAdminRequest_messageType{}

type fastReflection_MsgUpdateGroupAdminRequest_messageType struct{}

func (x fastReflection_MsgUpdateGroupAdminRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAdminRequest)(nil)
}
func (x fastReflection_MsgUpdateGroupAdminRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAdminRequest)
}
func (x fastReflection_MsgUpdateGroupAdminRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAdminRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAdminRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAdminRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAdminRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAdminRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAdminRequest) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAdminRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAdminRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAdminRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAdminRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgUpdateGroupAdminRequest_admin, value) {
			return
		}
	}
	if x.GroupId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.GroupId)
		if !f(fd_MsgUpdateGroupAdminRequest_group_id, value) {
			return
		}
	}
	if x.NewAdmin != "" {
		value := protoreflect.ValueOfString(x.NewAdmin)
		if !f(fd_MsgUpdateGroupAdminRequest_new_admin, value) {
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
func (x *fastReflection_MsgUpdateGroupAdminRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.group_id":
		return x.GroupId != uint64(0)
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.new_admin":
		return x.NewAdmin != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAdminRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.group_id":
		x.GroupId = uint64(0)
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.new_admin":
		x.NewAdmin = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAdminRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.group_id":
		value := x.GroupId
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.new_admin":
		value := x.NewAdmin
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAdminRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.group_id":
		x.GroupId = value.Uint()
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.new_admin":
		x.NewAdmin = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAdminRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.group_id":
		panic(fmt.Errorf("field group_id of message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.new_admin":
		panic(fmt.Errorf("field new_admin of message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAdminRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.group_id":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.v1beta1.MsgUpdateGroupAdminRequest.new_admin":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAdminRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAdminRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAdminRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAdminRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAdminRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAdminRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAdminRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.GroupId != 0 {
			n += 1 + runtime.Sov(uint64(x.GroupId))
		}
		l = len(x.NewAdmin)
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
		x := input.Message.Interface().(*MsgUpdateGroupAdminRequest)
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
		if len(x.NewAdmin) > 0 {
			i -= len(x.NewAdmin)
			copy(dAtA[i:], x.NewAdmin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.NewAdmin)))
			i--
			dAtA[i] = 0x1a
		}
		if x.GroupId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.GroupId))
			i--
			dAtA[i] = 0x10
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgUpdateGroupAdminRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAdminRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAdminRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
				}
				x.GroupId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.GroupId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field NewAdmin", wireType)
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
				x.NewAdmin = string(dAtA[iNdEx:postIndex])
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

var (
	md_MsgUpdateGroupAdminResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAdminResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAdminResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAdminResponse)(nil)

type fastReflection_MsgUpdateGroupAdminResponse MsgUpdateGroupAdminResponse

func (x *MsgUpdateGroupAdminResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAdminResponse)(x)
}

func (x *MsgUpdateGroupAdminResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAdminResponse_messageType fastReflection_MsgUpdateGroupAdminResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAdminResponse_messageType{}

type fastReflection_MsgUpdateGroupAdminResponse_messageType struct{}

func (x fastReflection_MsgUpdateGroupAdminResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAdminResponse)(nil)
}
func (x fastReflection_MsgUpdateGroupAdminResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAdminResponse)
}
func (x fastReflection_MsgUpdateGroupAdminResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAdminResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAdminResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAdminResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAdminResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAdminResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAdminResponse) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAdminResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAdminResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAdminResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAdminResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgUpdateGroupAdminResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAdminResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAdminResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAdminResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAdminResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAdminResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAdminResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAdminResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAdminResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAdminResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAdminResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAdminResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAdminResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAdminResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgUpdateGroupAdminResponse)
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
		x := input.Message.Interface().(*MsgUpdateGroupAdminResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAdminResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAdminResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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
	md_MsgUpdateGroupMetadataRequest          protoreflect.MessageDescriptor
	fd_MsgUpdateGroupMetadataRequest_admin    protoreflect.FieldDescriptor
	fd_MsgUpdateGroupMetadataRequest_group_id protoreflect.FieldDescriptor
	fd_MsgUpdateGroupMetadataRequest_metadata protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupMetadataRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupMetadataRequest")
	fd_MsgUpdateGroupMetadataRequest_admin = md_MsgUpdateGroupMetadataRequest.Fields().ByName("admin")
	fd_MsgUpdateGroupMetadataRequest_group_id = md_MsgUpdateGroupMetadataRequest.Fields().ByName("group_id")
	fd_MsgUpdateGroupMetadataRequest_metadata = md_MsgUpdateGroupMetadataRequest.Fields().ByName("metadata")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupMetadataRequest)(nil)

type fastReflection_MsgUpdateGroupMetadataRequest MsgUpdateGroupMetadataRequest

func (x *MsgUpdateGroupMetadataRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMetadataRequest)(x)
}

func (x *MsgUpdateGroupMetadataRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupMetadataRequest_messageType fastReflection_MsgUpdateGroupMetadataRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupMetadataRequest_messageType{}

type fastReflection_MsgUpdateGroupMetadataRequest_messageType struct{}

func (x fastReflection_MsgUpdateGroupMetadataRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMetadataRequest)(nil)
}
func (x fastReflection_MsgUpdateGroupMetadataRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMetadataRequest)
}
func (x fastReflection_MsgUpdateGroupMetadataRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMetadataRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMetadataRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupMetadataRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMetadataRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupMetadataRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgUpdateGroupMetadataRequest_admin, value) {
			return
		}
	}
	if x.GroupId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.GroupId)
		if !f(fd_MsgUpdateGroupMetadataRequest_group_id, value) {
			return
		}
	}
	if len(x.Metadata) != 0 {
		value := protoreflect.ValueOfBytes(x.Metadata)
		if !f(fd_MsgUpdateGroupMetadataRequest_metadata, value) {
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
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.group_id":
		return x.GroupId != uint64(0)
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.metadata":
		return len(x.Metadata) != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.group_id":
		x.GroupId = uint64(0)
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.metadata":
		x.Metadata = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.group_id":
		value := x.GroupId
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.metadata":
		value := x.Metadata
		return protoreflect.ValueOfBytes(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.group_id":
		x.GroupId = value.Uint()
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.metadata":
		x.Metadata = value.Bytes()
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMetadataRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.group_id":
		panic(fmt.Errorf("field group_id of message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.metadata":
		panic(fmt.Errorf("field metadata of message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.group_id":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest.metadata":
		return protoreflect.ValueOfBytes(nil)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupMetadataRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupMetadataRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupMetadataRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.GroupId != 0 {
			n += 1 + runtime.Sov(uint64(x.GroupId))
		}
		l = len(x.Metadata)
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
		x := input.Message.Interface().(*MsgUpdateGroupMetadataRequest)
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
		if len(x.Metadata) > 0 {
			i -= len(x.Metadata)
			copy(dAtA[i:], x.Metadata)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Metadata)))
			i--
			dAtA[i] = 0x1a
		}
		if x.GroupId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.GroupId))
			i--
			dAtA[i] = 0x10
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgUpdateGroupMetadataRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMetadataRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMetadataRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
				}
				x.GroupId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.GroupId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
				x.Metadata = append(x.Metadata[:0], dAtA[iNdEx:postIndex]...)
				if x.Metadata == nil {
					x.Metadata = []byte{}
				}
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

var (
	md_MsgUpdateGroupMetadataResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupMetadataResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupMetadataResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupMetadataResponse)(nil)

type fastReflection_MsgUpdateGroupMetadataResponse MsgUpdateGroupMetadataResponse

func (x *MsgUpdateGroupMetadataResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMetadataResponse)(x)
}

func (x *MsgUpdateGroupMetadataResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupMetadataResponse_messageType fastReflection_MsgUpdateGroupMetadataResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupMetadataResponse_messageType{}

type fastReflection_MsgUpdateGroupMetadataResponse_messageType struct{}

func (x fastReflection_MsgUpdateGroupMetadataResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupMetadataResponse)(nil)
}
func (x fastReflection_MsgUpdateGroupMetadataResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMetadataResponse)
}
func (x fastReflection_MsgUpdateGroupMetadataResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMetadataResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupMetadataResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupMetadataResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupMetadataResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupMetadataResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupMetadataResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupMetadataResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupMetadataResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupMetadataResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgUpdateGroupMetadataResponse)
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
		x := input.Message.Interface().(*MsgUpdateGroupMetadataResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMetadataResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupMetadataResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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
	md_MsgCreateGroupAccountRequest                 protoreflect.MessageDescriptor
	fd_MsgCreateGroupAccountRequest_admin           protoreflect.FieldDescriptor
	fd_MsgCreateGroupAccountRequest_group_id        protoreflect.FieldDescriptor
	fd_MsgCreateGroupAccountRequest_metadata        protoreflect.FieldDescriptor
	fd_MsgCreateGroupAccountRequest_decision_policy protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgCreateGroupAccountRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgCreateGroupAccountRequest")
	fd_MsgCreateGroupAccountRequest_admin = md_MsgCreateGroupAccountRequest.Fields().ByName("admin")
	fd_MsgCreateGroupAccountRequest_group_id = md_MsgCreateGroupAccountRequest.Fields().ByName("group_id")
	fd_MsgCreateGroupAccountRequest_metadata = md_MsgCreateGroupAccountRequest.Fields().ByName("metadata")
	fd_MsgCreateGroupAccountRequest_decision_policy = md_MsgCreateGroupAccountRequest.Fields().ByName("decision_policy")
}

var _ protoreflect.Message = (*fastReflection_MsgCreateGroupAccountRequest)(nil)

type fastReflection_MsgCreateGroupAccountRequest MsgCreateGroupAccountRequest

func (x *MsgCreateGroupAccountRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupAccountRequest)(x)
}

func (x *MsgCreateGroupAccountRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgCreateGroupAccountRequest_messageType fastReflection_MsgCreateGroupAccountRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgCreateGroupAccountRequest_messageType{}

type fastReflection_MsgCreateGroupAccountRequest_messageType struct{}

func (x fastReflection_MsgCreateGroupAccountRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupAccountRequest)(nil)
}
func (x fastReflection_MsgCreateGroupAccountRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupAccountRequest)
}
func (x fastReflection_MsgCreateGroupAccountRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupAccountRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgCreateGroupAccountRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupAccountRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgCreateGroupAccountRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgCreateGroupAccountRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgCreateGroupAccountRequest) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupAccountRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgCreateGroupAccountRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgCreateGroupAccountRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgCreateGroupAccountRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgCreateGroupAccountRequest_admin, value) {
			return
		}
	}
	if x.GroupId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.GroupId)
		if !f(fd_MsgCreateGroupAccountRequest_group_id, value) {
			return
		}
	}
	if len(x.Metadata) != 0 {
		value := protoreflect.ValueOfBytes(x.Metadata)
		if !f(fd_MsgCreateGroupAccountRequest_metadata, value) {
			return
		}
	}
	if x.DecisionPolicy != nil {
		value := protoreflect.ValueOfMessage(x.DecisionPolicy.ProtoReflect())
		if !f(fd_MsgCreateGroupAccountRequest_decision_policy, value) {
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
func (x *fastReflection_MsgCreateGroupAccountRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.group_id":
		return x.GroupId != uint64(0)
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.metadata":
		return len(x.Metadata) != 0
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.decision_policy":
		return x.DecisionPolicy != nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupAccountRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.group_id":
		x.GroupId = uint64(0)
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.metadata":
		x.Metadata = nil
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.decision_policy":
		x.DecisionPolicy = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgCreateGroupAccountRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.group_id":
		value := x.GroupId
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.metadata":
		value := x.Metadata
		return protoreflect.ValueOfBytes(value)
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.decision_policy":
		value := x.DecisionPolicy
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgCreateGroupAccountRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.group_id":
		x.GroupId = value.Uint()
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.metadata":
		x.Metadata = value.Bytes()
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.decision_policy":
		x.DecisionPolicy = value.Message().Interface().(*anypb.Any)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgCreateGroupAccountRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.decision_policy":
		if x.DecisionPolicy == nil {
			x.DecisionPolicy = new(anypb.Any)
		}
		return protoreflect.ValueOfMessage(x.DecisionPolicy.ProtoReflect())
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgCreateGroupAccountRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.group_id":
		panic(fmt.Errorf("field group_id of message cosmos.group.v1beta1.MsgCreateGroupAccountRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.metadata":
		panic(fmt.Errorf("field metadata of message cosmos.group.v1beta1.MsgCreateGroupAccountRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgCreateGroupAccountRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.group_id":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.metadata":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.group.v1beta1.MsgCreateGroupAccountRequest.decision_policy":
		m := new(anypb.Any)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgCreateGroupAccountRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgCreateGroupAccountRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgCreateGroupAccountRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupAccountRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgCreateGroupAccountRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgCreateGroupAccountRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgCreateGroupAccountRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.GroupId != 0 {
			n += 1 + runtime.Sov(uint64(x.GroupId))
		}
		l = len(x.Metadata)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.DecisionPolicy != nil {
			l = options.Size(x.DecisionPolicy)
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
		x := input.Message.Interface().(*MsgCreateGroupAccountRequest)
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
		if x.DecisionPolicy != nil {
			encoded, err := options.Marshal(x.DecisionPolicy)
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
			dAtA[i] = 0x22
		}
		if len(x.Metadata) > 0 {
			i -= len(x.Metadata)
			copy(dAtA[i:], x.Metadata)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Metadata)))
			i--
			dAtA[i] = 0x1a
		}
		if x.GroupId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.GroupId))
			i--
			dAtA[i] = 0x10
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgCreateGroupAccountRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupAccountRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupAccountRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field GroupId", wireType)
				}
				x.GroupId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.GroupId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
				x.Metadata = append(x.Metadata[:0], dAtA[iNdEx:postIndex]...)
				if x.Metadata == nil {
					x.Metadata = []byte{}
				}
				iNdEx = postIndex
			case 4:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field DecisionPolicy", wireType)
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
				if x.DecisionPolicy == nil {
					x.DecisionPolicy = &anypb.Any{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.DecisionPolicy); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
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

var (
	md_MsgCreateGroupAccountResponse         protoreflect.MessageDescriptor
	fd_MsgCreateGroupAccountResponse_address protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgCreateGroupAccountResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgCreateGroupAccountResponse")
	fd_MsgCreateGroupAccountResponse_address = md_MsgCreateGroupAccountResponse.Fields().ByName("address")
}

var _ protoreflect.Message = (*fastReflection_MsgCreateGroupAccountResponse)(nil)

type fastReflection_MsgCreateGroupAccountResponse MsgCreateGroupAccountResponse

func (x *MsgCreateGroupAccountResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupAccountResponse)(x)
}

func (x *MsgCreateGroupAccountResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgCreateGroupAccountResponse_messageType fastReflection_MsgCreateGroupAccountResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgCreateGroupAccountResponse_messageType{}

type fastReflection_MsgCreateGroupAccountResponse_messageType struct{}

func (x fastReflection_MsgCreateGroupAccountResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgCreateGroupAccountResponse)(nil)
}
func (x fastReflection_MsgCreateGroupAccountResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupAccountResponse)
}
func (x fastReflection_MsgCreateGroupAccountResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupAccountResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgCreateGroupAccountResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateGroupAccountResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgCreateGroupAccountResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgCreateGroupAccountResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgCreateGroupAccountResponse) New() protoreflect.Message {
	return new(fastReflection_MsgCreateGroupAccountResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgCreateGroupAccountResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgCreateGroupAccountResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgCreateGroupAccountResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Address != "" {
		value := protoreflect.ValueOfString(x.Address)
		if !f(fd_MsgCreateGroupAccountResponse_address, value) {
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
func (x *fastReflection_MsgCreateGroupAccountResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountResponse.address":
		return x.Address != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupAccountResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountResponse.address":
		x.Address = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgCreateGroupAccountResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountResponse.address":
		value := x.Address
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgCreateGroupAccountResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountResponse.address":
		x.Address = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgCreateGroupAccountResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountResponse.address":
		panic(fmt.Errorf("field address of message cosmos.group.v1beta1.MsgCreateGroupAccountResponse is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgCreateGroupAccountResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateGroupAccountResponse.address":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateGroupAccountResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateGroupAccountResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgCreateGroupAccountResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgCreateGroupAccountResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgCreateGroupAccountResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateGroupAccountResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgCreateGroupAccountResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgCreateGroupAccountResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgCreateGroupAccountResponse)
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
		l = len(x.Address)
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
		x := input.Message.Interface().(*MsgCreateGroupAccountResponse)
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
		if len(x.Address) > 0 {
			i -= len(x.Address)
			copy(dAtA[i:], x.Address)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Address)))
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
		x := input.Message.Interface().(*MsgCreateGroupAccountResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupAccountResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateGroupAccountResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
				x.Address = string(dAtA[iNdEx:postIndex])
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

var (
	md_MsgUpdateGroupAccountAdminRequest           protoreflect.MessageDescriptor
	fd_MsgUpdateGroupAccountAdminRequest_admin     protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAccountAdminRequest_address   protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAccountAdminRequest_new_admin protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAccountAdminRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAccountAdminRequest")
	fd_MsgUpdateGroupAccountAdminRequest_admin = md_MsgUpdateGroupAccountAdminRequest.Fields().ByName("admin")
	fd_MsgUpdateGroupAccountAdminRequest_address = md_MsgUpdateGroupAccountAdminRequest.Fields().ByName("address")
	fd_MsgUpdateGroupAccountAdminRequest_new_admin = md_MsgUpdateGroupAccountAdminRequest.Fields().ByName("new_admin")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAccountAdminRequest)(nil)

type fastReflection_MsgUpdateGroupAccountAdminRequest MsgUpdateGroupAccountAdminRequest

func (x *MsgUpdateGroupAccountAdminRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountAdminRequest)(x)
}

func (x *MsgUpdateGroupAccountAdminRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAccountAdminRequest_messageType fastReflection_MsgUpdateGroupAccountAdminRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAccountAdminRequest_messageType{}

type fastReflection_MsgUpdateGroupAccountAdminRequest_messageType struct{}

func (x fastReflection_MsgUpdateGroupAccountAdminRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountAdminRequest)(nil)
}
func (x fastReflection_MsgUpdateGroupAccountAdminRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountAdminRequest)
}
func (x fastReflection_MsgUpdateGroupAccountAdminRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountAdminRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountAdminRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAccountAdminRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountAdminRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAccountAdminRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgUpdateGroupAccountAdminRequest_admin, value) {
			return
		}
	}
	if x.Address != "" {
		value := protoreflect.ValueOfString(x.Address)
		if !f(fd_MsgUpdateGroupAccountAdminRequest_address, value) {
			return
		}
	}
	if x.NewAdmin != "" {
		value := protoreflect.ValueOfString(x.NewAdmin)
		if !f(fd_MsgUpdateGroupAccountAdminRequest_new_admin, value) {
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
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.address":
		return x.Address != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.new_admin":
		return x.NewAdmin != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.address":
		x.Address = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.new_admin":
		x.NewAdmin = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.address":
		value := x.Address
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.new_admin":
		value := x.NewAdmin
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.address":
		x.Address = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.new_admin":
		x.NewAdmin = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.address":
		panic(fmt.Errorf("field address of message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.new_admin":
		panic(fmt.Errorf("field new_admin of message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.address":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest.new_admin":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAccountAdminRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAccountAdminRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Address)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.NewAdmin)
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountAdminRequest)
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
		if len(x.NewAdmin) > 0 {
			i -= len(x.NewAdmin)
			copy(dAtA[i:], x.NewAdmin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.NewAdmin)))
			i--
			dAtA[i] = 0x1a
		}
		if len(x.Address) > 0 {
			i -= len(x.Address)
			copy(dAtA[i:], x.Address)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Address)))
			i--
			dAtA[i] = 0x12
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountAdminRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountAdminRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountAdminRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
				x.Address = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field NewAdmin", wireType)
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
				x.NewAdmin = string(dAtA[iNdEx:postIndex])
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

var (
	md_MsgUpdateGroupAccountAdminResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAccountAdminResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAccountAdminResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAccountAdminResponse)(nil)

type fastReflection_MsgUpdateGroupAccountAdminResponse MsgUpdateGroupAccountAdminResponse

func (x *MsgUpdateGroupAccountAdminResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountAdminResponse)(x)
}

func (x *MsgUpdateGroupAccountAdminResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAccountAdminResponse_messageType fastReflection_MsgUpdateGroupAccountAdminResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAccountAdminResponse_messageType{}

type fastReflection_MsgUpdateGroupAccountAdminResponse_messageType struct{}

func (x fastReflection_MsgUpdateGroupAccountAdminResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountAdminResponse)(nil)
}
func (x fastReflection_MsgUpdateGroupAccountAdminResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountAdminResponse)
}
func (x fastReflection_MsgUpdateGroupAccountAdminResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountAdminResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountAdminResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAccountAdminResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountAdminResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAccountAdminResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAccountAdminResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAccountAdminResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgUpdateGroupAccountAdminResponse)
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountAdminResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountAdminResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountAdminResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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
	md_MsgUpdateGroupAccountDecisionPolicyRequest                 protoreflect.MessageDescriptor
	fd_MsgUpdateGroupAccountDecisionPolicyRequest_admin           protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAccountDecisionPolicyRequest_address         protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAccountDecisionPolicyRequest_decision_policy protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAccountDecisionPolicyRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAccountDecisionPolicyRequest")
	fd_MsgUpdateGroupAccountDecisionPolicyRequest_admin = md_MsgUpdateGroupAccountDecisionPolicyRequest.Fields().ByName("admin")
	fd_MsgUpdateGroupAccountDecisionPolicyRequest_address = md_MsgUpdateGroupAccountDecisionPolicyRequest.Fields().ByName("address")
	fd_MsgUpdateGroupAccountDecisionPolicyRequest_decision_policy = md_MsgUpdateGroupAccountDecisionPolicyRequest.Fields().ByName("decision_policy")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest)(nil)

type fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest MsgUpdateGroupAccountDecisionPolicyRequest

func (x *MsgUpdateGroupAccountDecisionPolicyRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest)(x)
}

func (x *MsgUpdateGroupAccountDecisionPolicyRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType{}

type fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType struct{}

func (x fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest)(nil)
}
func (x fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest)
}
func (x fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountDecisionPolicyRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountDecisionPolicyRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAccountDecisionPolicyRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgUpdateGroupAccountDecisionPolicyRequest_admin, value) {
			return
		}
	}
	if x.Address != "" {
		value := protoreflect.ValueOfString(x.Address)
		if !f(fd_MsgUpdateGroupAccountDecisionPolicyRequest_address, value) {
			return
		}
	}
	if x.DecisionPolicy != nil {
		value := protoreflect.ValueOfMessage(x.DecisionPolicy.ProtoReflect())
		if !f(fd_MsgUpdateGroupAccountDecisionPolicyRequest_decision_policy, value) {
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.address":
		return x.Address != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.decision_policy":
		return x.DecisionPolicy != nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.address":
		x.Address = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.decision_policy":
		x.DecisionPolicy = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.address":
		value := x.Address
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.decision_policy":
		value := x.DecisionPolicy
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.address":
		x.Address = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.decision_policy":
		x.DecisionPolicy = value.Message().Interface().(*anypb.Any)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.decision_policy":
		if x.DecisionPolicy == nil {
			x.DecisionPolicy = new(anypb.Any)
		}
		return protoreflect.ValueOfMessage(x.DecisionPolicy.ProtoReflect())
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.address":
		panic(fmt.Errorf("field address of message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.address":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.decision_policy":
		m := new(anypb.Any)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAccountDecisionPolicyRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Address)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.DecisionPolicy != nil {
			l = options.Size(x.DecisionPolicy)
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountDecisionPolicyRequest)
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
		if x.DecisionPolicy != nil {
			encoded, err := options.Marshal(x.DecisionPolicy)
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
			dAtA[i] = 0x1a
		}
		if len(x.Address) > 0 {
			i -= len(x.Address)
			copy(dAtA[i:], x.Address)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Address)))
			i--
			dAtA[i] = 0x12
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountDecisionPolicyRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountDecisionPolicyRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountDecisionPolicyRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
				x.Address = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field DecisionPolicy", wireType)
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
				if x.DecisionPolicy == nil {
					x.DecisionPolicy = &anypb.Any{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.DecisionPolicy); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
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

var (
	md_MsgUpdateGroupAccountDecisionPolicyResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAccountDecisionPolicyResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAccountDecisionPolicyResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse)(nil)

type fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse MsgUpdateGroupAccountDecisionPolicyResponse

func (x *MsgUpdateGroupAccountDecisionPolicyResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse)(x)
}

func (x *MsgUpdateGroupAccountDecisionPolicyResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType{}

type fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType struct{}

func (x fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse)(nil)
}
func (x fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse)
}
func (x fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountDecisionPolicyResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountDecisionPolicyResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAccountDecisionPolicyResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAccountDecisionPolicyResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAccountDecisionPolicyResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgUpdateGroupAccountDecisionPolicyResponse)
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountDecisionPolicyResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountDecisionPolicyResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountDecisionPolicyResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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
	md_MsgUpdateGroupAccountMetadataRequest          protoreflect.MessageDescriptor
	fd_MsgUpdateGroupAccountMetadataRequest_admin    protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAccountMetadataRequest_address  protoreflect.FieldDescriptor
	fd_MsgUpdateGroupAccountMetadataRequest_metadata protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAccountMetadataRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAccountMetadataRequest")
	fd_MsgUpdateGroupAccountMetadataRequest_admin = md_MsgUpdateGroupAccountMetadataRequest.Fields().ByName("admin")
	fd_MsgUpdateGroupAccountMetadataRequest_address = md_MsgUpdateGroupAccountMetadataRequest.Fields().ByName("address")
	fd_MsgUpdateGroupAccountMetadataRequest_metadata = md_MsgUpdateGroupAccountMetadataRequest.Fields().ByName("metadata")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAccountMetadataRequest)(nil)

type fastReflection_MsgUpdateGroupAccountMetadataRequest MsgUpdateGroupAccountMetadataRequest

func (x *MsgUpdateGroupAccountMetadataRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountMetadataRequest)(x)
}

func (x *MsgUpdateGroupAccountMetadataRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType{}

type fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType struct{}

func (x fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountMetadataRequest)(nil)
}
func (x fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountMetadataRequest)
}
func (x fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountMetadataRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountMetadataRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAccountMetadataRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountMetadataRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAccountMetadataRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Admin != "" {
		value := protoreflect.ValueOfString(x.Admin)
		if !f(fd_MsgUpdateGroupAccountMetadataRequest_admin, value) {
			return
		}
	}
	if x.Address != "" {
		value := protoreflect.ValueOfString(x.Address)
		if !f(fd_MsgUpdateGroupAccountMetadataRequest_address, value) {
			return
		}
	}
	if len(x.Metadata) != 0 {
		value := protoreflect.ValueOfBytes(x.Metadata)
		if !f(fd_MsgUpdateGroupAccountMetadataRequest_metadata, value) {
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.admin":
		return x.Admin != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.address":
		return x.Address != ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.metadata":
		return len(x.Metadata) != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.admin":
		x.Admin = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.address":
		x.Address = ""
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.metadata":
		x.Metadata = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.admin":
		value := x.Admin
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.address":
		value := x.Address
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.metadata":
		value := x.Metadata
		return protoreflect.ValueOfBytes(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.admin":
		x.Admin = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.address":
		x.Address = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.metadata":
		x.Metadata = value.Bytes()
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.admin":
		panic(fmt.Errorf("field admin of message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.address":
		panic(fmt.Errorf("field address of message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.metadata":
		panic(fmt.Errorf("field metadata of message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.admin":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.address":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest.metadata":
		return protoreflect.ValueOfBytes(nil)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAccountMetadataRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAccountMetadataRequest)
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
		l = len(x.Admin)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Address)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Metadata)
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountMetadataRequest)
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
		if len(x.Metadata) > 0 {
			i -= len(x.Metadata)
			copy(dAtA[i:], x.Metadata)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Metadata)))
			i--
			dAtA[i] = 0x1a
		}
		if len(x.Address) > 0 {
			i -= len(x.Address)
			copy(dAtA[i:], x.Address)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Address)))
			i--
			dAtA[i] = 0x12
		}
		if len(x.Admin) > 0 {
			i -= len(x.Admin)
			copy(dAtA[i:], x.Admin)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Admin)))
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountMetadataRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountMetadataRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountMetadataRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Admin", wireType)
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
				x.Admin = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
				x.Address = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
				x.Metadata = append(x.Metadata[:0], dAtA[iNdEx:postIndex]...)
				if x.Metadata == nil {
					x.Metadata = []byte{}
				}
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

var (
	md_MsgUpdateGroupAccountMetadataResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgUpdateGroupAccountMetadataResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgUpdateGroupAccountMetadataResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgUpdateGroupAccountMetadataResponse)(nil)

type fastReflection_MsgUpdateGroupAccountMetadataResponse MsgUpdateGroupAccountMetadataResponse

func (x *MsgUpdateGroupAccountMetadataResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountMetadataResponse)(x)
}

func (x *MsgUpdateGroupAccountMetadataResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[15]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType{}

type fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType struct{}

func (x fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgUpdateGroupAccountMetadataResponse)(nil)
}
func (x fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountMetadataResponse)
}
func (x fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountMetadataResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgUpdateGroupAccountMetadataResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgUpdateGroupAccountMetadataResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) New() protoreflect.Message {
	return new(fastReflection_MsgUpdateGroupAccountMetadataResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgUpdateGroupAccountMetadataResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgUpdateGroupAccountMetadataResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgUpdateGroupAccountMetadataResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgUpdateGroupAccountMetadataResponse)
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
		x := input.Message.Interface().(*MsgUpdateGroupAccountMetadataResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountMetadataResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgUpdateGroupAccountMetadataResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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

var _ protoreflect.List = (*_MsgCreateProposalRequest_2_list)(nil)

type _MsgCreateProposalRequest_2_list struct {
	list *[]string
}

func (x *_MsgCreateProposalRequest_2_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_MsgCreateProposalRequest_2_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfString((*x.list)[i])
}

func (x *_MsgCreateProposalRequest_2_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.String()
	concreteValue := valueUnwrapped
	(*x.list)[i] = concreteValue
}

func (x *_MsgCreateProposalRequest_2_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.String()
	concreteValue := valueUnwrapped
	*x.list = append(*x.list, concreteValue)
}

func (x *_MsgCreateProposalRequest_2_list) AppendMutable() protoreflect.Value {
	panic(fmt.Errorf("AppendMutable can not be called on message MsgCreateProposalRequest at list field Proposers as it is not of Message kind"))
}

func (x *_MsgCreateProposalRequest_2_list) Truncate(n int) {
	*x.list = (*x.list)[:n]
}

func (x *_MsgCreateProposalRequest_2_list) NewElement() protoreflect.Value {
	v := ""
	return protoreflect.ValueOfString(v)
}

func (x *_MsgCreateProposalRequest_2_list) IsValid() bool {
	return x.list != nil
}

var _ protoreflect.List = (*_MsgCreateProposalRequest_4_list)(nil)

type _MsgCreateProposalRequest_4_list struct {
	list *[]*anypb.Any
}

func (x *_MsgCreateProposalRequest_4_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_MsgCreateProposalRequest_4_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfMessage((*x.list)[i].ProtoReflect())
}

func (x *_MsgCreateProposalRequest_4_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*anypb.Any)
	(*x.list)[i] = concreteValue
}

func (x *_MsgCreateProposalRequest_4_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*anypb.Any)
	*x.list = append(*x.list, concreteValue)
}

func (x *_MsgCreateProposalRequest_4_list) AppendMutable() protoreflect.Value {
	v := new(anypb.Any)
	*x.list = append(*x.list, v)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_MsgCreateProposalRequest_4_list) Truncate(n int) {
	for i := n; i < len(*x.list); i++ {
		(*x.list)[i] = nil
	}
	*x.list = (*x.list)[:n]
}

func (x *_MsgCreateProposalRequest_4_list) NewElement() protoreflect.Value {
	v := new(anypb.Any)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_MsgCreateProposalRequest_4_list) IsValid() bool {
	return x.list != nil
}

var (
	md_MsgCreateProposalRequest           protoreflect.MessageDescriptor
	fd_MsgCreateProposalRequest_address   protoreflect.FieldDescriptor
	fd_MsgCreateProposalRequest_proposers protoreflect.FieldDescriptor
	fd_MsgCreateProposalRequest_metadata  protoreflect.FieldDescriptor
	fd_MsgCreateProposalRequest_msgs      protoreflect.FieldDescriptor
	fd_MsgCreateProposalRequest_exec      protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgCreateProposalRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgCreateProposalRequest")
	fd_MsgCreateProposalRequest_address = md_MsgCreateProposalRequest.Fields().ByName("address")
	fd_MsgCreateProposalRequest_proposers = md_MsgCreateProposalRequest.Fields().ByName("proposers")
	fd_MsgCreateProposalRequest_metadata = md_MsgCreateProposalRequest.Fields().ByName("metadata")
	fd_MsgCreateProposalRequest_msgs = md_MsgCreateProposalRequest.Fields().ByName("msgs")
	fd_MsgCreateProposalRequest_exec = md_MsgCreateProposalRequest.Fields().ByName("exec")
}

var _ protoreflect.Message = (*fastReflection_MsgCreateProposalRequest)(nil)

type fastReflection_MsgCreateProposalRequest MsgCreateProposalRequest

func (x *MsgCreateProposalRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgCreateProposalRequest)(x)
}

func (x *MsgCreateProposalRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[16]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgCreateProposalRequest_messageType fastReflection_MsgCreateProposalRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgCreateProposalRequest_messageType{}

type fastReflection_MsgCreateProposalRequest_messageType struct{}

func (x fastReflection_MsgCreateProposalRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgCreateProposalRequest)(nil)
}
func (x fastReflection_MsgCreateProposalRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgCreateProposalRequest)
}
func (x fastReflection_MsgCreateProposalRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateProposalRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgCreateProposalRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateProposalRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgCreateProposalRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgCreateProposalRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgCreateProposalRequest) New() protoreflect.Message {
	return new(fastReflection_MsgCreateProposalRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgCreateProposalRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgCreateProposalRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgCreateProposalRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Address != "" {
		value := protoreflect.ValueOfString(x.Address)
		if !f(fd_MsgCreateProposalRequest_address, value) {
			return
		}
	}
	if len(x.Proposers) != 0 {
		value := protoreflect.ValueOfList(&_MsgCreateProposalRequest_2_list{list: &x.Proposers})
		if !f(fd_MsgCreateProposalRequest_proposers, value) {
			return
		}
	}
	if len(x.Metadata) != 0 {
		value := protoreflect.ValueOfBytes(x.Metadata)
		if !f(fd_MsgCreateProposalRequest_metadata, value) {
			return
		}
	}
	if len(x.Msgs) != 0 {
		value := protoreflect.ValueOfList(&_MsgCreateProposalRequest_4_list{list: &x.Msgs})
		if !f(fd_MsgCreateProposalRequest_msgs, value) {
			return
		}
	}
	if x.Exec != 0 {
		value := protoreflect.ValueOfEnum((protoreflect.EnumNumber)(x.Exec))
		if !f(fd_MsgCreateProposalRequest_exec, value) {
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
func (x *fastReflection_MsgCreateProposalRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.address":
		return x.Address != ""
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.proposers":
		return len(x.Proposers) != 0
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.metadata":
		return len(x.Metadata) != 0
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.msgs":
		return len(x.Msgs) != 0
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.exec":
		return x.Exec != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateProposalRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.address":
		x.Address = ""
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.proposers":
		x.Proposers = nil
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.metadata":
		x.Metadata = nil
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.msgs":
		x.Msgs = nil
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.exec":
		x.Exec = 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgCreateProposalRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.address":
		value := x.Address
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.proposers":
		if len(x.Proposers) == 0 {
			return protoreflect.ValueOfList(&_MsgCreateProposalRequest_2_list{})
		}
		listValue := &_MsgCreateProposalRequest_2_list{list: &x.Proposers}
		return protoreflect.ValueOfList(listValue)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.metadata":
		value := x.Metadata
		return protoreflect.ValueOfBytes(value)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.msgs":
		if len(x.Msgs) == 0 {
			return protoreflect.ValueOfList(&_MsgCreateProposalRequest_4_list{})
		}
		listValue := &_MsgCreateProposalRequest_4_list{list: &x.Msgs}
		return protoreflect.ValueOfList(listValue)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.exec":
		value := x.Exec
		return protoreflect.ValueOfEnum((protoreflect.EnumNumber)(value))
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgCreateProposalRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.address":
		x.Address = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.proposers":
		lv := value.List()
		clv := lv.(*_MsgCreateProposalRequest_2_list)
		x.Proposers = *clv.list
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.metadata":
		x.Metadata = value.Bytes()
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.msgs":
		lv := value.List()
		clv := lv.(*_MsgCreateProposalRequest_4_list)
		x.Msgs = *clv.list
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.exec":
		x.Exec = (Exec)(value.Enum())
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgCreateProposalRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.proposers":
		if x.Proposers == nil {
			x.Proposers = []string{}
		}
		value := &_MsgCreateProposalRequest_2_list{list: &x.Proposers}
		return protoreflect.ValueOfList(value)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.msgs":
		if x.Msgs == nil {
			x.Msgs = []*anypb.Any{}
		}
		value := &_MsgCreateProposalRequest_4_list{list: &x.Msgs}
		return protoreflect.ValueOfList(value)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.address":
		panic(fmt.Errorf("field address of message cosmos.group.v1beta1.MsgCreateProposalRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.metadata":
		panic(fmt.Errorf("field metadata of message cosmos.group.v1beta1.MsgCreateProposalRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.exec":
		panic(fmt.Errorf("field exec of message cosmos.group.v1beta1.MsgCreateProposalRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgCreateProposalRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.address":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.proposers":
		list := []string{}
		return protoreflect.ValueOfList(&_MsgCreateProposalRequest_2_list{list: &list})
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.metadata":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.msgs":
		list := []*anypb.Any{}
		return protoreflect.ValueOfList(&_MsgCreateProposalRequest_4_list{list: &list})
	case "cosmos.group.v1beta1.MsgCreateProposalRequest.exec":
		return protoreflect.ValueOfEnum(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgCreateProposalRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgCreateProposalRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgCreateProposalRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateProposalRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgCreateProposalRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgCreateProposalRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgCreateProposalRequest)
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
		l = len(x.Address)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if len(x.Proposers) > 0 {
			for _, s := range x.Proposers {
				l = len(s)
				n += 1 + l + runtime.Sov(uint64(l))
			}
		}
		l = len(x.Metadata)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if len(x.Msgs) > 0 {
			for _, e := range x.Msgs {
				l = options.Size(e)
				n += 1 + l + runtime.Sov(uint64(l))
			}
		}
		if x.Exec != 0 {
			n += 1 + runtime.Sov(uint64(x.Exec))
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
		x := input.Message.Interface().(*MsgCreateProposalRequest)
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
		if x.Exec != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Exec))
			i--
			dAtA[i] = 0x28
		}
		if len(x.Msgs) > 0 {
			for iNdEx := len(x.Msgs) - 1; iNdEx >= 0; iNdEx-- {
				encoded, err := options.Marshal(x.Msgs[iNdEx])
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
				dAtA[i] = 0x22
			}
		}
		if len(x.Metadata) > 0 {
			i -= len(x.Metadata)
			copy(dAtA[i:], x.Metadata)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Metadata)))
			i--
			dAtA[i] = 0x1a
		}
		if len(x.Proposers) > 0 {
			for iNdEx := len(x.Proposers) - 1; iNdEx >= 0; iNdEx-- {
				i -= len(x.Proposers[iNdEx])
				copy(dAtA[i:], x.Proposers[iNdEx])
				i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Proposers[iNdEx])))
				i--
				dAtA[i] = 0x12
			}
		}
		if len(x.Address) > 0 {
			i -= len(x.Address)
			copy(dAtA[i:], x.Address)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Address)))
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
		x := input.Message.Interface().(*MsgCreateProposalRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateProposalRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateProposalRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
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
				x.Address = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Proposers", wireType)
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
				x.Proposers = append(x.Proposers, string(dAtA[iNdEx:postIndex]))
				iNdEx = postIndex
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
				x.Metadata = append(x.Metadata[:0], dAtA[iNdEx:postIndex]...)
				if x.Metadata == nil {
					x.Metadata = []byte{}
				}
				iNdEx = postIndex
			case 4:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Msgs", wireType)
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
				x.Msgs = append(x.Msgs, &anypb.Any{})
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Msgs[len(x.Msgs)-1]); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 5:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Exec", wireType)
				}
				x.Exec = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Exec |= Exec(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
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
	md_MsgCreateProposalResponse             protoreflect.MessageDescriptor
	fd_MsgCreateProposalResponse_proposal_id protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgCreateProposalResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgCreateProposalResponse")
	fd_MsgCreateProposalResponse_proposal_id = md_MsgCreateProposalResponse.Fields().ByName("proposal_id")
}

var _ protoreflect.Message = (*fastReflection_MsgCreateProposalResponse)(nil)

type fastReflection_MsgCreateProposalResponse MsgCreateProposalResponse

func (x *MsgCreateProposalResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgCreateProposalResponse)(x)
}

func (x *MsgCreateProposalResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[17]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgCreateProposalResponse_messageType fastReflection_MsgCreateProposalResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgCreateProposalResponse_messageType{}

type fastReflection_MsgCreateProposalResponse_messageType struct{}

func (x fastReflection_MsgCreateProposalResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgCreateProposalResponse)(nil)
}
func (x fastReflection_MsgCreateProposalResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgCreateProposalResponse)
}
func (x fastReflection_MsgCreateProposalResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateProposalResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgCreateProposalResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgCreateProposalResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgCreateProposalResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgCreateProposalResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgCreateProposalResponse) New() protoreflect.Message {
	return new(fastReflection_MsgCreateProposalResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgCreateProposalResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgCreateProposalResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgCreateProposalResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.ProposalId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.ProposalId)
		if !f(fd_MsgCreateProposalResponse_proposal_id, value) {
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
func (x *fastReflection_MsgCreateProposalResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalResponse.proposal_id":
		return x.ProposalId != uint64(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateProposalResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalResponse.proposal_id":
		x.ProposalId = uint64(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgCreateProposalResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalResponse.proposal_id":
		value := x.ProposalId
		return protoreflect.ValueOfUint64(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgCreateProposalResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalResponse.proposal_id":
		x.ProposalId = value.Uint()
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgCreateProposalResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalResponse.proposal_id":
		panic(fmt.Errorf("field proposal_id of message cosmos.group.v1beta1.MsgCreateProposalResponse is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgCreateProposalResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgCreateProposalResponse.proposal_id":
		return protoreflect.ValueOfUint64(uint64(0))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgCreateProposalResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgCreateProposalResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgCreateProposalResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgCreateProposalResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgCreateProposalResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgCreateProposalResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgCreateProposalResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgCreateProposalResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgCreateProposalResponse)
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
		if x.ProposalId != 0 {
			n += 1 + runtime.Sov(uint64(x.ProposalId))
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
		x := input.Message.Interface().(*MsgCreateProposalResponse)
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
		if x.ProposalId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.ProposalId))
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
		x := input.Message.Interface().(*MsgCreateProposalResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateProposalResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgCreateProposalResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
				}
				x.ProposalId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.ProposalId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
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
	md_MsgVoteRequest             protoreflect.MessageDescriptor
	fd_MsgVoteRequest_proposal_id protoreflect.FieldDescriptor
	fd_MsgVoteRequest_voter       protoreflect.FieldDescriptor
	fd_MsgVoteRequest_choice      protoreflect.FieldDescriptor
	fd_MsgVoteRequest_metadata    protoreflect.FieldDescriptor
	fd_MsgVoteRequest_exec        protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgVoteRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgVoteRequest")
	fd_MsgVoteRequest_proposal_id = md_MsgVoteRequest.Fields().ByName("proposal_id")
	fd_MsgVoteRequest_voter = md_MsgVoteRequest.Fields().ByName("voter")
	fd_MsgVoteRequest_choice = md_MsgVoteRequest.Fields().ByName("choice")
	fd_MsgVoteRequest_metadata = md_MsgVoteRequest.Fields().ByName("metadata")
	fd_MsgVoteRequest_exec = md_MsgVoteRequest.Fields().ByName("exec")
}

var _ protoreflect.Message = (*fastReflection_MsgVoteRequest)(nil)

type fastReflection_MsgVoteRequest MsgVoteRequest

func (x *MsgVoteRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgVoteRequest)(x)
}

func (x *MsgVoteRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[18]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgVoteRequest_messageType fastReflection_MsgVoteRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgVoteRequest_messageType{}

type fastReflection_MsgVoteRequest_messageType struct{}

func (x fastReflection_MsgVoteRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgVoteRequest)(nil)
}
func (x fastReflection_MsgVoteRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgVoteRequest)
}
func (x fastReflection_MsgVoteRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgVoteRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgVoteRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgVoteRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgVoteRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgVoteRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgVoteRequest) New() protoreflect.Message {
	return new(fastReflection_MsgVoteRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgVoteRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgVoteRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgVoteRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.ProposalId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.ProposalId)
		if !f(fd_MsgVoteRequest_proposal_id, value) {
			return
		}
	}
	if x.Voter != "" {
		value := protoreflect.ValueOfString(x.Voter)
		if !f(fd_MsgVoteRequest_voter, value) {
			return
		}
	}
	if x.Choice != 0 {
		value := protoreflect.ValueOfEnum((protoreflect.EnumNumber)(x.Choice))
		if !f(fd_MsgVoteRequest_choice, value) {
			return
		}
	}
	if len(x.Metadata) != 0 {
		value := protoreflect.ValueOfBytes(x.Metadata)
		if !f(fd_MsgVoteRequest_metadata, value) {
			return
		}
	}
	if x.Exec != 0 {
		value := protoreflect.ValueOfEnum((protoreflect.EnumNumber)(x.Exec))
		if !f(fd_MsgVoteRequest_exec, value) {
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
func (x *fastReflection_MsgVoteRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgVoteRequest.proposal_id":
		return x.ProposalId != uint64(0)
	case "cosmos.group.v1beta1.MsgVoteRequest.voter":
		return x.Voter != ""
	case "cosmos.group.v1beta1.MsgVoteRequest.choice":
		return x.Choice != 0
	case "cosmos.group.v1beta1.MsgVoteRequest.metadata":
		return len(x.Metadata) != 0
	case "cosmos.group.v1beta1.MsgVoteRequest.exec":
		return x.Exec != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgVoteRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgVoteRequest.proposal_id":
		x.ProposalId = uint64(0)
	case "cosmos.group.v1beta1.MsgVoteRequest.voter":
		x.Voter = ""
	case "cosmos.group.v1beta1.MsgVoteRequest.choice":
		x.Choice = 0
	case "cosmos.group.v1beta1.MsgVoteRequest.metadata":
		x.Metadata = nil
	case "cosmos.group.v1beta1.MsgVoteRequest.exec":
		x.Exec = 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgVoteRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgVoteRequest.proposal_id":
		value := x.ProposalId
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.v1beta1.MsgVoteRequest.voter":
		value := x.Voter
		return protoreflect.ValueOfString(value)
	case "cosmos.group.v1beta1.MsgVoteRequest.choice":
		value := x.Choice
		return protoreflect.ValueOfEnum((protoreflect.EnumNumber)(value))
	case "cosmos.group.v1beta1.MsgVoteRequest.metadata":
		value := x.Metadata
		return protoreflect.ValueOfBytes(value)
	case "cosmos.group.v1beta1.MsgVoteRequest.exec":
		value := x.Exec
		return protoreflect.ValueOfEnum((protoreflect.EnumNumber)(value))
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgVoteRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgVoteRequest.proposal_id":
		x.ProposalId = value.Uint()
	case "cosmos.group.v1beta1.MsgVoteRequest.voter":
		x.Voter = value.Interface().(string)
	case "cosmos.group.v1beta1.MsgVoteRequest.choice":
		x.Choice = (Choice)(value.Enum())
	case "cosmos.group.v1beta1.MsgVoteRequest.metadata":
		x.Metadata = value.Bytes()
	case "cosmos.group.v1beta1.MsgVoteRequest.exec":
		x.Exec = (Exec)(value.Enum())
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgVoteRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgVoteRequest.proposal_id":
		panic(fmt.Errorf("field proposal_id of message cosmos.group.v1beta1.MsgVoteRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgVoteRequest.voter":
		panic(fmt.Errorf("field voter of message cosmos.group.v1beta1.MsgVoteRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgVoteRequest.choice":
		panic(fmt.Errorf("field choice of message cosmos.group.v1beta1.MsgVoteRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgVoteRequest.metadata":
		panic(fmt.Errorf("field metadata of message cosmos.group.v1beta1.MsgVoteRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgVoteRequest.exec":
		panic(fmt.Errorf("field exec of message cosmos.group.v1beta1.MsgVoteRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgVoteRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgVoteRequest.proposal_id":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.v1beta1.MsgVoteRequest.voter":
		return protoreflect.ValueOfString("")
	case "cosmos.group.v1beta1.MsgVoteRequest.choice":
		return protoreflect.ValueOfEnum(0)
	case "cosmos.group.v1beta1.MsgVoteRequest.metadata":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.group.v1beta1.MsgVoteRequest.exec":
		return protoreflect.ValueOfEnum(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgVoteRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgVoteRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgVoteRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgVoteRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgVoteRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgVoteRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgVoteRequest)
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
		if x.ProposalId != 0 {
			n += 1 + runtime.Sov(uint64(x.ProposalId))
		}
		l = len(x.Voter)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Choice != 0 {
			n += 1 + runtime.Sov(uint64(x.Choice))
		}
		l = len(x.Metadata)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Exec != 0 {
			n += 1 + runtime.Sov(uint64(x.Exec))
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
		x := input.Message.Interface().(*MsgVoteRequest)
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
		if x.Exec != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Exec))
			i--
			dAtA[i] = 0x28
		}
		if len(x.Metadata) > 0 {
			i -= len(x.Metadata)
			copy(dAtA[i:], x.Metadata)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Metadata)))
			i--
			dAtA[i] = 0x22
		}
		if x.Choice != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Choice))
			i--
			dAtA[i] = 0x18
		}
		if len(x.Voter) > 0 {
			i -= len(x.Voter)
			copy(dAtA[i:], x.Voter)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Voter)))
			i--
			dAtA[i] = 0x12
		}
		if x.ProposalId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.ProposalId))
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
		x := input.Message.Interface().(*MsgVoteRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgVoteRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgVoteRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
				}
				x.ProposalId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.ProposalId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Voter", wireType)
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
				x.Voter = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 3:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Choice", wireType)
				}
				x.Choice = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Choice |= Choice(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 4:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
				x.Metadata = append(x.Metadata[:0], dAtA[iNdEx:postIndex]...)
				if x.Metadata == nil {
					x.Metadata = []byte{}
				}
				iNdEx = postIndex
			case 5:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Exec", wireType)
				}
				x.Exec = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Exec |= Exec(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
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
	md_MsgVoteResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgVoteResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgVoteResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgVoteResponse)(nil)

type fastReflection_MsgVoteResponse MsgVoteResponse

func (x *MsgVoteResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgVoteResponse)(x)
}

func (x *MsgVoteResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[19]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgVoteResponse_messageType fastReflection_MsgVoteResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgVoteResponse_messageType{}

type fastReflection_MsgVoteResponse_messageType struct{}

func (x fastReflection_MsgVoteResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgVoteResponse)(nil)
}
func (x fastReflection_MsgVoteResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgVoteResponse)
}
func (x fastReflection_MsgVoteResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgVoteResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgVoteResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgVoteResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgVoteResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgVoteResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgVoteResponse) New() protoreflect.Message {
	return new(fastReflection_MsgVoteResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgVoteResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgVoteResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgVoteResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgVoteResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgVoteResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgVoteResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgVoteResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgVoteResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgVoteResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgVoteResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgVoteResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgVoteResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgVoteResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgVoteResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgVoteResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgVoteResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgVoteResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgVoteResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgVoteResponse)
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
		x := input.Message.Interface().(*MsgVoteResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgVoteResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgVoteResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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
	md_MsgExecRequest             protoreflect.MessageDescriptor
	fd_MsgExecRequest_proposal_id protoreflect.FieldDescriptor
	fd_MsgExecRequest_signer      protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgExecRequest = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgExecRequest")
	fd_MsgExecRequest_proposal_id = md_MsgExecRequest.Fields().ByName("proposal_id")
	fd_MsgExecRequest_signer = md_MsgExecRequest.Fields().ByName("signer")
}

var _ protoreflect.Message = (*fastReflection_MsgExecRequest)(nil)

type fastReflection_MsgExecRequest MsgExecRequest

func (x *MsgExecRequest) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgExecRequest)(x)
}

func (x *MsgExecRequest) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[20]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgExecRequest_messageType fastReflection_MsgExecRequest_messageType
var _ protoreflect.MessageType = fastReflection_MsgExecRequest_messageType{}

type fastReflection_MsgExecRequest_messageType struct{}

func (x fastReflection_MsgExecRequest_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgExecRequest)(nil)
}
func (x fastReflection_MsgExecRequest_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgExecRequest)
}
func (x fastReflection_MsgExecRequest_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgExecRequest
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgExecRequest) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgExecRequest
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgExecRequest) Type() protoreflect.MessageType {
	return _fastReflection_MsgExecRequest_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgExecRequest) New() protoreflect.Message {
	return new(fastReflection_MsgExecRequest)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgExecRequest) Interface() protoreflect.ProtoMessage {
	return (*MsgExecRequest)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgExecRequest) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.ProposalId != uint64(0) {
		value := protoreflect.ValueOfUint64(x.ProposalId)
		if !f(fd_MsgExecRequest_proposal_id, value) {
			return
		}
	}
	if x.Signer != "" {
		value := protoreflect.ValueOfString(x.Signer)
		if !f(fd_MsgExecRequest_signer, value) {
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
func (x *fastReflection_MsgExecRequest) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgExecRequest.proposal_id":
		return x.ProposalId != uint64(0)
	case "cosmos.group.v1beta1.MsgExecRequest.signer":
		return x.Signer != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecRequest does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgExecRequest) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgExecRequest.proposal_id":
		x.ProposalId = uint64(0)
	case "cosmos.group.v1beta1.MsgExecRequest.signer":
		x.Signer = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecRequest does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgExecRequest) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.v1beta1.MsgExecRequest.proposal_id":
		value := x.ProposalId
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.v1beta1.MsgExecRequest.signer":
		value := x.Signer
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecRequest does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgExecRequest) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgExecRequest.proposal_id":
		x.ProposalId = value.Uint()
	case "cosmos.group.v1beta1.MsgExecRequest.signer":
		x.Signer = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecRequest does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgExecRequest) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgExecRequest.proposal_id":
		panic(fmt.Errorf("field proposal_id of message cosmos.group.v1beta1.MsgExecRequest is not mutable"))
	case "cosmos.group.v1beta1.MsgExecRequest.signer":
		panic(fmt.Errorf("field signer of message cosmos.group.v1beta1.MsgExecRequest is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecRequest does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgExecRequest) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.v1beta1.MsgExecRequest.proposal_id":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.v1beta1.MsgExecRequest.signer":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecRequest"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecRequest does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgExecRequest) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgExecRequest", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgExecRequest) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgExecRequest) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgExecRequest) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgExecRequest) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgExecRequest)
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
		if x.ProposalId != 0 {
			n += 1 + runtime.Sov(uint64(x.ProposalId))
		}
		l = len(x.Signer)
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
		x := input.Message.Interface().(*MsgExecRequest)
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
		if len(x.Signer) > 0 {
			i -= len(x.Signer)
			copy(dAtA[i:], x.Signer)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Signer)))
			i--
			dAtA[i] = 0x12
		}
		if x.ProposalId != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.ProposalId))
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
		x := input.Message.Interface().(*MsgExecRequest)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgExecRequest: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgExecRequest: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field ProposalId", wireType)
				}
				x.ProposalId = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.ProposalId |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Signer", wireType)
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
				x.Signer = string(dAtA[iNdEx:postIndex])
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

var (
	md_MsgExecResponse protoreflect.MessageDescriptor
)

func init() {
	file_cosmos_group_v1beta1_tx_proto_init()
	md_MsgExecResponse = File_cosmos_group_v1beta1_tx_proto.Messages().ByName("MsgExecResponse")
}

var _ protoreflect.Message = (*fastReflection_MsgExecResponse)(nil)

type fastReflection_MsgExecResponse MsgExecResponse

func (x *MsgExecResponse) ProtoReflect() protoreflect.Message {
	return (*fastReflection_MsgExecResponse)(x)
}

func (x *MsgExecResponse) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[21]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_MsgExecResponse_messageType fastReflection_MsgExecResponse_messageType
var _ protoreflect.MessageType = fastReflection_MsgExecResponse_messageType{}

type fastReflection_MsgExecResponse_messageType struct{}

func (x fastReflection_MsgExecResponse_messageType) Zero() protoreflect.Message {
	return (*fastReflection_MsgExecResponse)(nil)
}
func (x fastReflection_MsgExecResponse_messageType) New() protoreflect.Message {
	return new(fastReflection_MsgExecResponse)
}
func (x fastReflection_MsgExecResponse_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgExecResponse
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_MsgExecResponse) Descriptor() protoreflect.MessageDescriptor {
	return md_MsgExecResponse
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_MsgExecResponse) Type() protoreflect.MessageType {
	return _fastReflection_MsgExecResponse_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_MsgExecResponse) New() protoreflect.Message {
	return new(fastReflection_MsgExecResponse)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_MsgExecResponse) Interface() protoreflect.ProtoMessage {
	return (*MsgExecResponse)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_MsgExecResponse) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
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
func (x *fastReflection_MsgExecResponse) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecResponse does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgExecResponse) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecResponse does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_MsgExecResponse) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecResponse does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_MsgExecResponse) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecResponse does not contain field %s", fd.FullName()))
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
func (x *fastReflection_MsgExecResponse) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecResponse does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_MsgExecResponse) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.v1beta1.MsgExecResponse"))
		}
		panic(fmt.Errorf("message cosmos.group.v1beta1.MsgExecResponse does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_MsgExecResponse) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.v1beta1.MsgExecResponse", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_MsgExecResponse) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_MsgExecResponse) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_MsgExecResponse) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_MsgExecResponse) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*MsgExecResponse)
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
		if x.unknownFields != nil {
			n += len(x.unknownFields)
		}
		return protoiface.SizeOutput{
			NoUnkeyedLiterals: input.NoUnkeyedLiterals,
			Size:              n,
		}
	}

	marshal := func(input protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		x := input.Message.Interface().(*MsgExecResponse)
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
		x := input.Message.Interface().(*MsgExecResponse)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgExecResponse: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: MsgExecResponse: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
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

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MsgClient interface {
	// CreateGroup creates a new group with an admin account address, a list of members and some optional metadata.
	CreateGroup(ctx context.Context, in *MsgCreateGroupRequest, opts ...grpc.CallOption) (*MsgCreateGroupResponse, error)
	// UpdateGroupMembers updates the group members with given group id and admin address.
	UpdateGroupMembers(ctx context.Context, in *MsgUpdateGroupMembersRequest, opts ...grpc.CallOption) (*MsgUpdateGroupMembersResponse, error)
	// UpdateGroupAdmin updates the group admin with given group id and previous admin address.
	UpdateGroupAdmin(ctx context.Context, in *MsgUpdateGroupAdminRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAdminResponse, error)
	// UpdateGroupMetadata updates the group metadata with given group id and admin address.
	UpdateGroupMetadata(ctx context.Context, in *MsgUpdateGroupMetadataRequest, opts ...grpc.CallOption) (*MsgUpdateGroupMetadataResponse, error)
	// CreateGroupAccount creates a new group account using given DecisionPolicy.
	CreateGroupAccount(ctx context.Context, in *MsgCreateGroupAccountRequest, opts ...grpc.CallOption) (*MsgCreateGroupAccountResponse, error)
	// UpdateGroupAccountAdmin updates a group account admin.
	UpdateGroupAccountAdmin(ctx context.Context, in *MsgUpdateGroupAccountAdminRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAccountAdminResponse, error)
	// UpdateGroupAccountDecisionPolicy allows a group account decision policy to be updated.
	UpdateGroupAccountDecisionPolicy(ctx context.Context, in *MsgUpdateGroupAccountDecisionPolicyRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAccountDecisionPolicyResponse, error)
	// UpdateGroupAccountMetadata updates a group account metadata.
	UpdateGroupAccountMetadata(ctx context.Context, in *MsgUpdateGroupAccountMetadataRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAccountMetadataResponse, error)
	// CreateProposal submits a new proposal.
	CreateProposal(ctx context.Context, in *MsgCreateProposalRequest, opts ...grpc.CallOption) (*MsgCreateProposalResponse, error)
	// Vote allows a voter to vote on a proposal.
	Vote(ctx context.Context, in *MsgVoteRequest, opts ...grpc.CallOption) (*MsgVoteResponse, error)
	// Exec executes a proposal.
	Exec(ctx context.Context, in *MsgExecRequest, opts ...grpc.CallOption) (*MsgExecResponse, error)
}

type msgClient struct {
	cc grpc.ClientConnInterface
}

func NewMsgClient(cc grpc.ClientConnInterface) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) CreateGroup(ctx context.Context, in *MsgCreateGroupRequest, opts ...grpc.CallOption) (*MsgCreateGroupResponse, error) {
	out := new(MsgCreateGroupResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/CreateGroup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateGroupMembers(ctx context.Context, in *MsgUpdateGroupMembersRequest, opts ...grpc.CallOption) (*MsgUpdateGroupMembersResponse, error) {
	out := new(MsgUpdateGroupMembersResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/UpdateGroupMembers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateGroupAdmin(ctx context.Context, in *MsgUpdateGroupAdminRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAdminResponse, error) {
	out := new(MsgUpdateGroupAdminResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/UpdateGroupAdmin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateGroupMetadata(ctx context.Context, in *MsgUpdateGroupMetadataRequest, opts ...grpc.CallOption) (*MsgUpdateGroupMetadataResponse, error) {
	out := new(MsgUpdateGroupMetadataResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/UpdateGroupMetadata", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateGroupAccount(ctx context.Context, in *MsgCreateGroupAccountRequest, opts ...grpc.CallOption) (*MsgCreateGroupAccountResponse, error) {
	out := new(MsgCreateGroupAccountResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/CreateGroupAccount", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateGroupAccountAdmin(ctx context.Context, in *MsgUpdateGroupAccountAdminRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAccountAdminResponse, error) {
	out := new(MsgUpdateGroupAccountAdminResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/UpdateGroupAccountAdmin", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateGroupAccountDecisionPolicy(ctx context.Context, in *MsgUpdateGroupAccountDecisionPolicyRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAccountDecisionPolicyResponse, error) {
	out := new(MsgUpdateGroupAccountDecisionPolicyResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/UpdateGroupAccountDecisionPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateGroupAccountMetadata(ctx context.Context, in *MsgUpdateGroupAccountMetadataRequest, opts ...grpc.CallOption) (*MsgUpdateGroupAccountMetadataResponse, error) {
	out := new(MsgUpdateGroupAccountMetadataResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/UpdateGroupAccountMetadata", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) CreateProposal(ctx context.Context, in *MsgCreateProposalRequest, opts ...grpc.CallOption) (*MsgCreateProposalResponse, error) {
	out := new(MsgCreateProposalResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/CreateProposal", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Vote(ctx context.Context, in *MsgVoteRequest, opts ...grpc.CallOption) (*MsgVoteResponse, error) {
	out := new(MsgVoteResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/Vote", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) Exec(ctx context.Context, in *MsgExecRequest, opts ...grpc.CallOption) (*MsgExecResponse, error) {
	out := new(MsgExecResponse)
	err := c.cc.Invoke(ctx, "/cosmos.group.v1beta1.Msg/Exec", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
// All implementations must embed UnimplementedMsgServer
// for forward compatibility
type MsgServer interface {
	// CreateGroup creates a new group with an admin account address, a list of members and some optional metadata.
	CreateGroup(context.Context, *MsgCreateGroupRequest) (*MsgCreateGroupResponse, error)
	// UpdateGroupMembers updates the group members with given group id and admin address.
	UpdateGroupMembers(context.Context, *MsgUpdateGroupMembersRequest) (*MsgUpdateGroupMembersResponse, error)
	// UpdateGroupAdmin updates the group admin with given group id and previous admin address.
	UpdateGroupAdmin(context.Context, *MsgUpdateGroupAdminRequest) (*MsgUpdateGroupAdminResponse, error)
	// UpdateGroupMetadata updates the group metadata with given group id and admin address.
	UpdateGroupMetadata(context.Context, *MsgUpdateGroupMetadataRequest) (*MsgUpdateGroupMetadataResponse, error)
	// CreateGroupAccount creates a new group account using given DecisionPolicy.
	CreateGroupAccount(context.Context, *MsgCreateGroupAccountRequest) (*MsgCreateGroupAccountResponse, error)
	// UpdateGroupAccountAdmin updates a group account admin.
	UpdateGroupAccountAdmin(context.Context, *MsgUpdateGroupAccountAdminRequest) (*MsgUpdateGroupAccountAdminResponse, error)
	// UpdateGroupAccountDecisionPolicy allows a group account decision policy to be updated.
	UpdateGroupAccountDecisionPolicy(context.Context, *MsgUpdateGroupAccountDecisionPolicyRequest) (*MsgUpdateGroupAccountDecisionPolicyResponse, error)
	// UpdateGroupAccountMetadata updates a group account metadata.
	UpdateGroupAccountMetadata(context.Context, *MsgUpdateGroupAccountMetadataRequest) (*MsgUpdateGroupAccountMetadataResponse, error)
	// CreateProposal submits a new proposal.
	CreateProposal(context.Context, *MsgCreateProposalRequest) (*MsgCreateProposalResponse, error)
	// Vote allows a voter to vote on a proposal.
	Vote(context.Context, *MsgVoteRequest) (*MsgVoteResponse, error)
	// Exec executes a proposal.
	Exec(context.Context, *MsgExecRequest) (*MsgExecResponse, error)
	mustEmbedUnimplementedMsgServer()
}

// UnimplementedMsgServer must be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (UnimplementedMsgServer) CreateGroup(context.Context, *MsgCreateGroupRequest) (*MsgCreateGroupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGroup not implemented")
}
func (UnimplementedMsgServer) UpdateGroupMembers(context.Context, *MsgUpdateGroupMembersRequest) (*MsgUpdateGroupMembersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroupMembers not implemented")
}
func (UnimplementedMsgServer) UpdateGroupAdmin(context.Context, *MsgUpdateGroupAdminRequest) (*MsgUpdateGroupAdminResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroupAdmin not implemented")
}
func (UnimplementedMsgServer) UpdateGroupMetadata(context.Context, *MsgUpdateGroupMetadataRequest) (*MsgUpdateGroupMetadataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroupMetadata not implemented")
}
func (UnimplementedMsgServer) CreateGroupAccount(context.Context, *MsgCreateGroupAccountRequest) (*MsgCreateGroupAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateGroupAccount not implemented")
}
func (UnimplementedMsgServer) UpdateGroupAccountAdmin(context.Context, *MsgUpdateGroupAccountAdminRequest) (*MsgUpdateGroupAccountAdminResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroupAccountAdmin not implemented")
}
func (UnimplementedMsgServer) UpdateGroupAccountDecisionPolicy(context.Context, *MsgUpdateGroupAccountDecisionPolicyRequest) (*MsgUpdateGroupAccountDecisionPolicyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroupAccountDecisionPolicy not implemented")
}
func (UnimplementedMsgServer) UpdateGroupAccountMetadata(context.Context, *MsgUpdateGroupAccountMetadataRequest) (*MsgUpdateGroupAccountMetadataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateGroupAccountMetadata not implemented")
}
func (UnimplementedMsgServer) CreateProposal(context.Context, *MsgCreateProposalRequest) (*MsgCreateProposalResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateProposal not implemented")
}
func (UnimplementedMsgServer) Vote(context.Context, *MsgVoteRequest) (*MsgVoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Vote not implemented")
}
func (UnimplementedMsgServer) Exec(context.Context, *MsgExecRequest) (*MsgExecResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Exec not implemented")
}
func (UnimplementedMsgServer) mustEmbedUnimplementedMsgServer() {}

// UnsafeMsgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MsgServer will
// result in compilation errors.
type UnsafeMsgServer interface {
	mustEmbedUnimplementedMsgServer()
}

func RegisterMsgServer(s grpc.ServiceRegistrar, srv MsgServer) {
	s.RegisterService(&Msg_ServiceDesc, srv)
}

func _Msg_CreateGroup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateGroupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateGroup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/CreateGroup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateGroup(ctx, req.(*MsgCreateGroupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateGroupMembers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateGroupMembersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateGroupMembers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/UpdateGroupMembers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateGroupMembers(ctx, req.(*MsgUpdateGroupMembersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateGroupAdmin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateGroupAdminRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateGroupAdmin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/UpdateGroupAdmin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateGroupAdmin(ctx, req.(*MsgUpdateGroupAdminRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateGroupMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateGroupMetadataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateGroupMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/UpdateGroupMetadata",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateGroupMetadata(ctx, req.(*MsgUpdateGroupMetadataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateGroupAccount_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateGroupAccountRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateGroupAccount(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/CreateGroupAccount",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateGroupAccount(ctx, req.(*MsgCreateGroupAccountRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateGroupAccountAdmin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateGroupAccountAdminRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateGroupAccountAdmin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/UpdateGroupAccountAdmin",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateGroupAccountAdmin(ctx, req.(*MsgUpdateGroupAccountAdminRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateGroupAccountDecisionPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateGroupAccountDecisionPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateGroupAccountDecisionPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/UpdateGroupAccountDecisionPolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateGroupAccountDecisionPolicy(ctx, req.(*MsgUpdateGroupAccountDecisionPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateGroupAccountMetadata_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateGroupAccountMetadataRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateGroupAccountMetadata(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/UpdateGroupAccountMetadata",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateGroupAccountMetadata(ctx, req.(*MsgUpdateGroupAccountMetadataRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CreateProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateProposalRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/CreateProposal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateProposal(ctx, req.(*MsgCreateProposalRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Vote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgVoteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Vote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/Vote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Vote(ctx, req.(*MsgVoteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Exec_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgExecRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Exec(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.group.v1beta1.Msg/Exec",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Exec(ctx, req.(*MsgExecRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Msg_ServiceDesc is the grpc.ServiceDesc for Msg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Msg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cosmos.group.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateGroup",
			Handler:    _Msg_CreateGroup_Handler,
		},
		{
			MethodName: "UpdateGroupMembers",
			Handler:    _Msg_UpdateGroupMembers_Handler,
		},
		{
			MethodName: "UpdateGroupAdmin",
			Handler:    _Msg_UpdateGroupAdmin_Handler,
		},
		{
			MethodName: "UpdateGroupMetadata",
			Handler:    _Msg_UpdateGroupMetadata_Handler,
		},
		{
			MethodName: "CreateGroupAccount",
			Handler:    _Msg_CreateGroupAccount_Handler,
		},
		{
			MethodName: "UpdateGroupAccountAdmin",
			Handler:    _Msg_UpdateGroupAccountAdmin_Handler,
		},
		{
			MethodName: "UpdateGroupAccountDecisionPolicy",
			Handler:    _Msg_UpdateGroupAccountDecisionPolicy_Handler,
		},
		{
			MethodName: "UpdateGroupAccountMetadata",
			Handler:    _Msg_UpdateGroupAccountMetadata_Handler,
		},
		{
			MethodName: "CreateProposal",
			Handler:    _Msg_CreateProposal_Handler,
		},
		{
			MethodName: "Vote",
			Handler:    _Msg_Vote_Handler,
		},
		{
			MethodName: "Exec",
			Handler:    _Msg_Exec_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cosmos/group/v1beta1/tx.proto",
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.0
// 	protoc        v3.19.1
// source: cosmos/group/v1beta1/tx.proto

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Exec defines modes of execution of a proposal on creation or on new vote.
type Exec int32

const (
	// An empty value means that there should be a separate
	// MsgExec request for the proposal to execute.
	Exec_EXEC_UNSPECIFIED Exec = 0
	// Try to execute the proposal immediately.
	// If the proposal is not allowed per the DecisionPolicy,
	// the proposal will still be open and could
	// be executed at a later point.
	Exec_EXEC_TRY Exec = 1
)

// Enum value maps for Exec.
var (
	Exec_name = map[int32]string{
		0: "EXEC_UNSPECIFIED",
		1: "EXEC_TRY",
	}
	Exec_value = map[string]int32{
		"EXEC_UNSPECIFIED": 0,
		"EXEC_TRY":         1,
	}
)

func (x Exec) Enum() *Exec {
	p := new(Exec)
	*p = x
	return p
}

func (x Exec) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Exec) Descriptor() protoreflect.EnumDescriptor {
	return file_cosmos_group_v1beta1_tx_proto_enumTypes[0].Descriptor()
}

func (Exec) Type() protoreflect.EnumType {
	return &file_cosmos_group_v1beta1_tx_proto_enumTypes[0]
}

func (x Exec) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Exec.Descriptor instead.
func (Exec) EnumDescriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{0}
}

// MsgCreateGroupRequest is the Msg/CreateGroup request type.
type MsgCreateGroupRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// members defines the group members.
	Members []*Member `protobuf:"bytes,2,rep,name=members,proto3" json:"members,omitempty"`
	// metadata is any arbitrary metadata to attached to the group.
	Metadata []byte `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (x *MsgCreateGroupRequest) Reset() {
	*x = MsgCreateGroupRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgCreateGroupRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgCreateGroupRequest) ProtoMessage() {}

// Deprecated: Use MsgCreateGroupRequest.ProtoReflect.Descriptor instead.
func (*MsgCreateGroupRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{0}
}

func (x *MsgCreateGroupRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgCreateGroupRequest) GetMembers() []*Member {
	if x != nil {
		return x.Members
	}
	return nil
}

func (x *MsgCreateGroupRequest) GetMetadata() []byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

// MsgCreateGroupResponse is the Msg/CreateGroup response type.
type MsgCreateGroupResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// group_id is the unique ID of the newly created group.
	GroupId uint64 `protobuf:"varint,1,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
}

func (x *MsgCreateGroupResponse) Reset() {
	*x = MsgCreateGroupResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgCreateGroupResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgCreateGroupResponse) ProtoMessage() {}

// Deprecated: Use MsgCreateGroupResponse.ProtoReflect.Descriptor instead.
func (*MsgCreateGroupResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{1}
}

func (x *MsgCreateGroupResponse) GetGroupId() uint64 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

// MsgUpdateGroupMembersRequest is the Msg/UpdateGroupMembers request type.
type MsgUpdateGroupMembersRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// group_id is the unique ID of the group.
	GroupId uint64 `protobuf:"varint,2,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	// member_updates is the list of members to update,
	// set weight to 0 to remove a member.
	MemberUpdates []*Member `protobuf:"bytes,3,rep,name=member_updates,json=memberUpdates,proto3" json:"member_updates,omitempty"`
}

func (x *MsgUpdateGroupMembersRequest) Reset() {
	*x = MsgUpdateGroupMembersRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupMembersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupMembersRequest) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupMembersRequest.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupMembersRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{2}
}

func (x *MsgUpdateGroupMembersRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgUpdateGroupMembersRequest) GetGroupId() uint64 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

func (x *MsgUpdateGroupMembersRequest) GetMemberUpdates() []*Member {
	if x != nil {
		return x.MemberUpdates
	}
	return nil
}

// MsgUpdateGroupMembersResponse is the Msg/UpdateGroupMembers response type.
type MsgUpdateGroupMembersResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgUpdateGroupMembersResponse) Reset() {
	*x = MsgUpdateGroupMembersResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupMembersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupMembersResponse) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupMembersResponse.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupMembersResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{3}
}

// MsgUpdateGroupAdminRequest is the Msg/UpdateGroupAdmin request type.
type MsgUpdateGroupAdminRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the current account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// group_id is the unique ID of the group.
	GroupId uint64 `protobuf:"varint,2,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	// new_admin is the group new admin account address.
	NewAdmin string `protobuf:"bytes,3,opt,name=new_admin,json=newAdmin,proto3" json:"new_admin,omitempty"`
}

func (x *MsgUpdateGroupAdminRequest) Reset() {
	*x = MsgUpdateGroupAdminRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAdminRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAdminRequest) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAdminRequest.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAdminRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{4}
}

func (x *MsgUpdateGroupAdminRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgUpdateGroupAdminRequest) GetGroupId() uint64 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

func (x *MsgUpdateGroupAdminRequest) GetNewAdmin() string {
	if x != nil {
		return x.NewAdmin
	}
	return ""
}

// MsgUpdateGroupAdminResponse is the Msg/UpdateGroupAdmin response type.
type MsgUpdateGroupAdminResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgUpdateGroupAdminResponse) Reset() {
	*x = MsgUpdateGroupAdminResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAdminResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAdminResponse) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAdminResponse.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAdminResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{5}
}

// MsgUpdateGroupMetadataRequest is the Msg/UpdateGroupMetadata request type.
type MsgUpdateGroupMetadataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// group_id is the unique ID of the group.
	GroupId uint64 `protobuf:"varint,2,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	// metadata is the updated group's metadata.
	Metadata []byte `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (x *MsgUpdateGroupMetadataRequest) Reset() {
	*x = MsgUpdateGroupMetadataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupMetadataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupMetadataRequest) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupMetadataRequest.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupMetadataRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{6}
}

func (x *MsgUpdateGroupMetadataRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgUpdateGroupMetadataRequest) GetGroupId() uint64 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

func (x *MsgUpdateGroupMetadataRequest) GetMetadata() []byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

// MsgUpdateGroupMetadataResponse is the Msg/UpdateGroupMetadata response type.
type MsgUpdateGroupMetadataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgUpdateGroupMetadataResponse) Reset() {
	*x = MsgUpdateGroupMetadataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupMetadataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupMetadataResponse) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupMetadataResponse.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupMetadataResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{7}
}

// MsgCreateGroupAccountRequest is the Msg/CreateGroupAccount request type.
type MsgCreateGroupAccountRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// group_id is the unique ID of the group.
	GroupId uint64 `protobuf:"varint,2,opt,name=group_id,json=groupId,proto3" json:"group_id,omitempty"`
	// metadata is any arbitrary metadata to attached to the group account.
	Metadata []byte `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// decision_policy specifies the group account's decision policy.
	DecisionPolicy *anypb.Any `protobuf:"bytes,4,opt,name=decision_policy,json=decisionPolicy,proto3" json:"decision_policy,omitempty"`
}

func (x *MsgCreateGroupAccountRequest) Reset() {
	*x = MsgCreateGroupAccountRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgCreateGroupAccountRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgCreateGroupAccountRequest) ProtoMessage() {}

// Deprecated: Use MsgCreateGroupAccountRequest.ProtoReflect.Descriptor instead.
func (*MsgCreateGroupAccountRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{8}
}

func (x *MsgCreateGroupAccountRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgCreateGroupAccountRequest) GetGroupId() uint64 {
	if x != nil {
		return x.GroupId
	}
	return 0
}

func (x *MsgCreateGroupAccountRequest) GetMetadata() []byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *MsgCreateGroupAccountRequest) GetDecisionPolicy() *anypb.Any {
	if x != nil {
		return x.DecisionPolicy
	}
	return nil
}

// MsgCreateGroupAccountResponse is the Msg/CreateGroupAccount response type.
type MsgCreateGroupAccountResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// address is the account address of the newly created group account.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (x *MsgCreateGroupAccountResponse) Reset() {
	*x = MsgCreateGroupAccountResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgCreateGroupAccountResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgCreateGroupAccountResponse) ProtoMessage() {}

// Deprecated: Use MsgCreateGroupAccountResponse.ProtoReflect.Descriptor instead.
func (*MsgCreateGroupAccountResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{9}
}

func (x *MsgCreateGroupAccountResponse) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

// MsgUpdateGroupAccountAdminRequest is the Msg/UpdateGroupAccountAdmin request type.
type MsgUpdateGroupAccountAdminRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// address is the group account address.
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	// new_admin is the new group account admin.
	NewAdmin string `protobuf:"bytes,3,opt,name=new_admin,json=newAdmin,proto3" json:"new_admin,omitempty"`
}

func (x *MsgUpdateGroupAccountAdminRequest) Reset() {
	*x = MsgUpdateGroupAccountAdminRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAccountAdminRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAccountAdminRequest) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAccountAdminRequest.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAccountAdminRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{10}
}

func (x *MsgUpdateGroupAccountAdminRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgUpdateGroupAccountAdminRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *MsgUpdateGroupAccountAdminRequest) GetNewAdmin() string {
	if x != nil {
		return x.NewAdmin
	}
	return ""
}

// MsgUpdateGroupAccountAdminResponse is the Msg/UpdateGroupAccountAdmin response type.
type MsgUpdateGroupAccountAdminResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgUpdateGroupAccountAdminResponse) Reset() {
	*x = MsgUpdateGroupAccountAdminResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAccountAdminResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAccountAdminResponse) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAccountAdminResponse.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAccountAdminResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{11}
}

// MsgUpdateGroupAccountDecisionPolicyRequest is the Msg/UpdateGroupAccountDecisionPolicy request type.
type MsgUpdateGroupAccountDecisionPolicyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// address is the group account address.
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	// decision_policy is the updated group account decision policy.
	DecisionPolicy *anypb.Any `protobuf:"bytes,3,opt,name=decision_policy,json=decisionPolicy,proto3" json:"decision_policy,omitempty"`
}

func (x *MsgUpdateGroupAccountDecisionPolicyRequest) Reset() {
	*x = MsgUpdateGroupAccountDecisionPolicyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAccountDecisionPolicyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAccountDecisionPolicyRequest) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAccountDecisionPolicyRequest.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAccountDecisionPolicyRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{12}
}

func (x *MsgUpdateGroupAccountDecisionPolicyRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgUpdateGroupAccountDecisionPolicyRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *MsgUpdateGroupAccountDecisionPolicyRequest) GetDecisionPolicy() *anypb.Any {
	if x != nil {
		return x.DecisionPolicy
	}
	return nil
}

// MsgUpdateGroupAccountDecisionPolicyResponse is the Msg/UpdateGroupAccountDecisionPolicy response type.
type MsgUpdateGroupAccountDecisionPolicyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgUpdateGroupAccountDecisionPolicyResponse) Reset() {
	*x = MsgUpdateGroupAccountDecisionPolicyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAccountDecisionPolicyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAccountDecisionPolicyResponse) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAccountDecisionPolicyResponse.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAccountDecisionPolicyResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{13}
}

// MsgUpdateGroupAccountMetadataRequest is the Msg/UpdateGroupAccountMetadata request type.
type MsgUpdateGroupAccountMetadataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// admin is the account address of the group admin.
	Admin string `protobuf:"bytes,1,opt,name=admin,proto3" json:"admin,omitempty"`
	// address is the group account address.
	Address string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	// metadata is the updated group account metadata.
	Metadata []byte `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (x *MsgUpdateGroupAccountMetadataRequest) Reset() {
	*x = MsgUpdateGroupAccountMetadataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAccountMetadataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAccountMetadataRequest) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAccountMetadataRequest.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAccountMetadataRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{14}
}

func (x *MsgUpdateGroupAccountMetadataRequest) GetAdmin() string {
	if x != nil {
		return x.Admin
	}
	return ""
}

func (x *MsgUpdateGroupAccountMetadataRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *MsgUpdateGroupAccountMetadataRequest) GetMetadata() []byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

// MsgUpdateGroupAccountMetadataResponse is the Msg/UpdateGroupAccountMetadata response type.
type MsgUpdateGroupAccountMetadataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgUpdateGroupAccountMetadataResponse) Reset() {
	*x = MsgUpdateGroupAccountMetadataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[15]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgUpdateGroupAccountMetadataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgUpdateGroupAccountMetadataResponse) ProtoMessage() {}

// Deprecated: Use MsgUpdateGroupAccountMetadataResponse.ProtoReflect.Descriptor instead.
func (*MsgUpdateGroupAccountMetadataResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{15}
}

// MsgCreateProposalRequest is the Msg/CreateProposal request type.
type MsgCreateProposalRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// address is the group account address.
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	// proposers are the account addresses of the proposers.
	// Proposers signatures will be counted as yes votes.
	Proposers []string `protobuf:"bytes,2,rep,name=proposers,proto3" json:"proposers,omitempty"`
	// metadata is any arbitrary metadata to attached to the proposal.
	Metadata []byte `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// msgs is a list of Msgs that will be executed if the proposal passes.
	Msgs []*anypb.Any `protobuf:"bytes,4,rep,name=msgs,proto3" json:"msgs,omitempty"`
	// exec defines the mode of execution of the proposal,
	// whether it should be executed immediately on creation or not.
	// If so, proposers signatures are considered as Yes votes.
	Exec Exec `protobuf:"varint,5,opt,name=exec,proto3,enum=cosmos.group.v1beta1.Exec" json:"exec,omitempty"`
}

func (x *MsgCreateProposalRequest) Reset() {
	*x = MsgCreateProposalRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[16]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgCreateProposalRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgCreateProposalRequest) ProtoMessage() {}

// Deprecated: Use MsgCreateProposalRequest.ProtoReflect.Descriptor instead.
func (*MsgCreateProposalRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{16}
}

func (x *MsgCreateProposalRequest) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *MsgCreateProposalRequest) GetProposers() []string {
	if x != nil {
		return x.Proposers
	}
	return nil
}

func (x *MsgCreateProposalRequest) GetMetadata() []byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *MsgCreateProposalRequest) GetMsgs() []*anypb.Any {
	if x != nil {
		return x.Msgs
	}
	return nil
}

func (x *MsgCreateProposalRequest) GetExec() Exec {
	if x != nil {
		return x.Exec
	}
	return Exec_EXEC_UNSPECIFIED
}

// MsgCreateProposalResponse is the Msg/CreateProposal response type.
type MsgCreateProposalResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// proposal is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
}

func (x *MsgCreateProposalResponse) Reset() {
	*x = MsgCreateProposalResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[17]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgCreateProposalResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgCreateProposalResponse) ProtoMessage() {}

// Deprecated: Use MsgCreateProposalResponse.ProtoReflect.Descriptor instead.
func (*MsgCreateProposalResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{17}
}

func (x *MsgCreateProposalResponse) GetProposalId() uint64 {
	if x != nil {
		return x.ProposalId
	}
	return 0
}

// MsgVoteRequest is the Msg/Vote request type.
type MsgVoteRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// proposal is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	// voter is the voter account address.
	Voter string `protobuf:"bytes,2,opt,name=voter,proto3" json:"voter,omitempty"`
	// choice is the voter's choice on the proposal.
	Choice Choice `protobuf:"varint,3,opt,name=choice,proto3,enum=cosmos.group.v1beta1.Choice" json:"choice,omitempty"`
	// metadata is any arbitrary metadata to attached to the vote.
	Metadata []byte `protobuf:"bytes,4,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// exec defines whether the proposal should be executed
	// immediately after voting or not.
	Exec Exec `protobuf:"varint,5,opt,name=exec,proto3,enum=cosmos.group.v1beta1.Exec" json:"exec,omitempty"`
}

func (x *MsgVoteRequest) Reset() {
	*x = MsgVoteRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[18]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgVoteRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgVoteRequest) ProtoMessage() {}

// Deprecated: Use MsgVoteRequest.ProtoReflect.Descriptor instead.
func (*MsgVoteRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{18}
}

func (x *MsgVoteRequest) GetProposalId() uint64 {
	if x != nil {
		return x.ProposalId
	}
	return 0
}

func (x *MsgVoteRequest) GetVoter() string {
	if x != nil {
		return x.Voter
	}
	return ""
}

func (x *MsgVoteRequest) GetChoice() Choice {
	if x != nil {
		return x.Choice
	}
	return Choice_CHOICE_UNSPECIFIED
}

func (x *MsgVoteRequest) GetMetadata() []byte {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *MsgVoteRequest) GetExec() Exec {
	if x != nil {
		return x.Exec
	}
	return Exec_EXEC_UNSPECIFIED
}

// MsgVoteResponse is the Msg/Vote response type.
type MsgVoteResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgVoteResponse) Reset() {
	*x = MsgVoteResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[19]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgVoteResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgVoteResponse) ProtoMessage() {}

// Deprecated: Use MsgVoteResponse.ProtoReflect.Descriptor instead.
func (*MsgVoteResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{19}
}

// MsgExecRequest is the Msg/Exec request type.
type MsgExecRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// proposal is the unique ID of the proposal.
	ProposalId uint64 `protobuf:"varint,1,opt,name=proposal_id,json=proposalId,proto3" json:"proposal_id,omitempty"`
	// signer is the account address used to execute the proposal.
	Signer string `protobuf:"bytes,2,opt,name=signer,proto3" json:"signer,omitempty"`
}

func (x *MsgExecRequest) Reset() {
	*x = MsgExecRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[20]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgExecRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgExecRequest) ProtoMessage() {}

// Deprecated: Use MsgExecRequest.ProtoReflect.Descriptor instead.
func (*MsgExecRequest) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{20}
}

func (x *MsgExecRequest) GetProposalId() uint64 {
	if x != nil {
		return x.ProposalId
	}
	return 0
}

func (x *MsgExecRequest) GetSigner() string {
	if x != nil {
		return x.Signer
	}
	return ""
}

// MsgExecResponse is the Msg/Exec request type.
type MsgExecResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgExecResponse) Reset() {
	*x = MsgExecResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_v1beta1_tx_proto_msgTypes[21]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgExecResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgExecResponse) ProtoMessage() {}

// Deprecated: Use MsgExecResponse.ProtoReflect.Descriptor instead.
func (*MsgExecResponse) Descriptor() ([]byte, []int) {
	return file_cosmos_group_v1beta1_tx_proto_rawDescGZIP(), []int{21}
}

var File_cosmos_group_v1beta1_tx_proto protoreflect.FileDescriptor

var file_cosmos_group_v1beta1_tx_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2f, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x74, 0x78, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31,
	0x62, 0x65, 0x74, 0x61, 0x31, 0x1a, 0x14, 0x67, 0x6f, 0x67, 0x6f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x67, 0x6f, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x63, 0x6f, 0x73,
	0x6d, 0x6f, 0x73, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x20, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2f,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xa1, 0x01, 0x0a, 0x15, 0x4d, 0x73, 0x67, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a,
	0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4,
	0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x3c, 0x0a,
	0x07, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1c,
	0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31,
	0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x42, 0x04, 0xc8, 0xde,
	0x1f, 0x00, 0x52, 0x07, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x33, 0x0a, 0x16, 0x4d, 0x73, 0x67, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x04, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x22, 0xb4, 0x01, 0x0a,
	0x1c, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d,
	0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a,
	0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4,
	0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x19, 0x0a,
	0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x49, 0x0a, 0x0e, 0x6d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x5f, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x42, 0x04,
	0xc8, 0xde, 0x1f, 0x00, 0x52, 0x0d, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x73, 0x22, 0x1f, 0x0a, 0x1d, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x9e, 0x01, 0x0a, 0x1a, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x35,
	0x0a, 0x09, 0x6e, 0x65, 0x77, 0x5f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x08, 0x6e, 0x65, 0x77,
	0x41, 0x64, 0x6d, 0x69, 0x6e, 0x22, 0x1d, 0x0a, 0x1b, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x86, 0x01, 0x0a, 0x1d, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f,
	0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52,
	0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f,
	0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49,
	0x64, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x20, 0x0a,
	0x1e, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0xde, 0x01, 0x0a, 0x1c, 0x4d, 0x73, 0x67, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x2e, 0x0a, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x12, 0x19, 0x0a, 0x08, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x07, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x51, 0x0a, 0x0f, 0x64, 0x65, 0x63, 0x69, 0x73,
	0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x42, 0x12, 0xca, 0xb4, 0x2d, 0x0e, 0x44, 0x65, 0x63, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x0e, 0x64, 0x65, 0x63, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x3a, 0x04, 0x88, 0xa0, 0x1f, 0x00,
	0x22, 0x53, 0x0a, 0x1d, 0x4d, 0x73, 0x67, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x32, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x07, 0x61, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0xbe, 0x01, 0x0a, 0x21, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41,
	0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x05, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x32, 0x0a, 0x07, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4,
	0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x35, 0x0a, 0x09, 0x6e, 0x65, 0x77, 0x5f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x08, 0x6e, 0x65,
	0x77, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x22, 0x24, 0x0a, 0x22, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41,
	0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0xe9, 0x01, 0x0a,
	0x2a, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f,
	0x6c, 0x69, 0x63, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x2e, 0x0a, 0x05, 0x61,
	0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x32, 0x0a, 0x07, 0x61,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4,
	0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12,
	0x51, 0x0a, 0x0f, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x42, 0x12,
	0xca, 0xb4, 0x2d, 0x0e, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x52, 0x0e, 0x64, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69,
	0x63, 0x79, 0x3a, 0x04, 0x88, 0xa0, 0x1f, 0x00, 0x22, 0x2d, 0x0a, 0x2b, 0x4d, 0x73, 0x67, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0xa6, 0x01, 0x0a, 0x24, 0x4d, 0x73, 0x67, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x2e, 0x0a, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42,
	0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x05, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x12, 0x32, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x07, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x22, 0x27, 0x0a, 0x25, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0xe8, 0x01, 0x0a, 0x18, 0x4d, 0x73,
	0x67, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x32, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73,
	0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e,
	0x67, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x72,
	0x6f, 0x70, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x70,
	0x72, 0x6f, 0x70, 0x6f, 0x73, 0x65, 0x72, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x12, 0x28, 0x0a, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x18, 0x04, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x04, 0x6d, 0x73, 0x67, 0x73, 0x12, 0x2e,
	0x0a, 0x04, 0x65, 0x78, 0x65, 0x63, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x63,
	0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2e, 0x45, 0x78, 0x65, 0x63, 0x52, 0x04, 0x65, 0x78, 0x65, 0x63, 0x3a, 0x04,
	0x88, 0xa0, 0x1f, 0x00, 0x22, 0x3c, 0x0a, 0x19, 0x4d, 0x73, 0x67, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c,
	0x49, 0x64, 0x22, 0xe3, 0x01, 0x0a, 0x0e, 0x4d, 0x73, 0x67, 0x56, 0x6f, 0x74, 0x65, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a, 0x0b, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61,
	0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70,
	0x6f, 0x73, 0x61, 0x6c, 0x49, 0x64, 0x12, 0x2e, 0x0a, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18, 0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f,
	0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52,
	0x05, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x12, 0x34, 0x0a, 0x06, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x43, 0x68,
	0x6f, 0x69, 0x63, 0x65, 0x52, 0x06, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x12, 0x1a, 0x0a, 0x08,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x08,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x2e, 0x0a, 0x04, 0x65, 0x78, 0x65, 0x63,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x45, 0x78,
	0x65, 0x63, 0x52, 0x04, 0x65, 0x78, 0x65, 0x63, 0x22, 0x11, 0x0a, 0x0f, 0x4d, 0x73, 0x67, 0x56,
	0x6f, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x63, 0x0a, 0x0e, 0x4d,
	0x73, 0x67, 0x45, 0x78, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1f, 0x0a,
	0x0b, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x0a, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x49, 0x64, 0x12, 0x30,
	0x0a, 0x06, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x18,
	0xd2, 0xb4, 0x2d, 0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x06, 0x73, 0x69, 0x67, 0x6e, 0x65, 0x72,
	0x22, 0x11, 0x0a, 0x0f, 0x4d, 0x73, 0x67, 0x45, 0x78, 0x65, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x2a, 0x2a, 0x0a, 0x04, 0x45, 0x78, 0x65, 0x63, 0x12, 0x14, 0x0a, 0x10, 0x45,
	0x58, 0x45, 0x43, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10,
	0x00, 0x12, 0x0c, 0x0a, 0x08, 0x45, 0x58, 0x45, 0x43, 0x5f, 0x54, 0x52, 0x59, 0x10, 0x01, 0x32,
	0xd7, 0x0a, 0x0a, 0x03, 0x4d, 0x73, 0x67, 0x12, 0x68, 0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x12, 0x2b, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73,
	0x67, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x2c, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x7d, 0x0a, 0x12, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70,
	0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x12, 0x32, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d,
	0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x6d,
	0x62, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x33, 0x2e, 0x63, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x77, 0x0a, 0x10, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41,
	0x64, 0x6d, 0x69, 0x6e, 0x12, 0x30, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x31, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73,
	0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x80, 0x01, 0x0a, 0x13, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x12, 0x33, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x34, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73,
	0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x7d, 0x0a, 0x12,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x12, 0x32, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75,
	0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x33, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73,
	0x67, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x8c, 0x01, 0x0a, 0x17,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x12, 0x37, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d,
	0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63,
	0x6f, 0x75, 0x6e, 0x74, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x38, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e,
	0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x41, 0x64, 0x6d,
	0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0xa7, 0x01, 0x0a, 0x20, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x44, 0x65, 0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x12,
	0x40, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x44, 0x65, 0x63, 0x69,
	0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x41, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70,
	0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x44, 0x65,
	0x63, 0x69, 0x73, 0x69, 0x6f, 0x6e, 0x50, 0x6f, 0x6c, 0x69, 0x63, 0x79, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x95, 0x01, 0x0a, 0x1a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x47,
	0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x3a, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x3b, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x47, 0x72, 0x6f, 0x75, 0x70, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x71, 0x0a, 0x0e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x12, 0x2e,
	0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31,
	0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50,
	0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2f,
	0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31,
	0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x50,
	0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x53, 0x0a, 0x04, 0x56, 0x6f, 0x74, 0x65, 0x12, 0x24, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d,
	0x73, 0x67, 0x56, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x25, 0x2e,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x56, 0x6f, 0x74, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x53, 0x0a, 0x04, 0x45, 0x78, 0x65, 0x63, 0x12, 0x24, 0x2e, 0x63,
	0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x45, 0x78, 0x65, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x25, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75,
	0x70, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d, 0x73, 0x67, 0x45, 0x78, 0x65,
	0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xd9, 0x01, 0x0a, 0x18, 0x63, 0x6f,
	0x6d, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x76,
	0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x42, 0x07, 0x54, 0x78, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50,
	0x01, 0x5a, 0x42, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2d, 0x73, 0x64, 0x6b, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70,
	0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x3b, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x76, 0x31,
	0x62, 0x65, 0x74, 0x61, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x47, 0x58, 0xaa, 0x02, 0x14, 0x43, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x56, 0x31, 0x62, 0x65, 0x74,
	0x61, 0x31, 0xca, 0x02, 0x14, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x47, 0x72, 0x6f, 0x75,
	0x70, 0x5c, 0x56, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0xe2, 0x02, 0x20, 0x43, 0x6f, 0x73, 0x6d,
	0x6f, 0x73, 0x5c, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x5c, 0x56, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x16, 0x43,
	0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x3a, 0x3a, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x3a, 0x3a, 0x56, 0x31,
	0x62, 0x65, 0x74, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cosmos_group_v1beta1_tx_proto_rawDescOnce sync.Once
	file_cosmos_group_v1beta1_tx_proto_rawDescData = file_cosmos_group_v1beta1_tx_proto_rawDesc
)

func file_cosmos_group_v1beta1_tx_proto_rawDescGZIP() []byte {
	file_cosmos_group_v1beta1_tx_proto_rawDescOnce.Do(func() {
		file_cosmos_group_v1beta1_tx_proto_rawDescData = protoimpl.X.CompressGZIP(file_cosmos_group_v1beta1_tx_proto_rawDescData)
	})
	return file_cosmos_group_v1beta1_tx_proto_rawDescData
}

var file_cosmos_group_v1beta1_tx_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_cosmos_group_v1beta1_tx_proto_msgTypes = make([]protoimpl.MessageInfo, 22)
var file_cosmos_group_v1beta1_tx_proto_goTypes = []interface{}{
	(Exec)(0),                                           // 0: cosmos.group.v1beta1.Exec
	(*MsgCreateGroupRequest)(nil),                       // 1: cosmos.group.v1beta1.MsgCreateGroupRequest
	(*MsgCreateGroupResponse)(nil),                      // 2: cosmos.group.v1beta1.MsgCreateGroupResponse
	(*MsgUpdateGroupMembersRequest)(nil),                // 3: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest
	(*MsgUpdateGroupMembersResponse)(nil),               // 4: cosmos.group.v1beta1.MsgUpdateGroupMembersResponse
	(*MsgUpdateGroupAdminRequest)(nil),                  // 5: cosmos.group.v1beta1.MsgUpdateGroupAdminRequest
	(*MsgUpdateGroupAdminResponse)(nil),                 // 6: cosmos.group.v1beta1.MsgUpdateGroupAdminResponse
	(*MsgUpdateGroupMetadataRequest)(nil),               // 7: cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest
	(*MsgUpdateGroupMetadataResponse)(nil),              // 8: cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse
	(*MsgCreateGroupAccountRequest)(nil),                // 9: cosmos.group.v1beta1.MsgCreateGroupAccountRequest
	(*MsgCreateGroupAccountResponse)(nil),               // 10: cosmos.group.v1beta1.MsgCreateGroupAccountResponse
	(*MsgUpdateGroupAccountAdminRequest)(nil),           // 11: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest
	(*MsgUpdateGroupAccountAdminResponse)(nil),          // 12: cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse
	(*MsgUpdateGroupAccountDecisionPolicyRequest)(nil),  // 13: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest
	(*MsgUpdateGroupAccountDecisionPolicyResponse)(nil), // 14: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse
	(*MsgUpdateGroupAccountMetadataRequest)(nil),        // 15: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest
	(*MsgUpdateGroupAccountMetadataResponse)(nil),       // 16: cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse
	(*MsgCreateProposalRequest)(nil),                    // 17: cosmos.group.v1beta1.MsgCreateProposalRequest
	(*MsgCreateProposalResponse)(nil),                   // 18: cosmos.group.v1beta1.MsgCreateProposalResponse
	(*MsgVoteRequest)(nil),                              // 19: cosmos.group.v1beta1.MsgVoteRequest
	(*MsgVoteResponse)(nil),                             // 20: cosmos.group.v1beta1.MsgVoteResponse
	(*MsgExecRequest)(nil),                              // 21: cosmos.group.v1beta1.MsgExecRequest
	(*MsgExecResponse)(nil),                             // 22: cosmos.group.v1beta1.MsgExecResponse
	(*Member)(nil),                                      // 23: cosmos.group.v1beta1.Member
	(*anypb.Any)(nil),                                   // 24: google.protobuf.Any
	(Choice)(0),                                         // 25: cosmos.group.v1beta1.Choice
}
var file_cosmos_group_v1beta1_tx_proto_depIdxs = []int32{
	23, // 0: cosmos.group.v1beta1.MsgCreateGroupRequest.members:type_name -> cosmos.group.v1beta1.Member
	23, // 1: cosmos.group.v1beta1.MsgUpdateGroupMembersRequest.member_updates:type_name -> cosmos.group.v1beta1.Member
	24, // 2: cosmos.group.v1beta1.MsgCreateGroupAccountRequest.decision_policy:type_name -> google.protobuf.Any
	24, // 3: cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest.decision_policy:type_name -> google.protobuf.Any
	24, // 4: cosmos.group.v1beta1.MsgCreateProposalRequest.msgs:type_name -> google.protobuf.Any
	0,  // 5: cosmos.group.v1beta1.MsgCreateProposalRequest.exec:type_name -> cosmos.group.v1beta1.Exec
	25, // 6: cosmos.group.v1beta1.MsgVoteRequest.choice:type_name -> cosmos.group.v1beta1.Choice
	0,  // 7: cosmos.group.v1beta1.MsgVoteRequest.exec:type_name -> cosmos.group.v1beta1.Exec
	1,  // 8: cosmos.group.v1beta1.Msg.CreateGroup:input_type -> cosmos.group.v1beta1.MsgCreateGroupRequest
	3,  // 9: cosmos.group.v1beta1.Msg.UpdateGroupMembers:input_type -> cosmos.group.v1beta1.MsgUpdateGroupMembersRequest
	5,  // 10: cosmos.group.v1beta1.Msg.UpdateGroupAdmin:input_type -> cosmos.group.v1beta1.MsgUpdateGroupAdminRequest
	7,  // 11: cosmos.group.v1beta1.Msg.UpdateGroupMetadata:input_type -> cosmos.group.v1beta1.MsgUpdateGroupMetadataRequest
	9,  // 12: cosmos.group.v1beta1.Msg.CreateGroupAccount:input_type -> cosmos.group.v1beta1.MsgCreateGroupAccountRequest
	11, // 13: cosmos.group.v1beta1.Msg.UpdateGroupAccountAdmin:input_type -> cosmos.group.v1beta1.MsgUpdateGroupAccountAdminRequest
	13, // 14: cosmos.group.v1beta1.Msg.UpdateGroupAccountDecisionPolicy:input_type -> cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyRequest
	15, // 15: cosmos.group.v1beta1.Msg.UpdateGroupAccountMetadata:input_type -> cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataRequest
	17, // 16: cosmos.group.v1beta1.Msg.CreateProposal:input_type -> cosmos.group.v1beta1.MsgCreateProposalRequest
	19, // 17: cosmos.group.v1beta1.Msg.Vote:input_type -> cosmos.group.v1beta1.MsgVoteRequest
	21, // 18: cosmos.group.v1beta1.Msg.Exec:input_type -> cosmos.group.v1beta1.MsgExecRequest
	2,  // 19: cosmos.group.v1beta1.Msg.CreateGroup:output_type -> cosmos.group.v1beta1.MsgCreateGroupResponse
	4,  // 20: cosmos.group.v1beta1.Msg.UpdateGroupMembers:output_type -> cosmos.group.v1beta1.MsgUpdateGroupMembersResponse
	6,  // 21: cosmos.group.v1beta1.Msg.UpdateGroupAdmin:output_type -> cosmos.group.v1beta1.MsgUpdateGroupAdminResponse
	8,  // 22: cosmos.group.v1beta1.Msg.UpdateGroupMetadata:output_type -> cosmos.group.v1beta1.MsgUpdateGroupMetadataResponse
	10, // 23: cosmos.group.v1beta1.Msg.CreateGroupAccount:output_type -> cosmos.group.v1beta1.MsgCreateGroupAccountResponse
	12, // 24: cosmos.group.v1beta1.Msg.UpdateGroupAccountAdmin:output_type -> cosmos.group.v1beta1.MsgUpdateGroupAccountAdminResponse
	14, // 25: cosmos.group.v1beta1.Msg.UpdateGroupAccountDecisionPolicy:output_type -> cosmos.group.v1beta1.MsgUpdateGroupAccountDecisionPolicyResponse
	16, // 26: cosmos.group.v1beta1.Msg.UpdateGroupAccountMetadata:output_type -> cosmos.group.v1beta1.MsgUpdateGroupAccountMetadataResponse
	18, // 27: cosmos.group.v1beta1.Msg.CreateProposal:output_type -> cosmos.group.v1beta1.MsgCreateProposalResponse
	20, // 28: cosmos.group.v1beta1.Msg.Vote:output_type -> cosmos.group.v1beta1.MsgVoteResponse
	22, // 29: cosmos.group.v1beta1.Msg.Exec:output_type -> cosmos.group.v1beta1.MsgExecResponse
	19, // [19:30] is the sub-list for method output_type
	8,  // [8:19] is the sub-list for method input_type
	8,  // [8:8] is the sub-list for extension type_name
	8,  // [8:8] is the sub-list for extension extendee
	0,  // [0:8] is the sub-list for field type_name
}

func init() { file_cosmos_group_v1beta1_tx_proto_init() }
func file_cosmos_group_v1beta1_tx_proto_init() {
	if File_cosmos_group_v1beta1_tx_proto != nil {
		return
	}
	file_cosmos_group_v1beta1_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_cosmos_group_v1beta1_tx_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgCreateGroupRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgCreateGroupResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupMembersRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupMembersResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAdminRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAdminResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupMetadataRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupMetadataResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgCreateGroupAccountRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgCreateGroupAccountResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAccountAdminRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAccountAdminResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAccountDecisionPolicyRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAccountDecisionPolicyResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAccountMetadataRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[15].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgUpdateGroupAccountMetadataResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[16].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgCreateProposalRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[17].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgCreateProposalResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[18].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgVoteRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[19].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgVoteResponse); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[20].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgExecRequest); i {
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
		file_cosmos_group_v1beta1_tx_proto_msgTypes[21].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgExecResponse); i {
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
			RawDescriptor: file_cosmos_group_v1beta1_tx_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   22,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_cosmos_group_v1beta1_tx_proto_goTypes,
		DependencyIndexes: file_cosmos_group_v1beta1_tx_proto_depIdxs,
		EnumInfos:         file_cosmos_group_v1beta1_tx_proto_enumTypes,
		MessageInfos:      file_cosmos_group_v1beta1_tx_proto_msgTypes,
	}.Build()
	File_cosmos_group_v1beta1_tx_proto = out.File
	file_cosmos_group_v1beta1_tx_proto_rawDesc = nil
	file_cosmos_group_v1beta1_tx_proto_goTypes = nil
	file_cosmos_group_v1beta1_tx_proto_depIdxs = nil
}
