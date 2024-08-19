// Code generated by protoc-gen-go-pulsar. DO NOT EDIT.
package modulev1

import (
	_ "cosmossdk.io/api/amino"
	_ "cosmossdk.io/api/cosmos/app/v1alpha1"
	fmt "fmt"
	runtime "github.com/cosmos/cosmos-proto/runtime"
	_ "github.com/cosmos/gogoproto/gogoproto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoiface "google.golang.org/protobuf/runtime/protoiface"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	io "io"
	reflect "reflect"
	sync "sync"
)

var (
	md_Module                          protoreflect.MessageDescriptor
	fd_Module_max_execution_period     protoreflect.FieldDescriptor
	fd_Module_max_metadata_len         protoreflect.FieldDescriptor
	fd_Module_max_proposal_title_len   protoreflect.FieldDescriptor
	fd_Module_max_proposal_summary_len protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_group_module_v1_module_proto_init()
	md_Module = File_cosmos_group_module_v1_module_proto.Messages().ByName("Module")
	fd_Module_max_execution_period = md_Module.Fields().ByName("max_execution_period")
	fd_Module_max_metadata_len = md_Module.Fields().ByName("max_metadata_len")
	fd_Module_max_proposal_title_len = md_Module.Fields().ByName("max_proposal_title_len")
	fd_Module_max_proposal_summary_len = md_Module.Fields().ByName("max_proposal_summary_len")
}

var _ protoreflect.Message = (*fastReflection_Module)(nil)

type fastReflection_Module Module

func (x *Module) ProtoReflect() protoreflect.Message {
	return (*fastReflection_Module)(x)
}

func (x *Module) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_group_module_v1_module_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_Module_messageType fastReflection_Module_messageType
var _ protoreflect.MessageType = fastReflection_Module_messageType{}

type fastReflection_Module_messageType struct{}

