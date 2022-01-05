package snapshotsv1beta1

import (
	fmt "fmt"
	runtime "github.com/cosmos/cosmos-proto/runtime"
	_ "github.com/gogo/protobuf/gogoproto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoiface "google.golang.org/protobuf/runtime/protoiface"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	io "io"
	reflect "reflect"
	sync "sync"
)

var (
	md_Snapshot          protoreflect.MessageDescriptor
	fd_Snapshot_height   protoreflect.FieldDescriptor
	fd_Snapshot_format   protoreflect.FieldDescriptor
	fd_Snapshot_chunks   protoreflect.FieldDescriptor
	fd_Snapshot_hash     protoreflect.FieldDescriptor
	fd_Snapshot_metadata protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_init()
	md_Snapshot = File_cosmos_base_snapshots_v1beta1_snapshot_proto.Messages().ByName("Snapshot")
	fd_Snapshot_height = md_Snapshot.Fields().ByName("height")
	fd_Snapshot_format = md_Snapshot.Fields().ByName("format")
	fd_Snapshot_chunks = md_Snapshot.Fields().ByName("chunks")
	fd_Snapshot_hash = md_Snapshot.Fields().ByName("hash")
	fd_Snapshot_metadata = md_Snapshot.Fields().ByName("metadata")
}

var _ protoreflect.Message = (*fastReflection_Snapshot)(nil)

type fastReflection_Snapshot Snapshot

func (x *Snapshot) ProtoReflect() protoreflect.Message {
	return (*fastReflection_Snapshot)(x)
}

func (x *Snapshot) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_Snapshot_messageType fastReflection_Snapshot_messageType
var _ protoreflect.MessageType = fastReflection_Snapshot_messageType{}

type fastReflection_Snapshot_messageType struct{}

