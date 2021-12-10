package ormv1alpha1

import (
	fmt "fmt"
	io "io"
	reflect "reflect"
	sync "sync"

	runtime "github.com/cosmos/cosmos-proto/runtime"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoiface "google.golang.org/protobuf/runtime/protoiface"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	descriptorpb "google.golang.org/protobuf/types/descriptorpb"
)

var _ protoreflect.List = (*_SchemaDescriptor_2_list)(nil)

type _SchemaDescriptor_2_list struct {
	list *[]*SchemaDescriptor_ModuleEntry
}

func (x *_SchemaDescriptor_2_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_SchemaDescriptor_2_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfMessage((*x.list)[i].ProtoReflect())
}

func (x *_SchemaDescriptor_2_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*SchemaDescriptor_ModuleEntry)
	(*x.list)[i] = concreteValue
}

func (x *_SchemaDescriptor_2_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*SchemaDescriptor_ModuleEntry)
	*x.list = append(*x.list, concreteValue)
}

func (x *_SchemaDescriptor_2_list) AppendMutable() protoreflect.Value {
	v := new(SchemaDescriptor_ModuleEntry)
	*x.list = append(*x.list, v)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_SchemaDescriptor_2_list) Truncate(n int) {
	for i := n; i < len(*x.list); i++ {
		(*x.list)[i] = nil
	}
	*x.list = (*x.list)[:n]
}

func (x *_SchemaDescriptor_2_list) NewElement() protoreflect.Value {
	v := new(SchemaDescriptor_ModuleEntry)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_SchemaDescriptor_2_list) IsValid() bool {
	return x.list != nil
}

var (
	md_SchemaDescriptor         protoreflect.MessageDescriptor
	fd_SchemaDescriptor_files   protoreflect.FieldDescriptor
	fd_SchemaDescriptor_modules protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_orm_v1alpha1_schema_proto_init()
	md_SchemaDescriptor = File_cosmos_orm_v1alpha1_schema_proto.Messages().ByName("SchemaDescriptor")
	fd_SchemaDescriptor_files = md_SchemaDescriptor.Fields().ByName("files")
	fd_SchemaDescriptor_modules = md_SchemaDescriptor.Fields().ByName("modules")
}

var _ protoreflect.Message = (*fastReflection_SchemaDescriptor)(nil)

type fastReflection_SchemaDescriptor SchemaDescriptor

func (x *SchemaDescriptor) ProtoReflect() protoreflect.Message {
	return (*fastReflection_SchemaDescriptor)(x)
}

func (x *SchemaDescriptor) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_orm_v1alpha1_schema_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_SchemaDescriptor_messageType fastReflection_SchemaDescriptor_messageType
var _ protoreflect.MessageType = fastReflection_SchemaDescriptor_messageType{}

type fastReflection_SchemaDescriptor_messageType struct{}