func (x fastReflection_Module_messageType) Zero() protoreflect.Message {
	return (*fastReflection_Module)(nil)
}
func (x fastReflection_Module_messageType) New() protoreflect.Message {
	return new(fastReflection_Module)
}
func (x fastReflection_Module_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_Module
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_Module) Descriptor() protoreflect.MessageDescriptor {
	return md_Module
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_Module) Type() protoreflect.MessageType {
	return _fastReflection_Module_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_Module) New() protoreflect.Message {
	return new(fastReflection_Module)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_Module) Interface() protoreflect.ProtoMessage {
	return (*Module)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_Module) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.MaxExecutionPeriod != nil {
		value := protoreflect.ValueOfMessage(x.MaxExecutionPeriod.ProtoReflect())
		if !f(fd_Module_max_execution_period, value) {
			return
		}
	}
	if x.MaxMetadataLen != uint64(0) {
		value := protoreflect.ValueOfUint64(x.MaxMetadataLen)
		if !f(fd_Module_max_metadata_len, value) {
			return
		}
	}
	if x.MaxProposalTitleLen != uint64(0) {
		value := protoreflect.ValueOfUint64(x.MaxProposalTitleLen)
		if !f(fd_Module_max_proposal_title_len, value) {
			return
		}
	}
	if x.MaxProposalSummaryLen != uint64(0) {
		value := protoreflect.ValueOfUint64(x.MaxProposalSummaryLen)
		if !f(fd_Module_max_proposal_summary_len, value) {
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
func (x *fastReflection_Module) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.group.module.v1.Module.max_execution_period":
		return x.MaxExecutionPeriod != nil
	case "cosmos.group.module.v1.Module.max_metadata_len":
		return x.MaxMetadataLen != uint64(0)
	case "cosmos.group.module.v1.Module.max_proposal_title_len":
		return x.MaxProposalTitleLen != uint64(0)
	case "cosmos.group.module.v1.Module.max_proposal_summary_len":
		return x.MaxProposalSummaryLen != uint64(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.module.v1.Module"))
		}
		panic(fmt.Errorf("message cosmos.group.module.v1.Module does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_Module) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.group.module.v1.Module.max_execution_period":
		x.MaxExecutionPeriod = nil
	case "cosmos.group.module.v1.Module.max_metadata_len":
		x.MaxMetadataLen = uint64(0)
	case "cosmos.group.module.v1.Module.max_proposal_title_len":
		x.MaxProposalTitleLen = uint64(0)
	case "cosmos.group.module.v1.Module.max_proposal_summary_len":
		x.MaxProposalSummaryLen = uint64(0)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.module.v1.Module"))
		}
		panic(fmt.Errorf("message cosmos.group.module.v1.Module does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_Module) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.group.module.v1.Module.max_execution_period":
		value := x.MaxExecutionPeriod
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	case "cosmos.group.module.v1.Module.max_metadata_len":
		value := x.MaxMetadataLen
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.module.v1.Module.max_proposal_title_len":
		value := x.MaxProposalTitleLen
		return protoreflect.ValueOfUint64(value)
	case "cosmos.group.module.v1.Module.max_proposal_summary_len":
		value := x.MaxProposalSummaryLen
		return protoreflect.ValueOfUint64(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.module.v1.Module"))
		}
		panic(fmt.Errorf("message cosmos.group.module.v1.Module does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_Module) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.group.module.v1.Module.max_execution_period":
		x.MaxExecutionPeriod = value.Message().Interface().(*durationpb.Duration)
	case "cosmos.group.module.v1.Module.max_metadata_len":
		x.MaxMetadataLen = value.Uint()
	case "cosmos.group.module.v1.Module.max_proposal_title_len":
		x.MaxProposalTitleLen = value.Uint()
	case "cosmos.group.module.v1.Module.max_proposal_summary_len":
		x.MaxProposalSummaryLen = value.Uint()
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.module.v1.Module"))
		}
		panic(fmt.Errorf("message cosmos.group.module.v1.Module does not contain field %s", fd.FullName()))
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
func (x *fastReflection_Module) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.module.v1.Module.max_execution_period":
		if x.MaxExecutionPeriod == nil {
			x.MaxExecutionPeriod = new(durationpb.Duration)
		}
		return protoreflect.ValueOfMessage(x.MaxExecutionPeriod.ProtoReflect())
	case "cosmos.group.module.v1.Module.max_metadata_len":
		panic(fmt.Errorf("field max_metadata_len of message cosmos.group.module.v1.Module is not mutable"))
	case "cosmos.group.module.v1.Module.max_proposal_title_len":
		panic(fmt.Errorf("field max_proposal_title_len of message cosmos.group.module.v1.Module is not mutable"))
	case "cosmos.group.module.v1.Module.max_proposal_summary_len":
		panic(fmt.Errorf("field max_proposal_summary_len of message cosmos.group.module.v1.Module is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.module.v1.Module"))
		}
		panic(fmt.Errorf("message cosmos.group.module.v1.Module does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_Module) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.group.module.v1.Module.max_execution_period":
		m := new(durationpb.Duration)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	case "cosmos.group.module.v1.Module.max_metadata_len":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.module.v1.Module.max_proposal_title_len":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.group.module.v1.Module.max_proposal_summary_len":
		return protoreflect.ValueOfUint64(uint64(0))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.group.module.v1.Module"))
		}
		panic(fmt.Errorf("message cosmos.group.module.v1.Module does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_Module) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.group.module.v1.Module", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_Module) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_Module) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_Module) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_Module) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*Module)
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
		if x.MaxExecutionPeriod != nil {
			l = options.Size(x.MaxExecutionPeriod)
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.MaxMetadataLen != 0 {
			n += 1 + runtime.Sov(uint64(x.MaxMetadataLen))
		}
		if x.MaxProposalTitleLen != 0 {
			n += 1 + runtime.Sov(uint64(x.MaxProposalTitleLen))
		}
		if x.MaxProposalSummaryLen != 0 {
			n += 1 + runtime.Sov(uint64(x.MaxProposalSummaryLen))
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
		x := input.Message.Interface().(*Module)
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
		if x.MaxProposalSummaryLen != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.MaxProposalSummaryLen))
			i--
			dAtA[i] = 0x20
		}
		if x.MaxProposalTitleLen != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.MaxProposalTitleLen))
			i--
			dAtA[i] = 0x18
		}
		if x.MaxMetadataLen != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.MaxMetadataLen))
			i--
			dAtA[i] = 0x10
		}
		if x.MaxExecutionPeriod != nil {
			encoded, err := options.Marshal(x.MaxExecutionPeriod)
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
		x := input.Message.Interface().(*Module)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: Module: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: Module: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field MaxExecutionPeriod", wireType)
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
				if x.MaxExecutionPeriod == nil {
					x.MaxExecutionPeriod = &durationpb.Duration{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.MaxExecutionPeriod); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 2:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field MaxMetadataLen", wireType)
				}
				x.MaxMetadataLen = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.MaxMetadataLen |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field MaxProposalTitleLen", wireType)
				}
				x.MaxProposalTitleLen = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.MaxProposalTitleLen |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 4:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field MaxProposalSummaryLen", wireType)
				}
				x.MaxProposalSummaryLen = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.MaxProposalSummaryLen |= uint64(b&0x7F) << shift
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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.0
// 	protoc        (unknown)
// source: cosmos/group/module/v1/module.proto

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Module is the config object of the group module.
type Module struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// max_execution_period defines the max duration after a proposal's voting period ends that members can send a MsgExec
	// to execute the proposal.
	MaxExecutionPeriod *durationpb.Duration `protobuf:"bytes,1,opt,name=max_execution_period,json=maxExecutionPeriod,proto3" json:"max_execution_period,omitempty"`
	// MaxMetadataLen defines the max chars allowed in all
	// messages that allows creating or updating a group
	// with a metadata field
	// Defaults to 255 if not explicitly set.
	MaxMetadataLen uint64 `protobuf:"varint,2,opt,name=max_metadata_len,json=maxMetadataLen,proto3" json:"max_metadata_len,omitempty"`
	// MaxProposalTitleLen defines the max chars allowed
	// in string for the MsgSubmitProposal and Proposal
	// summary field
	// Defaults to 255 if not explicitly set.
	MaxProposalTitleLen uint64 `protobuf:"varint,3,opt,name=max_proposal_title_len,json=maxProposalTitleLen,proto3" json:"max_proposal_title_len,omitempty"`
	// MaxProposalSummaryLen defines the max chars allowed
	// in string for the MsgSubmitProposal and Proposal
	// summary field
	// Defaults to 10200 if not explicitly set.
	MaxProposalSummaryLen uint64 `protobuf:"varint,4,opt,name=max_proposal_summary_len,json=maxProposalSummaryLen,proto3" json:"max_proposal_summary_len,omitempty"`
}