func (x fastReflection_Snapshot_messageType) Zero() protoreflect.Message {
	return (*fastReflection_Snapshot)(nil)
}
func (x fastReflection_Snapshot_messageType) New() protoreflect.Message {
	return new(fastReflection_Snapshot)
}
func (x fastReflection_Snapshot_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_Snapshot
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_Snapshot) Descriptor() protoreflect.MessageDescriptor {
	return md_Snapshot
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_Snapshot) Type() protoreflect.MessageType {
	return _fastReflection_Snapshot_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_Snapshot) New() protoreflect.Message {
	return new(fastReflection_Snapshot)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_Snapshot) Interface() protoreflect.ProtoMessage {
	return (*Snapshot)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_Snapshot) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Height != uint64(0) {
		value := protoreflect.ValueOfUint64(x.Height)
		if !f(fd_Snapshot_height, value) {
			return
		}
	}
	if x.Format != uint32(0) {
		value := protoreflect.ValueOfUint32(x.Format)
		if !f(fd_Snapshot_format, value) {
			return
		}
	}
	if x.Chunks != uint32(0) {
		value := protoreflect.ValueOfUint32(x.Chunks)
		if !f(fd_Snapshot_chunks, value) {
			return
		}
	}
	if len(x.Hash) != 0 {
		value := protoreflect.ValueOfBytes(x.Hash)
		if !f(fd_Snapshot_hash, value) {
			return
		}
	}
	if x.Metadata != nil {
		value := protoreflect.ValueOfMessage(x.Metadata.ProtoReflect())
		if !f(fd_Snapshot_metadata, value) {
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
func (x *fastReflection_Snapshot) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Snapshot.height":
		return x.Height != uint64(0)
	case "cosmos.base.snapshots.v1beta1.Snapshot.format":
		return x.Format != uint32(0)
	case "cosmos.base.snapshots.v1beta1.Snapshot.chunks":
		return x.Chunks != uint32(0)
	case "cosmos.base.snapshots.v1beta1.Snapshot.hash":
		return len(x.Hash) != 0
	case "cosmos.base.snapshots.v1beta1.Snapshot.metadata":
		return x.Metadata != nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Snapshot"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Snapshot does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_Snapshot) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Snapshot.height":
		x.Height = uint64(0)
	case "cosmos.base.snapshots.v1beta1.Snapshot.format":
		x.Format = uint32(0)
	case "cosmos.base.snapshots.v1beta1.Snapshot.chunks":
		x.Chunks = uint32(0)
	case "cosmos.base.snapshots.v1beta1.Snapshot.hash":
		x.Hash = nil
	case "cosmos.base.snapshots.v1beta1.Snapshot.metadata":
		x.Metadata = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Snapshot"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Snapshot does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_Snapshot) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.base.snapshots.v1beta1.Snapshot.height":
		value := x.Height
		return protoreflect.ValueOfUint64(value)
	case "cosmos.base.snapshots.v1beta1.Snapshot.format":
		value := x.Format
		return protoreflect.ValueOfUint32(value)
	case "cosmos.base.snapshots.v1beta1.Snapshot.chunks":
		value := x.Chunks
		return protoreflect.ValueOfUint32(value)
	case "cosmos.base.snapshots.v1beta1.Snapshot.hash":
		value := x.Hash
		return protoreflect.ValueOfBytes(value)
	case "cosmos.base.snapshots.v1beta1.Snapshot.metadata":
		value := x.Metadata
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Snapshot"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Snapshot does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_Snapshot) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Snapshot.height":
		x.Height = value.Uint()
	case "cosmos.base.snapshots.v1beta1.Snapshot.format":
		x.Format = uint32(value.Uint())
	case "cosmos.base.snapshots.v1beta1.Snapshot.chunks":
		x.Chunks = uint32(value.Uint())
	case "cosmos.base.snapshots.v1beta1.Snapshot.hash":
		x.Hash = value.Bytes()
	case "cosmos.base.snapshots.v1beta1.Snapshot.metadata":
		x.Metadata = value.Message().Interface().(*Metadata)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Snapshot"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Snapshot does not contain field %s", fd.FullName()))
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
func (x *fastReflection_Snapshot) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Snapshot.metadata":
		if x.Metadata == nil {
			x.Metadata = new(Metadata)
		}
		return protoreflect.ValueOfMessage(x.Metadata.ProtoReflect())
	case "cosmos.base.snapshots.v1beta1.Snapshot.height":
		panic(fmt.Errorf("field height of message cosmos.base.snapshots.v1beta1.Snapshot is not mutable"))
	case "cosmos.base.snapshots.v1beta1.Snapshot.format":
		panic(fmt.Errorf("field format of message cosmos.base.snapshots.v1beta1.Snapshot is not mutable"))
	case "cosmos.base.snapshots.v1beta1.Snapshot.chunks":
		panic(fmt.Errorf("field chunks of message cosmos.base.snapshots.v1beta1.Snapshot is not mutable"))
	case "cosmos.base.snapshots.v1beta1.Snapshot.hash":
		panic(fmt.Errorf("field hash of message cosmos.base.snapshots.v1beta1.Snapshot is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Snapshot"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Snapshot does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_Snapshot) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Snapshot.height":
		return protoreflect.ValueOfUint64(uint64(0))
	case "cosmos.base.snapshots.v1beta1.Snapshot.format":
		return protoreflect.ValueOfUint32(uint32(0))
	case "cosmos.base.snapshots.v1beta1.Snapshot.chunks":
		return protoreflect.ValueOfUint32(uint32(0))
	case "cosmos.base.snapshots.v1beta1.Snapshot.hash":
		return protoreflect.ValueOfBytes(nil)
	case "cosmos.base.snapshots.v1beta1.Snapshot.metadata":
		m := new(Metadata)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Snapshot"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Snapshot does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_Snapshot) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.base.snapshots.v1beta1.Snapshot", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_Snapshot) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_Snapshot) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_Snapshot) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_Snapshot) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*Snapshot)
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
		if x.Height != 0 {
			n += 1 + runtime.Sov(uint64(x.Height))
		}
		if x.Format != 0 {
			n += 1 + runtime.Sov(uint64(x.Format))
		}
		if x.Chunks != 0 {
			n += 1 + runtime.Sov(uint64(x.Chunks))
		}
		l = len(x.Hash)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Metadata != nil {
			l = options.Size(x.Metadata)
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
		x := input.Message.Interface().(*Snapshot)
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
		if x.Metadata != nil {
			encoded, err := options.Marshal(x.Metadata)
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
			dAtA[i] = 0x2a
		}
		if len(x.Hash) > 0 {
			i -= len(x.Hash)
			copy(dAtA[i:], x.Hash)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Hash)))
			i--
			dAtA[i] = 0x22
		}
		if x.Chunks != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Chunks))
			i--
			dAtA[i] = 0x18
		}
		if x.Format != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Format))
			i--
			dAtA[i] = 0x10
		}
		if x.Height != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Height))
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
		x := input.Message.Interface().(*Snapshot)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: Snapshot: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: Snapshot: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
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
					x.Height |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 2:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Format", wireType)
				}
				x.Format = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Format |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Chunks", wireType)
				}
				x.Chunks = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.Chunks |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 4:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Hash", wireType)
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
				x.Hash = append(x.Hash[:0], dAtA[iNdEx:postIndex]...)
				if x.Hash == nil {
					x.Hash = []byte{}
				}
				iNdEx = postIndex
			case 5:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Metadata", wireType)
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
				if x.Metadata == nil {
					x.Metadata = &Metadata{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Metadata); err != nil {
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

var _ protoreflect.List = (*_Metadata_1_list)(nil)

type _Metadata_1_list struct {
	list *[][]byte
}

func (x *_Metadata_1_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_Metadata_1_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfBytes((*x.list)[i])
}

func (x *_Metadata_1_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.Bytes()
	concreteValue := valueUnwrapped
	(*x.list)[i] = concreteValue
}

func (x *_Metadata_1_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.Bytes()
	concreteValue := valueUnwrapped
	*x.list = append(*x.list, concreteValue)
}

func (x *_Metadata_1_list) AppendMutable() protoreflect.Value {
	panic(fmt.Errorf("AppendMutable can not be called on message Metadata at list field ChunkHashes as it is not of Message kind"))
}

func (x *_Metadata_1_list) Truncate(n int) {
	*x.list = (*x.list)[:n]
}

func (x *_Metadata_1_list) NewElement() protoreflect.Value {
	var v []byte
	return protoreflect.ValueOfBytes(v)
}

func (x *_Metadata_1_list) IsValid() bool {
	return x.list != nil
}

var (
	md_Metadata              protoreflect.MessageDescriptor
	fd_Metadata_chunk_hashes protoreflect.FieldDescriptor
)

func init() {
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_init()
	md_Metadata = File_cosmos_base_snapshots_v1beta1_snapshot_proto.Messages().ByName("Metadata")
	fd_Metadata_chunk_hashes = md_Metadata.Fields().ByName("chunk_hashes")
}

var _ protoreflect.Message = (*fastReflection_Metadata)(nil)

type fastReflection_Metadata Metadata

func (x *Metadata) ProtoReflect() protoreflect.Message {
	return (*fastReflection_Metadata)(x)
}

func (x *Metadata) slowProtoReflect() protoreflect.Message {
	mi := &file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_Metadata_messageType fastReflection_Metadata_messageType
var _ protoreflect.MessageType = fastReflection_Metadata_messageType{}

type fastReflection_Metadata_messageType struct{}

func (x fastReflection_Metadata_messageType) Zero() protoreflect.Message {
	return (*fastReflection_Metadata)(nil)
}
func (x fastReflection_Metadata_messageType) New() protoreflect.Message {
	return new(fastReflection_Metadata)
}
func (x fastReflection_Metadata_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_Metadata
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_Metadata) Descriptor() protoreflect.MessageDescriptor {
	return md_Metadata
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_Metadata) Type() protoreflect.MessageType {
	return _fastReflection_Metadata_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_Metadata) New() protoreflect.Message {
	return new(fastReflection_Metadata)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_Metadata) Interface() protoreflect.ProtoMessage {
	return (*Metadata)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_Metadata) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if len(x.ChunkHashes) != 0 {
		value := protoreflect.ValueOfList(&_Metadata_1_list{list: &x.ChunkHashes})
		if !f(fd_Metadata_chunk_hashes, value) {
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
func (x *fastReflection_Metadata) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Metadata.chunk_hashes":
		return len(x.ChunkHashes) != 0
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Metadata"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Metadata does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_Metadata) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Metadata.chunk_hashes":
		x.ChunkHashes = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Metadata"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Metadata does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_Metadata) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "cosmos.base.snapshots.v1beta1.Metadata.chunk_hashes":
		if len(x.ChunkHashes) == 0 {
			return protoreflect.ValueOfList(&_Metadata_1_list{})
		}
		listValue := &_Metadata_1_list{list: &x.ChunkHashes}
		return protoreflect.ValueOfList(listValue)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Metadata"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Metadata does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_Metadata) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Metadata.chunk_hashes":
		lv := value.List()
		clv := lv.(*_Metadata_1_list)
		x.ChunkHashes = *clv.list
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Metadata"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Metadata does not contain field %s", fd.FullName()))
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
func (x *fastReflection_Metadata) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Metadata.chunk_hashes":
		if x.ChunkHashes == nil {
			x.ChunkHashes = [][]byte{}
		}
		value := &_Metadata_1_list{list: &x.ChunkHashes}
		return protoreflect.ValueOfList(value)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Metadata"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Metadata does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_Metadata) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "cosmos.base.snapshots.v1beta1.Metadata.chunk_hashes":
		list := [][]byte{}
		return protoreflect.ValueOfList(&_Metadata_1_list{list: &list})
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: cosmos.base.snapshots.v1beta1.Metadata"))
		}
		panic(fmt.Errorf("message cosmos.base.snapshots.v1beta1.Metadata does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_Metadata) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in cosmos.base.snapshots.v1beta1.Metadata", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_Metadata) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_Metadata) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_Metadata) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_Metadata) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*Metadata)
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
		if len(x.ChunkHashes) > 0 {
			for _, b := range x.ChunkHashes {
				l = len(b)
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
		x := input.Message.Interface().(*Metadata)
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
		if len(x.ChunkHashes) > 0 {
			for iNdEx := len(x.ChunkHashes) - 1; iNdEx >= 0; iNdEx-- {
				i -= len(x.ChunkHashes[iNdEx])
				copy(dAtA[i:], x.ChunkHashes[iNdEx])
				i = runtime.EncodeVarint(dAtA, i, uint64(len(x.ChunkHashes[iNdEx])))
				i--
				dAtA[i] = 0xa
			}
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
		x := input.Message.Interface().(*Metadata)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: Metadata: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: Metadata: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field ChunkHashes", wireType)
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
				x.ChunkHashes = append(x.ChunkHashes, make([]byte, postIndex-iNdEx))
				copy(x.ChunkHashes[len(x.ChunkHashes)-1], dAtA[iNdEx:postIndex])
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
// source: cosmos/base/snapshots/v1beta1/snapshot.proto

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Snapshot contains Tendermint state sync snapshot info.
type Snapshot struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Height   uint64    `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	Format   uint32    `protobuf:"varint,2,opt,name=format,proto3" json:"format,omitempty"`
	Chunks   uint32    `protobuf:"varint,3,opt,name=chunks,proto3" json:"chunks,omitempty"`
	Hash     []byte    `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
	Metadata *Metadata `protobuf:"bytes,5,opt,name=metadata,proto3" json:"metadata,omitempty"`
}

func (x *Snapshot) Reset() {
	*x = Snapshot{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Snapshot) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Snapshot) ProtoMessage() {}

// Deprecated: Use Snapshot.ProtoReflect.Descriptor instead.
func (*Snapshot) Descriptor() ([]byte, []int) {
	return file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescGZIP(), []int{0}
}

func (x *Snapshot) GetHeight() uint64 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *Snapshot) GetFormat() uint32 {
	if x != nil {
		return x.Format
	}
	return 0
}

func (x *Snapshot) GetChunks() uint32 {
	if x != nil {
		return x.Chunks
	}
	return 0
}

func (x *Snapshot) GetHash() []byte {
	if x != nil {
		return x.Hash
	}
	return nil
}

func (x *Snapshot) GetMetadata() *Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

// Metadata contains SDK-specific snapshot metadata.
type Metadata struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ChunkHashes [][]byte `protobuf:"bytes,1,rep,name=chunk_hashes,json=chunkHashes,proto3" json:"chunk_hashes,omitempty"` // SHA-256 chunk hashes
}