func (x fastReflection_SchemaDescriptor_messageType) Zero() protoreflect.Message {
	return (*fastReflection_SchemaDescriptor)(nil)
}
func (x fastReflection_SchemaDescriptor_messageType) New() protoreflect.Message {
	return new(fastReflection_SchemaDescriptor)
}
func (x fastReflection_SchemaDescriptor_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_SchemaDescriptor
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_SchemaDescriptor) Descriptor() protoreflect.MessageDescriptor {
	return md_SchemaDescriptor
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_SchemaDescriptor) Type() protoreflect.MessageType {
	return _fastReflection_SchemaDescriptor_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_SchemaDescriptor) New() protoreflect.Message {
	return new(fastReflection_SchemaDescriptor)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_SchemaDescriptor) Interface() protoreflect.ProtoMessage {
	return (*SchemaDescriptor)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_SchemaDescriptor) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Files != nil {
		value := protoreflect.ValueOfMessage(x.Files.ProtoReflect())
		if !f(fd_SchemaDescriptor_files, value) {
			return
		}
	}
	if len(x.Modules) != 0 {
		value := protoreflect.ValueOfList(&_SchemaDescriptor_2_list{list: &x.Modules})
		if !f(fd_SchemaDescriptor_modules, value) {
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
func (x *fastReflection_SchemaDescriptor) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.files":
		return x.Files != nil
	case "cosmos.orm.v1alpha1.SchemaDescriptor.modules":
		return len(x.Modules) != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_SchemaDescriptor) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.files":
		x.Files = nil
	case "cosmos.orm.v1alpha1.SchemaDescriptor.modules":
		x.Modules = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_SchemaDescriptor) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.files":
		value := x.Files
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	case "cosmos.orm.v1alpha1.SchemaDescriptor.modules":
		if len(x.Modules) == 0 {
			return protoreflect.ValueOfList(&_SchemaDescriptor_2_list{})
		}
		listValue := &_SchemaDescriptor_2_list{list: &x.Modules}
		return protoreflect.ValueOfList(listValue)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_SchemaDescriptor) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.files":
		x.Files = value.Message().Interface().(*descriptorpb.FileDescriptorSet)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.modules":
		lv := value.List()
		clv := lv.(*_SchemaDescriptor_2_list)
		x.Modules = *clv.list
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor does not contain field %s", fd.FullName()))
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
func (x *fastReflection_SchemaDescriptor) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.files":
		if x.Files == nil {
			x.Files = new(descriptorpb.FileDescriptorSet)
		}
		return protoreflect.ValueOfMessage(x.Files.ProtoReflect())
	case "cosmos.orm.v1alpha1.SchemaDescriptor.modules":
		if x.Modules == nil {
			x.Modules = []*SchemaDescriptor_ModuleEntry{}
		}
		value := &_SchemaDescriptor_2_list{list: &x.Modules}
		return protoreflect.ValueOfList(value)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_SchemaDescriptor) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.files":
		m := new(descriptorpb.FileDescriptorSet)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	case "cosmos.orm.v1alpha1.SchemaDescriptor.modules":
		list := []*SchemaDescriptor_ModuleEntry{}
		return protoreflect.ValueOfList(&_SchemaDescriptor_2_list{list: &list})
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_SchemaDescriptor) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.orm.v1alpha1.SchemaDescriptor", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_SchemaDescriptor) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_SchemaDescriptor) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_SchemaDescriptor) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_SchemaDescriptor) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*SchemaDescriptor)
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
		if x.Files != nil {
			l = options.Size(x.Files)
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if len(x.Modules) > 0 {
			for _, e := range x.Modules {
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
		x := input.Message.Interface().(*SchemaDescriptor)
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
		if len(x.Modules) > 0 {
			for iNdEx := len(x.Modules) - 1; iNdEx >= 0; iNdEx-- {
				encoded, err := options.Marshal(x.Modules[iNdEx])
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
		if x.Files != nil {
			encoded, err := options.Marshal(x.Files)
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
		x := input.Message.Interface().(*SchemaDescriptor)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: SchemaDescriptor: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: SchemaDescriptor: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Files", wireType)
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
				if x.Files == nil {
					x.Files = &descriptorpb.FileDescriptorSet{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Files); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Modules", wireType)
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
				x.Modules = append(x.Modules, &SchemaDescriptor_ModuleEntry{})
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Modules[len(x.Modules)-1]); err != nil {
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

var _ protoreflect.List = (*_SchemaDescriptor_ModuleEntry_3_list)(nil)

type _SchemaDescriptor_ModuleEntry_3_list struct {
	list *[]*SchemaDescriptor_FileEntry
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfMessage((*x.list)[i].ProtoReflect())
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*SchemaDescriptor_FileEntry)
	(*x.list)[i] = concreteValue
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.Message()
	concreteValue := valueUnwrapped.Interface().(*SchemaDescriptor_FileEntry)
	*x.list = append(*x.list, concreteValue)
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) AppendMutable() protoreflect.Value {
	v := new(SchemaDescriptor_FileEntry)
	*x.list = append(*x.list, v)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) Truncate(n int) {
	for i := n; i < len(*x.list); i++ {
		(*x.list)[i] = nil
	}
	*x.list = (*x.list)[:n]
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) NewElement() protoreflect.Value {
	v := new(SchemaDescriptor_FileEntry)
	return protoreflect.ValueOfMessage(v.ProtoReflect())
}

func (x *_SchemaDescriptor_ModuleEntry_3_list) IsValid() bool {
	return x.list != nil
}

var (
	md_SchemaDescriptor_ModuleEntry        protoreflect.MessageDescriptor
	fd_SchemaDescriptor_ModuleEntry_name   protoreflect.FieldDescriptor
	fd_SchemaDescriptor_ModuleEntry_prefix protoreflect.FieldDescriptor
	fd_SchemaDescriptor_ModuleEntry_files  protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_orm_v1alpha1_schema_proto_init()
	md_SchemaDescriptor_ModuleEntry = File_cosmos_orm_v1alpha1_schema_proto.Messages().ByName("SchemaDescriptor").Messages().ByName("ModuleEntry")
	fd_SchemaDescriptor_ModuleEntry_name = md_SchemaDescriptor_ModuleEntry.Fields().ByName("name")
	fd_SchemaDescriptor_ModuleEntry_prefix = md_SchemaDescriptor_ModuleEntry.Fields().ByName("prefix")
	fd_SchemaDescriptor_ModuleEntry_files = md_SchemaDescriptor_ModuleEntry.Fields().ByName("files")
}

var _ protoreflect.Message = (*fastReflection_SchemaDescriptor_ModuleEntry)(nil)

type fastReflection_SchemaDescriptor_ModuleEntry SchemaDescriptor_ModuleEntry

func (x *SchemaDescriptor_ModuleEntry) ProtoReflect() protoreflect.Message {
	return (*fastReflection_SchemaDescriptor_ModuleEntry)(x)
}

func (x *SchemaDescriptor_ModuleEntry) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_orm_v1alpha1_schema_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_SchemaDescriptor_ModuleEntry_messageType fastReflection_SchemaDescriptor_ModuleEntry_messageType
var _ protoreflect.MessageType = fastReflection_SchemaDescriptor_ModuleEntry_messageType{}

type fastReflection_SchemaDescriptor_ModuleEntry_messageType struct{}

func (x fastReflection_SchemaDescriptor_ModuleEntry_messageType) Zero() protoreflect.Message {
	return (*fastReflection_SchemaDescriptor_ModuleEntry)(nil)
}
func (x fastReflection_SchemaDescriptor_ModuleEntry_messageType) New() protoreflect.Message {
	return new(fastReflection_SchemaDescriptor_ModuleEntry)
}
func (x fastReflection_SchemaDescriptor_ModuleEntry_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_SchemaDescriptor_ModuleEntry
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Descriptor() protoreflect.MessageDescriptor {
	return md_SchemaDescriptor_ModuleEntry
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Type() protoreflect.MessageType {
	return _fastReflection_SchemaDescriptor_ModuleEntry_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) New() protoreflect.Message {
	return new(fastReflection_SchemaDescriptor_ModuleEntry)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Interface() protoreflect.ProtoMessage {
	return (*SchemaDescriptor_ModuleEntry)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Name != "" {
		value := protoreflect.ValueOfString(x.Name)
		if !f(fd_SchemaDescriptor_ModuleEntry_name, value) {
			return
		}
	}
	if len(x.Prefix) != 0 {
		value := protoreflect.ValueOfBytes(x.Prefix)
		if !f(fd_SchemaDescriptor_ModuleEntry_prefix, value) {
			return
		}
	}
	if len(x.Files) != 0 {
		value := protoreflect.ValueOfList(&_SchemaDescriptor_ModuleEntry_3_list{list: &x.Files})
		if !f(fd_SchemaDescriptor_ModuleEntry_files, value) {
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
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.name":
		return x.Name != ""
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.prefix":
		return len(x.Prefix) != 0
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.files":
		return len(x.Files) != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.name":
		x.Name = ""
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.prefix":
		x.Prefix = nil
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.files":
		x.Files = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.name":
		value := x.Name
		return protoreflect.ValueOfString(value)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.prefix":
		value := x.Prefix
		return protoreflect.ValueOfBytes(value)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.files":
		if len(x.Files) == 0 {
			return protoreflect.ValueOfList(&_SchemaDescriptor_ModuleEntry_3_list{})
		}
		listValue := &_SchemaDescriptor_ModuleEntry_3_list{list: &x.Files}
		return protoreflect.ValueOfList(listValue)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.name":
		x.Name = value.Interface().(string)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.prefix":
		x.Prefix = value.Bytes()
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.files":
		lv := value.List()
		clv := lv.(*_SchemaDescriptor_ModuleEntry_3_list)
		x.Files = *clv.list
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry does not contain field %s", fd.FullName()))
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
func (x *fastReflection_SchemaDescriptor_ModuleEntry) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.files":
		if x.Files == nil {
			x.Files = []*SchemaDescriptor_FileEntry{}
		}
		value := &_SchemaDescriptor_ModuleEntry_3_list{list: &x.Files}
		return protoreflect.ValueOfList(value)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.name":
		panic(fmt.Errorf("field name of message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry is not mutable"))
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.prefix":
		panic(fmt.Errorf("field prefix of message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.name":
		return protoreflect.ValueOfString("")
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.prefix":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.files":
		list := []*SchemaDescriptor_FileEntry{}
		return protoreflect.ValueOfList(&_SchemaDescriptor_ModuleEntry_3_list{list: &list})
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_SchemaDescriptor_ModuleEntry) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_SchemaDescriptor_ModuleEntry) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*SchemaDescriptor_ModuleEntry)
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
		l = len(x.Name)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Prefix)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if len(x.Files) > 0 {
			for _, e := range x.Files {
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
		x := input.Message.Interface().(*SchemaDescriptor_ModuleEntry)
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
		if len(x.Files) > 0 {
			for iNdEx := len(x.Files) - 1; iNdEx >= 0; iNdEx-- {
				encoded, err := options.Marshal(x.Files[iNdEx])
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
		if len(x.Prefix) > 0 {
			i -= len(x.Prefix)
			copy(dAtA[i:], x.Prefix)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Prefix)))
			i--
			dAtA[i] = 0x12
		}
		if len(x.Name) > 0 {
			i -= len(x.Name)
			copy(dAtA[i:], x.Name)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Name)))
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
		x := input.Message.Interface().(*SchemaDescriptor_ModuleEntry)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: SchemaDescriptor_ModuleEntry: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: SchemaDescriptor_ModuleEntry: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
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
				x.Name = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Prefix", wireType)
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
				x.Prefix = append(x.Prefix[:0], dAtA[iNdEx:postIndex]...)
				if x.Prefix == nil {
					x.Prefix = []byte{}
				}
				iNdEx = postIndex
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Files", wireType)
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
				x.Files = append(x.Files, &SchemaDescriptor_FileEntry{})
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Files[len(x.Files)-1]); err != nil {
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
	md_SchemaDescriptor_FileEntry           protoreflect.MessageDescriptor
	fd_SchemaDescriptor_FileEntry_id        protoreflect.FieldDescriptor
	fd_SchemaDescriptor_FileEntry_file_name protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_orm_v1alpha1_schema_proto_init()
	md_SchemaDescriptor_FileEntry = File_cosmos_orm_v1alpha1_schema_proto.Messages().ByName("SchemaDescriptor").Messages().ByName("FileEntry")
	fd_SchemaDescriptor_FileEntry_id = md_SchemaDescriptor_FileEntry.Fields().ByName("id")
	fd_SchemaDescriptor_FileEntry_file_name = md_SchemaDescriptor_FileEntry.Fields().ByName("file_name")
}

var _ protoreflect.Message = (*fastReflection_SchemaDescriptor_FileEntry)(nil)

type fastReflection_SchemaDescriptor_FileEntry SchemaDescriptor_FileEntry

func (x *SchemaDescriptor_FileEntry) ProtoReflect() protoreflect.Message {
	return (*fastReflection_SchemaDescriptor_FileEntry)(x)
}

func (x *SchemaDescriptor_FileEntry) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_orm_v1alpha1_schema_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_SchemaDescriptor_FileEntry_messageType fastReflection_SchemaDescriptor_FileEntry_messageType
var _ protoreflect.MessageType = fastReflection_SchemaDescriptor_FileEntry_messageType{}

type fastReflection_SchemaDescriptor_FileEntry_messageType struct{}

func (x fastReflection_SchemaDescriptor_FileEntry_messageType) Zero() protoreflect.Message {
	return (*fastReflection_SchemaDescriptor_FileEntry)(nil)
}
func (x fastReflection_SchemaDescriptor_FileEntry_messageType) New() protoreflect.Message {
	return new(fastReflection_SchemaDescriptor_FileEntry)
}
func (x fastReflection_SchemaDescriptor_FileEntry_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_SchemaDescriptor_FileEntry
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_SchemaDescriptor_FileEntry) Descriptor() protoreflect.MessageDescriptor {
	return md_SchemaDescriptor_FileEntry
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_SchemaDescriptor_FileEntry) Type() protoreflect.MessageType {
	return _fastReflection_SchemaDescriptor_FileEntry_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_SchemaDescriptor_FileEntry) New() protoreflect.Message {
	return new(fastReflection_SchemaDescriptor_FileEntry)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_SchemaDescriptor_FileEntry) Interface() protoreflect.ProtoMessage {
	return (*SchemaDescriptor_FileEntry)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_SchemaDescriptor_FileEntry) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Id != uint32(0) {
		value := protoreflect.ValueOfUint32(x.Id)
		if !f(fd_SchemaDescriptor_FileEntry_id, value) {
			return
		}
	}
	if x.FileName != "" {
		value := protoreflect.ValueOfString(x.FileName)
		if !f(fd_SchemaDescriptor_FileEntry_file_name, value) {
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
func (x *fastReflection_SchemaDescriptor_FileEntry) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.id":
		return x.Id != uint32(0)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.file_name":
		return x.FileName != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_SchemaDescriptor_FileEntry) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.id":
		x.Id = uint32(0)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.file_name":
		x.FileName = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_SchemaDescriptor_FileEntry) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.id":
		value := x.Id
		return protoreflect.ValueOfUint32(value)
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.file_name":
		value := x.FileName
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_SchemaDescriptor_FileEntry) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.id":
		x.Id = uint32(value.Uint())
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.file_name":
		x.FileName = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry does not contain field %s", fd.FullName()))
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
func (x *fastReflection_SchemaDescriptor_FileEntry) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.id":
		panic(fmt.Errorf("field id of message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry is not mutable"))
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.file_name":
		panic(fmt.Errorf("field file_name of message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_SchemaDescriptor_FileEntry) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.id":
		return protoreflect.ValueOfUint32(uint32(0))
	case "cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry.file_name":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry"))
		}
		panic(fmt.Errorf("message cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_SchemaDescriptor_FileEntry) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_SchemaDescriptor_FileEntry) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_SchemaDescriptor_FileEntry) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_SchemaDescriptor_FileEntry) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_SchemaDescriptor_FileEntry) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*SchemaDescriptor_FileEntry)
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
		if x.Id != 0 {
			n += 1 + runtime.Sov(uint64(x.Id))
		}
		l = len(x.FileName)
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
		x := input.Message.Interface().(*SchemaDescriptor_FileEntry)
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
		if len(x.FileName) > 0 {
			i -= len(x.FileName)
			copy(dAtA[i:], x.FileName)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.FileName)))
			i--
			dAtA[i] = 0x12
		}
		if x.Id != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Id))
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
		x := input.Message.Interface().(*SchemaDescriptor_FileEntry)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: SchemaDescriptor_FileEntry: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: SchemaDescriptor_FileEntry: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
				}
				x.Id = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Id |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field FileName", wireType)
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
				x.FileName = string(dAtA[iNdEx:postIndex])
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
// 	protoc        v3.19.1
// source: cosmos/orm/v1alpha1/schema.proto

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// SchemaDescriptor describes an ORM schema that contains all the information
// needed for a dynamic client to decode the stored data.
type SchemaDescriptor struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// files is the set of all FileDescriptorProto's needed to decode the stored data.
	// A schema imposes the constraint that every file and every table within that
	// schema have at most one instance in the store.
	Files *descriptorpb.FileDescriptorSet `protobuf:"bytes,1,opt,name=files,proto3" json:"files,omitempty"`
	// modules is the set of modules in the schema.
	Modules []*SchemaDescriptor_ModuleEntry `protobuf:"bytes,2,rep,name=modules,proto3" json:"modules,omitempty"`
}

func (x *SchemaDescriptor) Reset() {
	*x = SchemaDescriptor{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_orm_v1alpha1_schema_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SchemaDescriptor) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SchemaDescriptor) ProtoMessage() {}

// Deprecated: Use SchemaDescriptor.ProtoReflect.Descriptor instead.
func (*SchemaDescriptor) Descriptor() ([]byte, []int) {
	return file_cosmos_orm_v1alpha1_schema_proto_rawDescGZIP(), []int{0}
}

func (x *SchemaDescriptor) GetFiles() *descriptorpb.FileDescriptorSet {
	if x != nil {
		return x.Files
	}
	return nil
}

func (x *SchemaDescriptor) GetModules() []*SchemaDescriptor_ModuleEntry {
	if x != nil {
		return x.Modules
	}
	return nil
}

// ModuleEntry describes a single module's schema.
type SchemaDescriptor_ModuleEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// name is the name of the module. In the multi-store model this name is
	// used to locate the module's store.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// prefix is an optional prefix that precedes all keys in this module's
	// store.
	Prefix []byte `protobuf:"bytes,2,opt,name=prefix,proto3" json:"prefix,omitempty"`
	// files describes the schema files used in this module.
	Files []*SchemaDescriptor_FileEntry `protobuf:"bytes,3,rep,name=files,proto3" json:"files,omitempty"`
}

func (x *SchemaDescriptor_ModuleEntry) Reset() {
	*x = SchemaDescriptor_ModuleEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_orm_v1alpha1_schema_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SchemaDescriptor_ModuleEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SchemaDescriptor_ModuleEntry) ProtoMessage() {}

// Deprecated: Use SchemaDescriptor_ModuleEntry.ProtoReflect.Descriptor instead.
func (*SchemaDescriptor_ModuleEntry) Descriptor() ([]byte, []int) {
	return file_cosmos_orm_v1alpha1_schema_proto_rawDescGZIP(), []int{0, 0}
}

func (x *SchemaDescriptor_ModuleEntry) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *SchemaDescriptor_ModuleEntry) GetPrefix() []byte {
	if x != nil {
		return x.Prefix
	}
	return nil
}

func (x *SchemaDescriptor_ModuleEntry) GetFiles() []*SchemaDescriptor_FileEntry {
	if x != nil {
		return x.Files
	}
	return nil
}

// FileEntry describes an ORM file used in a module.
type SchemaDescriptor_FileEntry struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id is a prefix that will be varint encoded and prepended to all the
	// table keys specified in the file's tables.
	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// file_name is the name of a file in the FileDescriptor set that contains
	// table definitions.
	FileName string `protobuf:"bytes,2,opt,name=file_name,json=fileName,proto3" json:"file_name,omitempty"`
}