func (x *Module) Reset() {
	*x = Module{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_group_module_v1_module_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Module) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Module) ProtoMessage() {}

// Deprecated: Use Module.ProtoReflect.Descriptor instead.
func (*Module) Descriptor() ([]byte, []int) {
	return file_cosmos_group_module_v1_module_proto_rawDescGZIP(), []int{0}
}

func (x *Module) GetMaxExecutionPeriod() *durationpb.Duration {
	if x != nil {
		return x.MaxExecutionPeriod
	}
	return nil
}

func (x *Module) GetMaxMetadataLen() uint64 {
	if x != nil {
		return x.MaxMetadataLen
	}
	return 0
}

func (x *Module) GetMaxProposalTitleLen() uint64 {
	if x != nil {
		return x.MaxProposalTitleLen
	}
	return 0
}

func (x *Module) GetMaxProposalSummaryLen() uint64 {
	if x != nil {
		return x.MaxProposalSummaryLen
	}
	return 0
}

var File_cosmos_group_module_v1_module_proto protoreflect.FileDescriptor

var file_cosmos_group_module_v1_module_proto_rawDesc = []byte{
	0x0a, 0x23, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2f, 0x6d,
	0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72,
	0x6f, 0x75, 0x70, 0x2e, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x20, 0x63,
	0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x61, 0x70, 0x70, 0x2f, 0x76, 0x31, 0x61, 0x6c, 0x70, 0x68,
	0x61, 0x31, 0x2f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x14, 0x67, 0x6f, 0x67, 0x6f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x67, 0x6f, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x11, 0x61, 0x6d, 0x69, 0x6e, 0x6f, 0x2f, 0x61, 0x6d, 0x69,
	0x6e, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x9a, 0x02, 0x0a, 0x06, 0x4d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x12, 0x5a, 0x0a, 0x14, 0x6d, 0x61, 0x78, 0x5f, 0x65, 0x78, 0x65, 0x63, 0x75,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x0d, 0xc8, 0xde,
	0x1f, 0x00, 0x98, 0xdf, 0x1f, 0x01, 0xa8, 0xe7, 0xb0, 0x2a, 0x01, 0x52, 0x12, 0x6d, 0x61, 0x78,
	0x45, 0x78, 0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12,
	0x28, 0x0a, 0x10, 0x6d, 0x61, 0x78, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x5f,
	0x6c, 0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0e, 0x6d, 0x61, 0x78, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x4c, 0x65, 0x6e, 0x12, 0x33, 0x0a, 0x16, 0x6d, 0x61, 0x78,
	0x5f, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x74, 0x69, 0x74, 0x6c, 0x65, 0x5f,
	0x6c, 0x65, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x13, 0x6d, 0x61, 0x78, 0x50, 0x72,
	0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x54, 0x69, 0x74, 0x6c, 0x65, 0x4c, 0x65, 0x6e, 0x12, 0x37,
	0x0a, 0x18, 0x6d, 0x61, 0x78, 0x5f, 0x70, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x5f, 0x73,
	0x75, 0x6d, 0x6d, 0x61, 0x72, 0x79, 0x5f, 0x6c, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x15, 0x6d, 0x61, 0x78, 0x50, 0x72, 0x6f, 0x70, 0x6f, 0x73, 0x61, 0x6c, 0x53, 0x75, 0x6d,
	0x6d, 0x61, 0x72, 0x79, 0x4c, 0x65, 0x6e, 0x3a, 0x1c, 0xba, 0xc0, 0x96, 0xda, 0x01, 0x16, 0x0a,
	0x14, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x73, 0x64, 0x6b, 0x2e, 0x69, 0x6f, 0x2f, 0x78, 0x2f,
	0x67, 0x72, 0x6f, 0x75, 0x70, 0x42, 0xd6, 0x01, 0x0a, 0x1a, 0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x67, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x6d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x30, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x73, 0x64, 0x6b, 0x2e, 0x69,
	0x6f, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x67, 0x72, 0x6f,
	0x75, 0x70, 0x2f, 0x6d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x6d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x47, 0x4d, 0xaa, 0x02, 0x16, 0x43, 0x6f,
	0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x2e, 0x4d, 0x6f, 0x64, 0x75, 0x6c,
	0x65, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x16, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x47, 0x72,
	0x6f, 0x75, 0x70, 0x5c, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x22,
	0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x47, 0x72, 0x6f, 0x75, 0x70, 0x5c, 0x4d, 0x6f, 0x64,
	0x75, 0x6c, 0x65, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x19, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x3a, 0x3a, 0x47, 0x72, 0x6f,
	0x75, 0x70, 0x3a, 0x3a, 0x4d, 0x6f, 0x64, 0x75, 0x6c, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cosmos_group_module_v1_module_proto_rawDescOnce sync.Once
	file_cosmos_group_module_v1_module_proto_rawDescData = file_cosmos_group_module_v1_module_proto_rawDesc
)