func (x *Metadata) Reset() {
	*x = Metadata{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Metadata) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Metadata) ProtoMessage() {}

// Deprecated: Use Metadata.ProtoReflect.Descriptor instead.
func (*Metadata) Descriptor() ([]byte, []int) {
	return file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescGZIP(), []int{1}
}

func (x *Metadata) GetChunkHashes() [][]byte {
	if x != nil {
		return x.ChunkHashes
	}
	return nil
}

var File_cosmos_base_snapshots_v1beta1_snapshot_proto protoreflect.FileDescriptor

var file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x2f, 0x73, 0x6e,
	0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2f,
	0x73, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x1d,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x73, 0x6e, 0x61, 0x70,
	0x73, 0x68, 0x6f, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x1a, 0x14, 0x67,
	0x6f, 0x67, 0x6f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6f, 0x67, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xb1, 0x01, 0x0a, 0x08, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x06, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d,
	0x61, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0d,
	0x52, 0x06, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x49, 0x0a, 0x08,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x27,
	0x2e, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x73, 0x6e, 0x61,
	0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x2e, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x42, 0x04, 0xc8, 0xde, 0x1f, 0x00, 0x52, 0x08, 0x6d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x22, 0x2d, 0x0a, 0x08, 0x4d, 0x65, 0x74, 0x61, 0x64,
	0x61, 0x74, 0x61, 0x12, 0x21, 0x0a, 0x0c, 0x63, 0x68, 0x75, 0x6e, 0x6b, 0x5f, 0x68, 0x61, 0x73,
	0x68, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x0b, 0x63, 0x68, 0x75, 0x6e, 0x6b,
	0x48, 0x61, 0x73, 0x68, 0x65, 0x73, 0x42, 0x9a, 0x02, 0x0a, 0x21, 0x63, 0x6f, 0x6d, 0x2e, 0x63,
	0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x73, 0x6e, 0x61, 0x70, 0x73,
	0x68, 0x6f, 0x74, 0x73, 0x2e, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x42, 0x0d, 0x53, 0x6e,
	0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x4f, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73,
	0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2d, 0x73, 0x64, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x2f, 0x73, 0x6e, 0x61, 0x70,
	0x73, 0x68, 0x6f, 0x74, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0x3b, 0x73, 0x6e,
	0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x76, 0x31, 0x62, 0x65, 0x74, 0x61, 0x31, 0xa2, 0x02,
	0x03, 0x43, 0x42, 0x53, 0xaa, 0x02, 0x1d, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2e, 0x42, 0x61,
	0x73, 0x65, 0x2e, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x2e, 0x56, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0xca, 0x02, 0x1d, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x42, 0x61,
	0x73, 0x65, 0x5c, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x5c, 0x56, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0xe2, 0x02, 0x29, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x5c, 0x42, 0x61,
	0x73, 0x65, 0x5c, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x5c, 0x56, 0x31, 0x62,
	0x65, 0x74, 0x61, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x20, 0x43, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x3a, 0x3a, 0x42, 0x61, 0x73, 0x65, 0x3a,
	0x3a, 0x53, 0x6e, 0x61, 0x70, 0x73, 0x68, 0x6f, 0x74, 0x73, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x65,
	0x74, 0x61, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescOnce sync.Once
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescData = file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDesc
)