func (x *SchemaDescriptor_FileEntry) Reset() {
	*x = SchemaDescriptor_FileEntry{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_orm_v1alpha1_schema_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SchemaDescriptor_FileEntry) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SchemaDescriptor_FileEntry) ProtoMessage() {}

// Deprecated: Use SchemaDescriptor_FileEntry.ProtoReflect.Descriptor instead.
func (*SchemaDescriptor_FileEntry) Descriptor() ([]byte, []int) {
	return file_cosmos_orm_v1alpha1_schema_proto_rawDescGZIP(), []int{0, 1}
}

func (x *SchemaDescriptor_FileEntry) GetId() uint32 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *SchemaDescriptor_FileEntry) GetFileName() string {
	if x != nil {
		return x.FileName
	}
	return ""
}

var File_cosmos_orm_v1alpha1_schema_proto protoreflect.FileDescriptor

var file_cosmos_orm_v1alpha1_schema_proto_rawDesc = []byte{
	0x0a, 0x20, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x6f, 0x72, 0x6d, 0x2f, 0x76, 0x31, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x13, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x6f, 0x72, 0x6d, 0x2e, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd6, 0x02, 0x0a, 0x10, 0x53, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x12, 0x38,
	0x0a, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x22, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x46, 0x69, 0x6c, 0x65, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x53, 0x65,
	0x74, 0x52, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x12, 0x4b, 0x0a, 0x07, 0x6d, 0x6f, 0x64, 0x75,
	0x6c, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x31, 0x2e, 0x63, 0x6f, 0x73, 0x6d,
	0x6f, 0x73, 0x2e, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e,
	0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72,
	0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x07, 0x6d, 0x6f,
	0x64, 0x75, 0x6c, 0x65, 0x73, 0x1a, 0x80, 0x01, 0x0a, 0x0b, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x72, 0x65,
	0x66, 0x69, 0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x70, 0x72, 0x65, 0x66, 0x69,
	0x78, 0x12, 0x45, 0x0a, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x2f, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31,
	0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x44, 0x65, 0x73,
	0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x46, 0x69, 0x6c, 0x65, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x05, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x1a, 0x38, 0x0a, 0x09, 0x46, 0x69, 0x6c, 0x65,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0d, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x66, 0x69, 0x6c, 0x65, 0x5f, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x69, 0x6c, 0x65, 0x4e, 0x61,
	0x6d, 0x65, 0x42, 0xd6, 0x01, 0x0a, 0x17, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f,
	0x73, 0x2e, 0x6f, 0x72, 0x6d, 0x2e, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x42, 0x0b,
	0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x40, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2d, 0x73, 0x64, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x6f, 0x72, 0x6d, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70,
	0x68, 0x61, 0x31, 0x3b, 0x6f, 0x72, 0x6d, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xa2,
	0x02, 0x03, 0x43, 0x4f, 0x58, 0xaa, 0x02, 0x13, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x4f,
	0x72, 0x6d, 0x2e, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xca, 0x02, 0x13, 0x43, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x4f, 0x72, 0x6d, 0x5c, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61,
	0x31, 0xe2, 0x02, 0x1f, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x4f, 0x72, 0x6d, 0x5c, 0x56,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0xea, 0x02, 0x15, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x3a, 0x3a, 0x4f, 0x72,
	0x6d, 0x3a, 0x3a, 0x56, 0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_cosmos_orm_v1alpha1_schema_proto_rawDescOnce sync.Once
	file_cosmos_orm_v1alpha1_schema_proto_rawDescData = file_cosmos_orm_v1alpha1_schema_proto_rawDesc
)