func file_cosmos_group_module_v1_module_proto_rawDescGZIP() []byte {
	file_cosmos_group_module_v1_module_proto_rawDescOnce.Do(func() {
		file_cosmos_group_module_v1_module_proto_rawDescData = protoimpl.X.CompressGZIP(file_cosmos_group_module_v1_module_proto_rawDescData)
	})
	return file_cosmos_group_module_v1_module_proto_rawDescData
}

var file_cosmos_group_module_v1_module_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_cosmos_group_module_v1_module_proto_goTypes = []interface{}{
	(*Module)(nil),              // 0: cosmos.group.module.v1.Module
	(*durationpb.Duration)(nil), // 1: google.protobuf.Duration
}
var file_cosmos_group_module_v1_module_proto_depIdxs = []int32{
	1, // 0: cosmos.group.module.v1.Module.max_execution_period:type_name -> google.protobuf.Duration
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_cosmos_group_module_v1_module_proto_init() }
func file_cosmos_group_module_v1_module_proto_init() {
	if File_cosmos_group_module_v1_module_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cosmos_group_module_v1_module_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Module); i {
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
			RawDescriptor: file_cosmos_group_module_v1_module_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cosmos_group_module_v1_module_proto_goTypes,
		DependencyIndexes: file_cosmos_group_module_v1_module_proto_depIdxs,
		MessageInfos:      file_cosmos_group_module_v1_module_proto_msgTypes,
	}.Build()
	File_cosmos_group_module_v1_module_proto = out.File
	file_cosmos_group_module_v1_module_proto_rawDesc = nil
	file_cosmos_group_module_v1_module_proto_goTypes = nil
	file_cosmos_group_module_v1_module_proto_depIdxs = nil
}