func file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescGZIP() []byte {
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescOnce.Do(func() {
		file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescData = protoimpl.X.CompressGZIP(file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescData)
	})
	return file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDescData
}

var file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_cosmos_base_snapshots_v1beta1_snapshot_proto_goTypes = []interface{}{
	(*Snapshot)(nil), // 0: cosmos.base.snapshots.v1beta1.Snapshot
	(*Metadata)(nil), // 1: cosmos.base.snapshots.v1beta1.Metadata
}
var file_cosmos_base_snapshots_v1beta1_snapshot_proto_depIdxs = []int32{
	1, // 0: cosmos.base.snapshots.v1beta1.Snapshot.metadata:type_name -> cosmos.base.snapshots.v1beta1.Metadata
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_cosmos_base_snapshots_v1beta1_snapshot_proto_init() }
func file_cosmos_base_snapshots_v1beta1_snapshot_proto_init() {
	if File_cosmos_base_snapshots_v1beta1_snapshot_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Snapshot); i {
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
		file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Metadata); i {
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
			RawDescriptor: file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cosmos_base_snapshots_v1beta1_snapshot_proto_goTypes,
		DependencyIndexes: file_cosmos_base_snapshots_v1beta1_snapshot_proto_depIdxs,
		MessageInfos:      file_cosmos_base_snapshots_v1beta1_snapshot_proto_msgTypes,
	}.Build()
	File_cosmos_base_snapshots_v1beta1_snapshot_proto = out.File
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_rawDesc = nil
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_goTypes = nil
	file_cosmos_base_snapshots_v1beta1_snapshot_proto_depIdxs = nil
}