func file_cosmos_orm_v1alpha1_schema_proto_rawDescGZIP() []byte {
	file_cosmos_orm_v1alpha1_schema_proto_rawDescOnce.Do(func() {
		file_cosmos_orm_v1alpha1_schema_proto_rawDescData = protoimpl.X.CompressGZIP(file_cosmos_orm_v1alpha1_schema_proto_rawDescData)
	})
	return file_cosmos_orm_v1alpha1_schema_proto_rawDescData
}

var file_cosmos_orm_v1alpha1_schema_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_cosmos_orm_v1alpha1_schema_proto_goTypes = []interface{}{
	(*SchemaDescriptor)(nil),               // 0: cosmos.orm.v1alpha1.SchemaDescriptor
	(*SchemaDescriptor_ModuleEntry)(nil),   // 1: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry
	(*SchemaDescriptor_FileEntry)(nil),     // 2: cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry
	(*descriptorpb.FileDescriptorSet)(nil), // 3: google.protobuf.FileDescriptorSet
}
var file_cosmos_orm_v1alpha1_schema_proto_depIdxs = []int32{
	3, // 0: cosmos.orm.v1alpha1.SchemaDescriptor.files:type_name -> google.protobuf.FileDescriptorSet
	1, // 1: cosmos.orm.v1alpha1.SchemaDescriptor.modules:type_name -> cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry
	2, // 2: cosmos.orm.v1alpha1.SchemaDescriptor.ModuleEntry.files:type_name -> cosmos.orm.v1alpha1.SchemaDescriptor.FileEntry
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_cosmos_orm_v1alpha1_schema_proto_init() }
func file_cosmos_orm_v1alpha1_schema_proto_init() {
	if File_cosmos_orm_v1alpha1_schema_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cosmos_orm_v1alpha1_schema_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SchemaDescriptor); i {
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
		file_cosmos_orm_v1alpha1_schema_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SchemaDescriptor_ModuleEntry); i {
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
		file_cosmos_orm_v1alpha1_schema_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SchemaDescriptor_FileEntry); i {
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
			RawDescriptor: file_cosmos_orm_v1alpha1_schema_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cosmos_orm_v1alpha1_schema_proto_goTypes,
		DependencyIndexes: file_cosmos_orm_v1alpha1_schema_proto_depIdxs,
		MessageInfos:      file_cosmos_orm_v1alpha1_schema_proto_msgTypes,
	}.Build()
	File_cosmos_orm_v1alpha1_schema_proto = out.File
	file_cosmos_orm_v1alpha1_schema_proto_rawDesc = nil
	file_cosmos_orm_v1alpha1_schema_proto_goTypes = nil
	file_cosmos_orm_v1alpha1_schema_proto_depIdxs = nil
}
