package testpb

import (
	binary "encoding/binary"
	fmt "fmt"
	io "io"
	reflect "reflect"
	sort "sort"
	sync "sync"

	runtime "github.com/cosmos/cosmos-proto/runtime"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoiface "google.golang.org/protobuf/runtime/protoiface"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	_ "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
)

var _ protoreflect.List = (*_A_17_list)(nil)

type _A_17_list struct {
	list *[]uint32
}

func (x *_A_17_list) Len() int {
	if x.list == nil {
		return 0
	}
	return len(*x.list)
}

func (x *_A_17_list) Get(i int) protoreflect.Value {
	return protoreflect.ValueOfUint32((*x.list)[i])
}

func (x *_A_17_list) Set(i int, value protoreflect.Value) {
	valueUnwrapped := value.Uint()
	concreteValue := (uint32)(valueUnwrapped)
	(*x.list)[i] = concreteValue
}

func (x *_A_17_list) Append(value protoreflect.Value) {
	valueUnwrapped := value.Uint()
	concreteValue := (uint32)(valueUnwrapped)
	*x.list = append(*x.list, concreteValue)
}

func (x *_A_17_list) AppendMutable() protoreflect.Value {
	panic(fmt.Errorf("AppendMutable can not be called on message A at list field Repeated as it is not of Message kind"))
}

func (x *_A_17_list) Truncate(n int) {
	*x.list = (*x.list)[:n]
}

func (x *_A_17_list) NewElement() protoreflect.Value {
	v := uint32(0)
	return protoreflect.ValueOfUint32(v)
}

func (x *_A_17_list) IsValid() bool {
	return x.list != nil
}

var _ protoreflect.Map = (*_A_18_map)(nil)

type _A_18_map struct {
	m *map[string]uint32
}

func (x *_A_18_map) Len() int {
	if x.m == nil {
		return 0
	}
	return len(*x.m)
}

func (x *_A_18_map) Range(f func(protoreflect.MapKey, protoreflect.Value) bool) {
	if x.m == nil {
		return
	}
	for k, v := range *x.m {
		mapKey := (protoreflect.MapKey)(protoreflect.ValueOfString(k))
		mapValue := protoreflect.ValueOfUint32(v)
		if !f(mapKey, mapValue) {
			break
		}
	}
}

func (x *_A_18_map) Has(key protoreflect.MapKey) bool {
	if x.m == nil {
		return false
	}
	keyUnwrapped := key.String()
	concreteValue := keyUnwrapped
	_, ok := (*x.m)[concreteValue]
	return ok
}

func (x *_A_18_map) Clear(key protoreflect.MapKey) {
	if x.m == nil {
		return
	}
	keyUnwrapped := key.String()
	concreteKey := keyUnwrapped
	delete(*x.m, concreteKey)
}

func (x *_A_18_map) Get(key protoreflect.MapKey) protoreflect.Value {
	if x.m == nil {
		return protoreflect.Value{}
	}
	keyUnwrapped := key.String()
	concreteKey := keyUnwrapped
	v, ok := (*x.m)[concreteKey]
	if !ok {
		return protoreflect.Value{}
	}
	return protoreflect.ValueOfUint32(v)
}

func (x *_A_18_map) Set(key protoreflect.MapKey, value protoreflect.Value) {
	if !key.IsValid() || !value.IsValid() {
		panic("invalid key or value provided")
	}
	keyUnwrapped := key.String()
	concreteKey := keyUnwrapped
	valueUnwrapped := value.Uint()
	concreteValue := (uint32)(valueUnwrapped)
	(*x.m)[concreteKey] = concreteValue
}

func (x *_A_18_map) Mutable(key protoreflect.MapKey) protoreflect.Value {
	panic("should not call Mutable on protoreflect.Map whose value is not of type protoreflect.Message")
}

func (x *_A_18_map) NewValue() protoreflect.Value {
	v := uint32(0)
	return protoreflect.ValueOfUint32(v)
}

func (x *_A_18_map) IsValid() bool {
	return x.m != nil
}

var (
	md_A          protoreflect.MessageDescriptor
	fd_A_u32      protoreflect.FieldDescriptor
	fd_A_u64      protoreflect.FieldDescriptor
	fd_A_str      protoreflect.FieldDescriptor
	fd_A_bz       protoreflect.FieldDescriptor
	fd_A_ts       protoreflect.FieldDescriptor
	fd_A_dur      protoreflect.FieldDescriptor
	fd_A_i32      protoreflect.FieldDescriptor
	fd_A_s32      protoreflect.FieldDescriptor
	fd_A_sf32     protoreflect.FieldDescriptor
	fd_A_i64      protoreflect.FieldDescriptor
	fd_A_s64      protoreflect.FieldDescriptor
	fd_A_sf64     protoreflect.FieldDescriptor
	fd_A_f32      protoreflect.FieldDescriptor
	fd_A_f64      protoreflect.FieldDescriptor
	fd_A_b        protoreflect.FieldDescriptor
	fd_A_e        protoreflect.FieldDescriptor
	fd_A_repeated protoreflect.FieldDescriptor
	fd_A_map      protoreflect.FieldDescriptor
	fd_A_msg      protoreflect.FieldDescriptor
	fd_A_oneof    protoreflect.FieldDescriptor
)

func init() {
	file_testpb_testschema_proto_init()
	md_A = File_testpb_testschema_proto.Messages().ByName("A")
	fd_A_u32 = md_A.Fields().ByName("u32")
	fd_A_u64 = md_A.Fields().ByName("u64")
	fd_A_str = md_A.Fields().ByName("str")
	fd_A_bz = md_A.Fields().ByName("bz")
	fd_A_ts = md_A.Fields().ByName("ts")
	fd_A_dur = md_A.Fields().ByName("dur")
	fd_A_i32 = md_A.Fields().ByName("i32")
	fd_A_s32 = md_A.Fields().ByName("s32")
	fd_A_sf32 = md_A.Fields().ByName("sf32")
	fd_A_i64 = md_A.Fields().ByName("i64")
	fd_A_s64 = md_A.Fields().ByName("s64")
	fd_A_sf64 = md_A.Fields().ByName("sf64")
	fd_A_f32 = md_A.Fields().ByName("f32")
	fd_A_f64 = md_A.Fields().ByName("f64")
	fd_A_b = md_A.Fields().ByName("b")
	fd_A_e = md_A.Fields().ByName("e")
	fd_A_repeated = md_A.Fields().ByName("repeated")
	fd_A_map = md_A.Fields().ByName("map")
	fd_A_msg = md_A.Fields().ByName("msg")
	fd_A_oneof = md_A.Fields().ByName("oneof")
}

var _ protoreflect.Message = (*fastReflection_A)(nil)

type fastReflection_A A

func (x *A) ProtoReflect() protoreflect.Message {
	return (*fastReflection_A)(x)
}

func (x *A) slowProtoReflect() protoreflect.Message {
	mi := &file_testpb_testschema_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_A_messageType fastReflection_A_messageType
var _ protoreflect.MessageType = fastReflection_A_messageType{}

type fastReflection_A_messageType struct{}

func (x fastReflection_A_messageType) Zero() protoreflect.Message {
	return (*fastReflection_A)(nil)
}
func (x fastReflection_A_messageType) New() protoreflect.Message {
	return new(fastReflection_A)
}
func (x fastReflection_A_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_A
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_A) Descriptor() protoreflect.MessageDescriptor {
	return md_A
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_A) Type() protoreflect.MessageType {
	return _fastReflection_A_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_A) New() protoreflect.Message {
	return new(fastReflection_A)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_A) Interface() protoreflect.ProtoMessage {
	return (*A)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_A) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.U32 != uint32(0) {
		value := protoreflect.ValueOfUint32(x.U32)
		if !f(fd_A_u32, value) {
			return
		}
	}
	if x.U64 != uint64(0) {
		value := protoreflect.ValueOfUint64(x.U64)
		if !f(fd_A_u64, value) {
			return
		}
	}
	if x.Str != "" {
		value := protoreflect.ValueOfString(x.Str)
		if !f(fd_A_str, value) {
			return
		}
	}
	if len(x.Bz) != 0 {
		value := protoreflect.ValueOfBytes(x.Bz)
		if !f(fd_A_bz, value) {
			return
		}
	}
	if x.Ts != nil {
		value := protoreflect.ValueOfMessage(x.Ts.ProtoReflect())
		if !f(fd_A_ts, value) {
			return
		}
	}
	if x.Dur != nil {
		value := protoreflect.ValueOfMessage(x.Dur.ProtoReflect())
		if !f(fd_A_dur, value) {
			return
		}
	}
	if x.I32 != int32(0) {
		value := protoreflect.ValueOfInt32(x.I32)
		if !f(fd_A_i32, value) {
			return
		}
	}
	if x.S32 != int32(0) {
		value := protoreflect.ValueOfInt32(x.S32)
		if !f(fd_A_s32, value) {
			return
		}
	}
	if x.Sf32 != int32(0) {
		value := protoreflect.ValueOfInt32(x.Sf32)
		if !f(fd_A_sf32, value) {
			return
		}
	}
	if x.I64 != int64(0) {
		value := protoreflect.ValueOfInt64(x.I64)
		if !f(fd_A_i64, value) {
			return
		}
	}
	if x.S64 != int64(0) {
		value := protoreflect.ValueOfInt64(x.S64)
		if !f(fd_A_s64, value) {
			return
		}
	}
	if x.Sf64 != int64(0) {
		value := protoreflect.ValueOfInt64(x.Sf64)
		if !f(fd_A_sf64, value) {
			return
		}
	}
	if x.F32 != uint32(0) {
		value := protoreflect.ValueOfUint32(x.F32)
		if !f(fd_A_f32, value) {
			return
		}
	}
	if x.F64 != uint64(0) {
		value := protoreflect.ValueOfUint64(x.F64)
		if !f(fd_A_f64, value) {
			return
		}
	}
	if x.B != false {
		value := protoreflect.ValueOfBool(x.B)
		if !f(fd_A_b, value) {
			return
		}
	}
	if x.E != 0 {
		value := protoreflect.ValueOfEnum((protoreflect.EnumNumber)(x.E))
		if !f(fd_A_e, value) {
			return
		}
	}
	if len(x.Repeated) != 0 {
		value := protoreflect.ValueOfList(&_A_17_list{list: &x.Repeated})
		if !f(fd_A_repeated, value) {
			return
		}
	}
	if len(x.Map) != 0 {
		value := protoreflect.ValueOfMap(&_A_18_map{m: &x.Map})
		if !f(fd_A_map, value) {
			return
		}
	}
	if x.Msg != nil {
		value := protoreflect.ValueOfMessage(x.Msg.ProtoReflect())
		if !f(fd_A_msg, value) {
			return
		}
	}
	if x.Sum != nil {
		switch o := x.Sum.(type) {
		case *A_Oneof:
			v := o.Oneof
			value := protoreflect.ValueOfUint32(v)
			if !f(fd_A_oneof, value) {
				return
			}
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
func (x *fastReflection_A) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "testpb.A.u32":
		return x.U32 != uint32(0)
	case "testpb.A.u64":
		return x.U64 != uint64(0)
	case "testpb.A.str":
		return x.Str != ""
	case "testpb.A.bz":
		return len(x.Bz) != 0
	case "testpb.A.ts":
		return x.Ts != nil
	case "testpb.A.dur":
		return x.Dur != nil
	case "testpb.A.i32":
		return x.I32 != int32(0)
	case "testpb.A.s32":
		return x.S32 != int32(0)
	case "testpb.A.sf32":
		return x.Sf32 != int32(0)
	case "testpb.A.i64":
		return x.I64 != int64(0)
	case "testpb.A.s64":
		return x.S64 != int64(0)
	case "testpb.A.sf64":
		return x.Sf64 != int64(0)
	case "testpb.A.f32":
		return x.F32 != uint32(0)
	case "testpb.A.f64":
		return x.F64 != uint64(0)
	case "testpb.A.b":
		return x.B != false
	case "testpb.A.e":
		return x.E != 0
	case "testpb.A.repeated":
		return len(x.Repeated) != 0
	case "testpb.A.map":
		return len(x.Map) != 0
	case "testpb.A.msg":
		return x.Msg != nil
	case "testpb.A.oneof":
		if x.Sum == nil {
			return false
		} else if _, ok := x.Sum.(*A_Oneof); ok {
			return true
		} else {
			return false
		}
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.A"))
		}
		panic(fmt.Errorf("message testpb.A does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_A) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "testpb.A.u32":
		x.U32 = uint32(0)
	case "testpb.A.u64":
		x.U64 = uint64(0)
	case "testpb.A.str":
		x.Str = ""
	case "testpb.A.bz":
		x.Bz = nil
	case "testpb.A.ts":
		x.Ts = nil
	case "testpb.A.dur":
		x.Dur = nil
	case "testpb.A.i32":
		x.I32 = int32(0)
	case "testpb.A.s32":
		x.S32 = int32(0)
	case "testpb.A.sf32":
		x.Sf32 = int32(0)
	case "testpb.A.i64":
		x.I64 = int64(0)
	case "testpb.A.s64":
		x.S64 = int64(0)
	case "testpb.A.sf64":
		x.Sf64 = int64(0)
	case "testpb.A.f32":
		x.F32 = uint32(0)
	case "testpb.A.f64":
		x.F64 = uint64(0)
	case "testpb.A.b":
		x.B = false
	case "testpb.A.e":
		x.E = 0
	case "testpb.A.repeated":
		x.Repeated = nil
	case "testpb.A.map":
		x.Map = nil
	case "testpb.A.msg":
		x.Msg = nil
	case "testpb.A.oneof":
		x.Sum = nil
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.A"))
		}
		panic(fmt.Errorf("message testpb.A does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_A) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "testpb.A.u32":
		value := x.U32
		return protoreflect.ValueOfUint32(value)
	case "testpb.A.u64":
		value := x.U64
		return protoreflect.ValueOfUint64(value)
	case "testpb.A.str":
		value := x.Str
		return protoreflect.ValueOfString(value)
	case "testpb.A.bz":
		value := x.Bz
		return protoreflect.ValueOfBytes(value)
	case "testpb.A.ts":
		value := x.Ts
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	case "testpb.A.dur":
		value := x.Dur
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	case "testpb.A.i32":
		value := x.I32
		return protoreflect.ValueOfInt32(value)
	case "testpb.A.s32":
		value := x.S32
		return protoreflect.ValueOfInt32(value)
	case "testpb.A.sf32":
		value := x.Sf32
		return protoreflect.ValueOfInt32(value)
	case "testpb.A.i64":
		value := x.I64
		return protoreflect.ValueOfInt64(value)
	case "testpb.A.s64":
		value := x.S64
		return protoreflect.ValueOfInt64(value)
	case "testpb.A.sf64":
		value := x.Sf64
		return protoreflect.ValueOfInt64(value)
	case "testpb.A.f32":
		value := x.F32
		return protoreflect.ValueOfUint32(value)
	case "testpb.A.f64":
		value := x.F64
		return protoreflect.ValueOfUint64(value)
	case "testpb.A.b":
		value := x.B
		return protoreflect.ValueOfBool(value)
	case "testpb.A.e":
		value := x.E
		return protoreflect.ValueOfEnum((protoreflect.EnumNumber)(value))
	case "testpb.A.repeated":
		if len(x.Repeated) == 0 {
			return protoreflect.ValueOfList(&_A_17_list{})
		}
		listValue := &_A_17_list{list: &x.Repeated}
		return protoreflect.ValueOfList(listValue)
	case "testpb.A.map":
		if len(x.Map) == 0 {
			return protoreflect.ValueOfMap(&_A_18_map{})
		}
		mapValue := &_A_18_map{m: &x.Map}
		return protoreflect.ValueOfMap(mapValue)
	case "testpb.A.msg":
		value := x.Msg
		return protoreflect.ValueOfMessage(value.ProtoReflect())
	case "testpb.A.oneof":
		if x.Sum == nil {
			return protoreflect.ValueOfUint32(uint32(0))
		} else if v, ok := x.Sum.(*A_Oneof); ok {
			return protoreflect.ValueOfUint32(v.Oneof)
		} else {
			return protoreflect.ValueOfUint32(uint32(0))
		}
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.A"))
		}
		panic(fmt.Errorf("message testpb.A does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_A) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "testpb.A.u32":
		x.U32 = uint32(value.Uint())
	case "testpb.A.u64":
		x.U64 = value.Uint()
	case "testpb.A.str":
		x.Str = value.Interface().(string)
	case "testpb.A.bz":
		x.Bz = value.Bytes()
	case "testpb.A.ts":
		x.Ts = value.Message().Interface().(*timestamppb.Timestamp)
	case "testpb.A.dur":
		x.Dur = value.Message().Interface().(*durationpb.Duration)
	case "testpb.A.i32":
		x.I32 = int32(value.Int())
	case "testpb.A.s32":
		x.S32 = int32(value.Int())
	case "testpb.A.sf32":
		x.Sf32 = int32(value.Int())
	case "testpb.A.i64":
		x.I64 = value.Int()
	case "testpb.A.s64":
		x.S64 = value.Int()
	case "testpb.A.sf64":
		x.Sf64 = value.Int()
	case "testpb.A.f32":
		x.F32 = uint32(value.Uint())
	case "testpb.A.f64":
		x.F64 = value.Uint()
	case "testpb.A.b":
		x.B = value.Bool()
	case "testpb.A.e":
		x.E = (Enum)(value.Enum())
	case "testpb.A.repeated":
		lv := value.List()
		clv := lv.(*_A_17_list)
		x.Repeated = *clv.list
	case "testpb.A.map":
		mv := value.Map()
		cmv := mv.(*_A_18_map)
		x.Map = *cmv.m
	case "testpb.A.msg":
		x.Msg = value.Message().Interface().(*B)
	case "testpb.A.oneof":
		cv := uint32(value.Uint())
		x.Sum = &A_Oneof{Oneof: cv}
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.A"))
		}
		panic(fmt.Errorf("message testpb.A does not contain field %s", fd.FullName()))
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
func (x *fastReflection_A) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "testpb.A.ts":
		if x.Ts == nil {
			x.Ts = new(timestamppb.Timestamp)
		}
		return protoreflect.ValueOfMessage(x.Ts.ProtoReflect())
	case "testpb.A.dur":
		if x.Dur == nil {
			x.Dur = new(durationpb.Duration)
		}
		return protoreflect.ValueOfMessage(x.Dur.ProtoReflect())
	case "testpb.A.repeated":
		if x.Repeated == nil {
			x.Repeated = []uint32{}
		}
		value := &_A_17_list{list: &x.Repeated}
		return protoreflect.ValueOfList(value)
	case "testpb.A.map":
		if x.Map == nil {
			x.Map = make(map[string]uint32)
		}
		value := &_A_18_map{m: &x.Map}
		return protoreflect.ValueOfMap(value)
	case "testpb.A.msg":
		if x.Msg == nil {
			x.Msg = new(B)
		}
		return protoreflect.ValueOfMessage(x.Msg.ProtoReflect())
	case "testpb.A.u32":
		panic(fmt.Errorf("field u32 of message testpb.A is not mutable"))
	case "testpb.A.u64":
		panic(fmt.Errorf("field u64 of message testpb.A is not mutable"))
	case "testpb.A.str":
		panic(fmt.Errorf("field str of message testpb.A is not mutable"))
	case "testpb.A.bz":
		panic(fmt.Errorf("field bz of message testpb.A is not mutable"))
	case "testpb.A.i32":
		panic(fmt.Errorf("field i32 of message testpb.A is not mutable"))
	case "testpb.A.s32":
		panic(fmt.Errorf("field s32 of message testpb.A is not mutable"))
	case "testpb.A.sf32":
		panic(fmt.Errorf("field sf32 of message testpb.A is not mutable"))
	case "testpb.A.i64":
		panic(fmt.Errorf("field i64 of message testpb.A is not mutable"))
	case "testpb.A.s64":
		panic(fmt.Errorf("field s64 of message testpb.A is not mutable"))
	case "testpb.A.sf64":
		panic(fmt.Errorf("field sf64 of message testpb.A is not mutable"))
	case "testpb.A.f32":
		panic(fmt.Errorf("field f32 of message testpb.A is not mutable"))
	case "testpb.A.f64":
		panic(fmt.Errorf("field f64 of message testpb.A is not mutable"))
	case "testpb.A.b":
		panic(fmt.Errorf("field b of message testpb.A is not mutable"))
	case "testpb.A.e":
		panic(fmt.Errorf("field e of message testpb.A is not mutable"))
	case "testpb.A.oneof":
		panic(fmt.Errorf("field oneof of message testpb.A is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.A"))
		}
		panic(fmt.Errorf("message testpb.A does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_A) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "testpb.A.u32":
		return protoreflect.ValueOfUint32(uint32(0))
	case "testpb.A.u64":
		return protoreflect.ValueOfUint64(uint64(0))
	case "testpb.A.str":
		return protoreflect.ValueOfString("")
	case "testpb.A.bz":
		return protoreflect.ValueOfBytes(nil)
	case "testpb.A.ts":
		m := new(timestamppb.Timestamp)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	case "testpb.A.dur":
		m := new(durationpb.Duration)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	case "testpb.A.i32":
		return protoreflect.ValueOfInt32(int32(0))
	case "testpb.A.s32":
		return protoreflect.ValueOfInt32(int32(0))
	case "testpb.A.sf32":
		return protoreflect.ValueOfInt32(int32(0))
	case "testpb.A.i64":
		return protoreflect.ValueOfInt64(int64(0))
	case "testpb.A.s64":
		return protoreflect.ValueOfInt64(int64(0))
	case "testpb.A.sf64":
		return protoreflect.ValueOfInt64(int64(0))
	case "testpb.A.f32":
		return protoreflect.ValueOfUint32(uint32(0))
	case "testpb.A.f64":
		return protoreflect.ValueOfUint64(uint64(0))
	case "testpb.A.b":
		return protoreflect.ValueOfBool(false)
	case "testpb.A.e":
		return protoreflect.ValueOfEnum(0)
	case "testpb.A.repeated":
		list := []uint32{}
		return protoreflect.ValueOfList(&_A_17_list{list: &list})
	case "testpb.A.map":
		m := make(map[string]uint32)
		return protoreflect.ValueOfMap(&_A_18_map{m: &m})
	case "testpb.A.msg":
		m := new(B)
		return protoreflect.ValueOfMessage(m.ProtoReflect())
	case "testpb.A.oneof":
		return protoreflect.ValueOfUint32(uint32(0))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.A"))
		}
		panic(fmt.Errorf("message testpb.A does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_A) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	case "testpb.A.sum":
		if x.Sum == nil {
			return nil
		}
		switch x.Sum.(type) {
		case *A_Oneof:
			return x.Descriptor().Fields().ByName("oneof")
		}
	default:
		panic(fmt.Errorf("%s is not a oneof field in testpb.A", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_A) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_A) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_A) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_A) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*A)
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
		if x.U32 != 0 {
			n += 1 + runtime.Sov(uint64(x.U32))
		}
		if x.U64 != 0 {
			n += 1 + runtime.Sov(uint64(x.U64))
		}
		l = len(x.Str)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		l = len(x.Bz)
		if l > 0 {
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Ts != nil {
			l = options.Size(x.Ts)
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.Dur != nil {
			l = options.Size(x.Dur)
			n += 1 + l + runtime.Sov(uint64(l))
		}
		if x.I32 != 0 {
			n += 1 + runtime.Sov(uint64(x.I32))
		}
		if x.S32 != 0 {
			n += 1 + runtime.Soz(uint64(x.S32))
		}
		if x.Sf32 != 0 {
			n += 5
		}
		if x.I64 != 0 {
			n += 1 + runtime.Sov(uint64(x.I64))
		}
		if x.S64 != 0 {
			n += 1 + runtime.Soz(uint64(x.S64))
		}
		if x.Sf64 != 0 {
			n += 9
		}
		if x.F32 != 0 {
			n += 5
		}
		if x.F64 != 0 {
			n += 9
		}
		if x.B {
			n += 2
		}
		if x.E != 0 {
			n += 2 + runtime.Sov(uint64(x.E))
		}
		if len(x.Repeated) > 0 {
			l = 0
			for _, e := range x.Repeated {
				l += runtime.Sov(uint64(e))
			}
			n += 2 + runtime.Sov(uint64(l)) + l
		}
		if len(x.Map) > 0 {
			SiZeMaP := func(k string, v uint32) {
				mapEntrySize := 1 + len(k) + runtime.Sov(uint64(len(k))) + 1 + runtime.Sov(uint64(v))
				n += mapEntrySize + 2 + runtime.Sov(uint64(mapEntrySize))
			}
			if options.Deterministic {
				sortme := make([]string, 0, len(x.Map))
				for k := range x.Map {
					sortme = append(sortme, k)
				}
				sort.Strings(sortme)
				for _, k := range sortme {
					v := x.Map[k]
					SiZeMaP(k, v)
				}
			} else {
				for k, v := range x.Map {
					SiZeMaP(k, v)
				}
			}
		}
		if x.Msg != nil {
			l = options.Size(x.Msg)
			n += 2 + l + runtime.Sov(uint64(l))
		}
		switch x := x.Sum.(type) {
		case *A_Oneof:
			if x == nil {
				break
			}
			n += 2 + runtime.Sov(uint64(x.Oneof))
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
		x := input.Message.Interface().(*A)
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
		switch x := x.Sum.(type) {
		case *A_Oneof:
			i = runtime.EncodeVarint(dAtA, i, uint64(x.Oneof))
			i--
			dAtA[i] = 0x1
			i--
			dAtA[i] = 0xa0
		}
		if x.Msg != nil {
			encoded, err := options.Marshal(x.Msg)
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
			dAtA[i] = 0x1
			i--
			dAtA[i] = 0x9a
		}
		if len(x.Map) > 0 {
			MaRsHaLmAp := func(k string, v uint32) (protoiface.MarshalOutput, error) {
				baseI := i
				i = runtime.EncodeVarint(dAtA, i, uint64(v))
				i--
				dAtA[i] = 0x10
				i -= len(k)
				copy(dAtA[i:], k)
				i = runtime.EncodeVarint(dAtA, i, uint64(len(k)))
				i--
				dAtA[i] = 0xa
				i = runtime.EncodeVarint(dAtA, i, uint64(baseI-i))
				i--
				dAtA[i] = 0x1
				i--
				dAtA[i] = 0x92
				return protoiface.MarshalOutput{}, nil
			}
			if options.Deterministic {
				keysForMap := make([]string, 0, len(x.Map))
				for k := range x.Map {
					keysForMap = append(keysForMap, string(k))
				}
				sort.Slice(keysForMap, func(i, j int) bool {
					return keysForMap[i] < keysForMap[j]
				})
				for iNdEx := len(keysForMap) - 1; iNdEx >= 0; iNdEx-- {
					v := x.Map[string(keysForMap[iNdEx])]
					out, err := MaRsHaLmAp(keysForMap[iNdEx], v)
					if err != nil {
						return out, err
					}
				}
			} else {
				for k := range x.Map {
					v := x.Map[k]
					out, err := MaRsHaLmAp(k, v)
					if err != nil {
						return out, err
					}
				}
			}
		}
		if len(x.Repeated) > 0 {
			var pksize2 int
			for _, num := range x.Repeated {
				pksize2 += runtime.Sov(uint64(num))
			}
			i -= pksize2
			j1 := i
			for _, num := range x.Repeated {
				for num >= 1<<7 {
					dAtA[j1] = uint8(uint64(num)&0x7f | 0x80)
					num >>= 7
					j1++
				}
				dAtA[j1] = uint8(num)
				j1++
			}
			i = runtime.EncodeVarint(dAtA, i, uint64(pksize2))
			i--
			dAtA[i] = 0x1
			i--
			dAtA[i] = 0x8a
		}
		if x.E != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.E))
			i--
			dAtA[i] = 0x1
			i--
			dAtA[i] = 0x80
		}
		if x.B {
			i--
			if x.B {
				dAtA[i] = 1
			} else {
				dAtA[i] = 0
			}
			i--
			dAtA[i] = 0x78
		}
		if x.F64 != 0 {
			i -= 8
			binary.LittleEndian.PutUint64(dAtA[i:], uint64(x.F64))
			i--
			dAtA[i] = 0x71
		}
		if x.F32 != 0 {
			i -= 4
			binary.LittleEndian.PutUint32(dAtA[i:], uint32(x.F32))
			i--
			dAtA[i] = 0x6d
		}
		if x.Sf64 != 0 {
			i -= 8
			binary.LittleEndian.PutUint64(dAtA[i:], uint64(x.Sf64))
			i--
			dAtA[i] = 0x61
		}
		if x.S64 != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64((uint64(x.S64)<<1)^uint64((x.S64>>63))))
			i--
			dAtA[i] = 0x58
		}
		if x.I64 != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.I64))
			i--
			dAtA[i] = 0x50
		}
		if x.Sf32 != 0 {
			i -= 4
			binary.LittleEndian.PutUint32(dAtA[i:], uint32(x.Sf32))
			i--
			dAtA[i] = 0x4d
		}
		if x.S32 != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64((uint32(x.S32)<<1)^uint32((x.S32>>31))))
			i--
			dAtA[i] = 0x40
		}
		if x.I32 != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.I32))
			i--
			dAtA[i] = 0x38
		}
		if x.Dur != nil {
			encoded, err := options.Marshal(x.Dur)
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
			dAtA[i] = 0x32
		}
		if x.Ts != nil {
			encoded, err := options.Marshal(x.Ts)
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
		if len(x.Bz) > 0 {
			i -= len(x.Bz)
			copy(dAtA[i:], x.Bz)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Bz)))
			i--
			dAtA[i] = 0x22
		}
		if len(x.Str) > 0 {
			i -= len(x.Str)
			copy(dAtA[i:], x.Str)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.Str)))
			i--
			dAtA[i] = 0x1a
		}
		if x.U64 != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.U64))
			i--
			dAtA[i] = 0x10
		}
		if x.U32 != 0 {
			i = runtime.EncodeVarint(dAtA, i, uint64(x.U32))
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
		x := input.Message.Interface().(*A)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: A: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: A: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field U32", wireType)
				}
				x.U32 = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.U32 |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 2:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field U64", wireType)
				}
				x.U64 = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.U64 |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 3:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Str", wireType)
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
				x.Str = string(dAtA[iNdEx:postIndex])
				iNdEx = postIndex
			case 4:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Bz", wireType)
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
				x.Bz = append(x.Bz[:0], dAtA[iNdEx:postIndex]...)
				if x.Bz == nil {
					x.Bz = []byte{}
				}
				iNdEx = postIndex
			case 5:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Ts", wireType)
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
				if x.Ts == nil {
					x.Ts = &timestamppb.Timestamp{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Ts); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 6:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Dur", wireType)
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
				if x.Dur == nil {
					x.Dur = &durationpb.Duration{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Dur); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 7:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field I32", wireType)
				}
				x.I32 = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.I32 |= int32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 8:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field S32", wireType)
				}
				var v int32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= int32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				v = int32((uint32(v) >> 1) ^ uint32(((v&1)<<31)>>31))
				x.S32 = v
			case 9:
				if wireType != 5 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Sf32", wireType)
				}
				x.Sf32 = 0
				if (iNdEx + 4) > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Sf32 = int32(binary.LittleEndian.Uint32(dAtA[iNdEx:]))
				iNdEx += 4
			case 10:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field I64", wireType)
				}
				x.I64 = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.I64 |= int64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 11:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field S64", wireType)
				}
				var v uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				v = (v >> 1) ^ uint64((int64(v&1)<<63)>>63)
				x.S64 = int64(v)
			case 12:
				if wireType != 1 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Sf64", wireType)
				}
				x.Sf64 = 0
				if (iNdEx + 8) > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.Sf64 = int64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
				iNdEx += 8
			case 13:
				if wireType != 5 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field F32", wireType)
				}
				x.F32 = 0
				if (iNdEx + 4) > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.F32 = uint32(binary.LittleEndian.Uint32(dAtA[iNdEx:]))
				iNdEx += 4
			case 14:
				if wireType != 1 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field F64", wireType)
				}
				x.F64 = 0
				if (iNdEx + 8) > l {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
				}
				x.F64 = uint64(binary.LittleEndian.Uint64(dAtA[iNdEx:]))
				iNdEx += 8
			case 15:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field B", wireType)
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
				x.B = bool(v != 0)
			case 16:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field E", wireType)
				}
				x.E = 0
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					x.E |= Enum(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 17:
				if wireType == 0 {
					var v uint32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
						}
						if iNdEx >= l {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= uint32(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					x.Repeated = append(x.Repeated, v)
				} else if wireType == 2 {
					var packedLen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
						}
						if iNdEx >= l {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						packedLen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if packedLen < 0 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
					}
					postIndex := iNdEx + packedLen
					if postIndex < 0 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
					}
					if postIndex > l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					var elementCount int
					var count int
					for _, integer := range dAtA[iNdEx:postIndex] {
						if integer < 128 {
							count++
						}
					}
					elementCount = count
					if elementCount != 0 && len(x.Repeated) == 0 {
						x.Repeated = make([]uint32, 0, elementCount)
					}
					for iNdEx < postIndex {
						var v uint32
						for shift := uint(0); ; shift += 7 {
							if shift >= 64 {
								return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
							}
							if iNdEx >= l {
								return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
							}
							b := dAtA[iNdEx]
							iNdEx++
							v |= uint32(b&0x7F) << shift
							if b < 0x80 {
								break
							}
						}
						x.Repeated = append(x.Repeated, v)
					}
				} else {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Repeated", wireType)
				}
			case 18:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Map", wireType)
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
				if x.Map == nil {
					x.Map = make(map[string]uint32)
				}
				var mapkey string
				var mapvalue uint32
				for iNdEx < postIndex {
					entryPreIndex := iNdEx
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
					if fieldNum == 1 {
						var stringLenmapkey uint64
						for shift := uint(0); ; shift += 7 {
							if shift >= 64 {
								return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
							}
							if iNdEx >= l {
								return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
							}
							b := dAtA[iNdEx]
							iNdEx++
							stringLenmapkey |= uint64(b&0x7F) << shift
							if b < 0x80 {
								break
							}
						}
						intStringLenmapkey := int(stringLenmapkey)
						if intStringLenmapkey < 0 {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
						}
						postStringIndexmapkey := iNdEx + intStringLenmapkey
						if postStringIndexmapkey < 0 {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
						}
						if postStringIndexmapkey > l {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
						}
						mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
						iNdEx = postStringIndexmapkey
					} else if fieldNum == 2 {
						for shift := uint(0); ; shift += 7 {
							if shift >= 64 {
								return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
							}
							if iNdEx >= l {
								return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
							}
							b := dAtA[iNdEx]
							iNdEx++
							mapvalue |= uint32(b&0x7F) << shift
							if b < 0x80 {
								break
							}
						}
					} else {
						iNdEx = entryPreIndex
						skippy, err := runtime.Skip(dAtA[iNdEx:])
						if err != nil {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
						}
						if (skippy < 0) || (iNdEx+skippy) < 0 {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrInvalidLength
						}
						if (iNdEx + skippy) > postIndex {
							return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
						}
						iNdEx += skippy
					}
				}
				x.Map[mapkey] = mapvalue
				iNdEx = postIndex
			case 19:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Msg", wireType)
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
				if x.Msg == nil {
					x.Msg = &B{}
				}
				if err := options.Unmarshal(dAtA[iNdEx:postIndex], x.Msg); err != nil {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, err
				}
				iNdEx = postIndex
			case 20:
				if wireType != 0 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field Oneof", wireType)
				}
				var v uint32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, runtime.ErrIntOverflow
					}
					if iNdEx >= l {
						return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				x.Sum = &A_Oneof{v}
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
	md_B   protoreflect.MessageDescriptor
	fd_B_x protoreflect.FieldDescriptor
)

func init() {
	file_testpb_testschema_proto_init()
	md_B = File_testpb_testschema_proto.Messages().ByName("B")
	fd_B_x = md_B.Fields().ByName("x")
}

var _ protoreflect.Message = (*fastReflection_B)(nil)

type fastReflection_B B

func (x *B) ProtoReflect() protoreflect.Message {
	return (*fastReflection_B)(x)
}

func (x *B) slowProtoReflect() protoreflect.Message {
	mi := &file_testpb_testschema_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_B_messageType fastReflection_B_messageType
var _ protoreflect.MessageType = fastReflection_B_messageType{}

type fastReflection_B_messageType struct{}

func (x fastReflection_B_messageType) Zero() protoreflect.Message {
	return (*fastReflection_B)(nil)
}
func (x fastReflection_B_messageType) New() protoreflect.Message {
	return new(fastReflection_B)
}
func (x fastReflection_B_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_B
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_B) Descriptor() protoreflect.MessageDescriptor {
	return md_B
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_B) Type() protoreflect.MessageType {
	return _fastReflection_B_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_B) New() protoreflect.Message {
	return new(fastReflection_B)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_B) Interface() protoreflect.ProtoMessage {
	return (*B)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_B) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.X != "" {
		value := protoreflect.ValueOfString(x.X)
		if !f(fd_B_x, value) {
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
func (x *fastReflection_B) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "testpb.B.x":
		return x.X != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.B"))
		}
		panic(fmt.Errorf("message testpb.B does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_B) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "testpb.B.x":
		x.X = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.B"))
		}
		panic(fmt.Errorf("message testpb.B does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_B) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "testpb.B.x":
		value := x.X
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.B"))
		}
		panic(fmt.Errorf("message testpb.B does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_B) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "testpb.B.x":
		x.X = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.B"))
		}
		panic(fmt.Errorf("message testpb.B does not contain field %s", fd.FullName()))
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
func (x *fastReflection_B) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "testpb.B.x":
		panic(fmt.Errorf("field x of message testpb.B is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.B"))
		}
		panic(fmt.Errorf("message testpb.B does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_B) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "testpb.B.x":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.B"))
		}
		panic(fmt.Errorf("message testpb.B does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_B) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in testpb.B", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_B) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_B) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_B) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_B) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*B)
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
		l = len(x.X)
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
		x := input.Message.Interface().(*B)
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
		if len(x.X) > 0 {
			i -= len(x.X)
			copy(dAtA[i:], x.X)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.X)))
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
		x := input.Message.Interface().(*B)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: B: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: B: illegal tag %d (wire type %d)", fieldNum, wire)
			}
			switch fieldNum {
			case 1:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field X", wireType)
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
				x.X = string(dAtA[iNdEx:postIndex])
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
	md_C    protoreflect.MessageDescriptor
	fd_C_id protoreflect.FieldDescriptor
	fd_C_x  protoreflect.FieldDescriptor
)

func init() {
	file_testpb_testschema_proto_init()
	md_C = File_testpb_testschema_proto.Messages().ByName("C")
	fd_C_id = md_C.Fields().ByName("id")
	fd_C_x = md_C.Fields().ByName("x")
}

var _ protoreflect.Message = (*fastReflection_C)(nil)

type fastReflection_C C

func (x *C) ProtoReflect() protoreflect.Message {
	return (*fastReflection_C)(x)
}

func (x *C) slowProtoReflect() protoreflect.Message {
	mi := &file_testpb_testschema_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

var _fastReflection_C_messageType fastReflection_C_messageType
var _ protoreflect.MessageType = fastReflection_C_messageType{}

type fastReflection_C_messageType struct{}

func (x fastReflection_C_messageType) Zero() protoreflect.Message {
	return (*fastReflection_C)(nil)
}
func (x fastReflection_C_messageType) New() protoreflect.Message {
	return new(fastReflection_C)
}
func (x fastReflection_C_messageType) Descriptor() protoreflect.MessageDescriptor {
	return md_C
}

// Descriptor returns message descriptor, which contains only the protobuf
// type information for the message.
func (x *fastReflection_C) Descriptor() protoreflect.MessageDescriptor {
	return md_C
}

// Type returns the message type, which encapsulates both Go and protobuf
// type information. If the Go type information is not needed,
// it is recommended that the message descriptor be used instead.
func (x *fastReflection_C) Type() protoreflect.MessageType {
	return _fastReflection_C_messageType
}

// New returns a newly allocated and mutable empty message.
func (x *fastReflection_C) New() protoreflect.Message {
	return new(fastReflection_C)
}

// Interface unwraps the message reflection interface and
// returns the underlying ProtoMessage interface.
func (x *fastReflection_C) Interface() protoreflect.ProtoMessage {
	return (*C)(x)
}

// Range iterates over every populated field in an undefined order,
// calling f for each field descriptor and value encountered.
// Range returns immediately if f returns false.
// While iterating, mutating operations may only be performed
// on the current field descriptor.
func (x *fastReflection_C) Range(f func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	if x.Id != uint64(0) {
		value := protoreflect.ValueOfUint64(x.Id)
		if !f(fd_C_id, value) {
			return
		}
	}
	if x.X != "" {
		value := protoreflect.ValueOfString(x.X)
		if !f(fd_C_x, value) {
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
func (x *fastReflection_C) Has(fd protoreflect.FieldDescriptor) bool {
	switch fd.FullName() {
	case "testpb.C.id":
		return x.Id != uint64(0)
	case "testpb.C.x":
		return x.X != ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.C"))
		}
		panic(fmt.Errorf("message testpb.C does not contain field %s", fd.FullName()))
	}
}

// Clear clears the field such that a subsequent Has call reports false.
//
// Clearing an extension field clears both the extension type and value
// associated with the given field number.
//
// Clear is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_C) Clear(fd protoreflect.FieldDescriptor) {
	switch fd.FullName() {
	case "testpb.C.id":
		x.Id = uint64(0)
	case "testpb.C.x":
		x.X = ""
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.C"))
		}
		panic(fmt.Errorf("message testpb.C does not contain field %s", fd.FullName()))
	}
}

// Get retrieves the value for a field.
//
// For unpopulated scalars, it returns the default value, where
// the default value of a bytes scalar is guaranteed to be a copy.
// For unpopulated composite types, it returns an empty, read-only view
// of the value; to obtain a mutable reference, use Mutable.
func (x *fastReflection_C) Get(descriptor protoreflect.FieldDescriptor) protoreflect.Value {
	switch descriptor.FullName() {
	case "testpb.C.id":
		value := x.Id
		return protoreflect.ValueOfUint64(value)
	case "testpb.C.x":
		value := x.X
		return protoreflect.ValueOfString(value)
	default:
		if descriptor.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.C"))
		}
		panic(fmt.Errorf("message testpb.C does not contain field %s", descriptor.FullName()))
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
func (x *fastReflection_C) Set(fd protoreflect.FieldDescriptor, value protoreflect.Value) {
	switch fd.FullName() {
	case "testpb.C.id":
		x.Id = value.Uint()
	case "testpb.C.x":
		x.X = value.Interface().(string)
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.C"))
		}
		panic(fmt.Errorf("message testpb.C does not contain field %s", fd.FullName()))
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
func (x *fastReflection_C) Mutable(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "testpb.C.id":
		panic(fmt.Errorf("field id of message testpb.C is not mutable"))
	case "testpb.C.x":
		panic(fmt.Errorf("field x of message testpb.C is not mutable"))
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.C"))
		}
		panic(fmt.Errorf("message testpb.C does not contain field %s", fd.FullName()))
	}
}

// NewField returns a new value that is assignable to the field
// for the given descriptor. For scalars, this returns the default value.
// For lists, maps, and messages, this returns a new, empty, mutable value.
func (x *fastReflection_C) NewField(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.FullName() {
	case "testpb.C.id":
		return protoreflect.ValueOfUint64(uint64(0))
	case "testpb.C.x":
		return protoreflect.ValueOfString("")
	default:
		if fd.IsExtension() {
			panic(fmt.Errorf("proto3 declared messages do not support extensions: testpb.C"))
		}
		panic(fmt.Errorf("message testpb.C does not contain field %s", fd.FullName()))
	}
}

// WhichOneof reports which field within the oneof is populated,
// returning nil if none are populated.
// It panics if the oneof descriptor does not belong to this message.
func (x *fastReflection_C) WhichOneof(d protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	switch d.FullName() {
	default:
		panic(fmt.Errorf("%s is not a oneof field in testpb.C", d.FullName()))
	}
	panic("unreachable")
}

// GetUnknown retrieves the entire list of unknown fields.
// The caller may only mutate the contents of the RawFields
// if the mutated bytes are stored back into the message with SetUnknown.
func (x *fastReflection_C) GetUnknown() protoreflect.RawFields {
	return x.unknownFields
}

// SetUnknown stores an entire list of unknown fields.
// The raw fields must be syntactically valid according to the wire format.
// An implementation may panic if this is not the case.
// Once stored, the caller must not mutate the content of the RawFields.
// An empty RawFields may be passed to clear the fields.
//
// SetUnknown is a mutating operation and unsafe for concurrent use.
func (x *fastReflection_C) SetUnknown(fields protoreflect.RawFields) {
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
func (x *fastReflection_C) IsValid() bool {
	return x != nil
}

// ProtoMethods returns optional fastReflectionFeature-path implementations of various operations.
// This method may return nil.
//
// The returned methods type is identical to
// "google.golang.org/protobuf/runtime/protoiface".Methods.
// Consult the protoiface package documentation for details.
func (x *fastReflection_C) ProtoMethods() *protoiface.Methods {
	size := func(input protoiface.SizeInput) protoiface.SizeOutput {
		x := input.Message.Interface().(*C)
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
		l = len(x.X)
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
		x := input.Message.Interface().(*C)
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
		if len(x.X) > 0 {
			i -= len(x.X)
			copy(dAtA[i:], x.X)
			i = runtime.EncodeVarint(dAtA, i, uint64(len(x.X)))
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
		x := input.Message.Interface().(*C)
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
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: C: wiretype end group for non-group")
			}
			if fieldNum <= 0 {
				return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: C: illegal tag %d (wire type %d)", fieldNum, wire)
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
					x.Id |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
			case 2:
				if wireType != 2 {
					return protoiface.UnmarshalOutput{NoUnkeyedLiterals: input.NoUnkeyedLiterals, Flags: input.Flags}, fmt.Errorf("proto: wrong wireType = %d for field X", wireType)
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
				x.X = string(dAtA[iNdEx:postIndex])
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
// source: testpb/testschema.proto

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Enum int32

const (
	Enum_ENUM_UNSPECIFIED Enum = 0
	Enum_ENUM_ONE         Enum = 1
	Enum_ENUM_TWO         Enum = 2
	Enum_ENUM_FIVE        Enum = 5
	Enum_ENUM_NEG_THREE   Enum = -3
)

// Enum value maps for Enum.
var (
	Enum_name = map[int32]string{
		0:  "ENUM_UNSPECIFIED",
		1:  "ENUM_ONE",
		2:  "ENUM_TWO",
		5:  "ENUM_FIVE",
		-3: "ENUM_NEG_THREE",
	}
	Enum_value = map[string]int32{
		"ENUM_UNSPECIFIED": 0,
		"ENUM_ONE":         1,
		"ENUM_TWO":         2,
		"ENUM_FIVE":        5,
		"ENUM_NEG_THREE":   -3,
	}
)

func (x Enum) Enum() *Enum {
	p := new(Enum)
	*p = x
	return p
}

func (x Enum) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Enum) Descriptor() protoreflect.EnumDescriptor {
	return file_testpb_testschema_proto_enumTypes[0].Descriptor()
}

func (Enum) Type() protoreflect.EnumType {
	return &file_testpb_testschema_proto_enumTypes[0]
}

func (x Enum) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Enum.Descriptor instead.
func (Enum) EnumDescriptor() ([]byte, []int) {
	return file_testpb_testschema_proto_rawDescGZIP(), []int{0}
}

type A struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Valid key fields:
	U32  uint32                 `protobuf:"varint,1,opt,name=u32,proto3" json:"u32,omitempty"`
	U64  uint64                 `protobuf:"varint,2,opt,name=u64,proto3" json:"u64,omitempty"`
	Str  string                 `protobuf:"bytes,3,opt,name=str,proto3" json:"str,omitempty"`
	Bz   []byte                 `protobuf:"bytes,4,opt,name=bz,proto3" json:"bz,omitempty"`
	Ts   *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=ts,proto3" json:"ts,omitempty"`
	Dur  *durationpb.Duration   `protobuf:"bytes,6,opt,name=dur,proto3" json:"dur,omitempty"`
	I32  int32                  `protobuf:"varint,7,opt,name=i32,proto3" json:"i32,omitempty"`
	S32  int32                  `protobuf:"zigzag32,8,opt,name=s32,proto3" json:"s32,omitempty"`
	Sf32 int32                  `protobuf:"fixed32,9,opt,name=sf32,proto3" json:"sf32,omitempty"`
	I64  int64                  `protobuf:"varint,10,opt,name=i64,proto3" json:"i64,omitempty"`
	S64  int64                  `protobuf:"zigzag64,11,opt,name=s64,proto3" json:"s64,omitempty"`
	Sf64 int64                  `protobuf:"fixed64,12,opt,name=sf64,proto3" json:"sf64,omitempty"`
	F32  uint32                 `protobuf:"fixed32,13,opt,name=f32,proto3" json:"f32,omitempty"`
	F64  uint64                 `protobuf:"fixed64,14,opt,name=f64,proto3" json:"f64,omitempty"`
	B    bool                   `protobuf:"varint,15,opt,name=b,proto3" json:"b,omitempty"`
	E    Enum                   `protobuf:"varint,16,opt,name=e,proto3,enum=testpb.Enum" json:"e,omitempty"`
	// Invalid key fields:
	Repeated []uint32          `protobuf:"varint,17,rep,packed,name=repeated,proto3" json:"repeated,omitempty"`
	Map      map[string]uint32 `protobuf:"bytes,18,rep,name=map,proto3" json:"map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
	Msg      *B                `protobuf:"bytes,19,opt,name=msg,proto3" json:"msg,omitempty"`
	// Types that are assignable to Sum:
	//	*A_Oneof
	Sum isA_Sum `protobuf_oneof:"sum"`
}

func (x *A) Reset() {
	*x = A{}
	if protoimpl.UnsafeEnabled {
		mi := &file_testpb_testschema_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *A) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*A) ProtoMessage() {}

// Deprecated: Use A.ProtoReflect.Descriptor instead.
func (*A) Descriptor() ([]byte, []int) {
	return file_testpb_testschema_proto_rawDescGZIP(), []int{0}
}

func (x *A) GetU32() uint32 {
	if x != nil {
		return x.U32
	}
	return 0
}

func (x *A) GetU64() uint64 {
	if x != nil {
		return x.U64
	}
	return 0
}

func (x *A) GetStr() string {
	if x != nil {
		return x.Str
	}
	return ""
}

func (x *A) GetBz() []byte {
	if x != nil {
		return x.Bz
	}
	return nil
}

func (x *A) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

func (x *A) GetDur() *durationpb.Duration {
	if x != nil {
		return x.Dur
	}
	return nil
}

func (x *A) GetI32() int32 {
	if x != nil {
		return x.I32
	}
	return 0
}

func (x *A) GetS32() int32 {
	if x != nil {
		return x.S32
	}
	return 0
}

func (x *A) GetSf32() int32 {
	if x != nil {
		return x.Sf32
	}
	return 0
}

func (x *A) GetI64() int64 {
	if x != nil {
		return x.I64
	}
	return 0
}

func (x *A) GetS64() int64 {
	if x != nil {
		return x.S64
	}
	return 0
}

func (x *A) GetSf64() int64 {
	if x != nil {
		return x.Sf64
	}
	return 0
}

func (x *A) GetF32() uint32 {
	if x != nil {
		return x.F32
	}
	return 0
}

func (x *A) GetF64() uint64 {
	if x != nil {
		return x.F64
	}
	return 0
}

func (x *A) GetB() bool {
	if x != nil {
		return x.B
	}
	return false
}

func (x *A) GetE() Enum {
	if x != nil {
		return x.E
	}
	return Enum_ENUM_UNSPECIFIED
}

func (x *A) GetRepeated() []uint32 {
	if x != nil {
		return x.Repeated
	}
	return nil
}

func (x *A) GetMap() map[string]uint32 {
	if x != nil {
		return x.Map
	}
	return nil
}

func (x *A) GetMsg() *B {
	if x != nil {
		return x.Msg
	}
	return nil
}

func (x *A) GetSum() isA_Sum {
	if x != nil {
		return x.Sum
	}
	return nil
}

func (x *A) GetOneof() uint32 {
	if x, ok := x.GetSum().(*A_Oneof); ok {
		return x.Oneof
	}
	return 0
}

type isA_Sum interface {
	isA_Sum()
}

type A_Oneof struct {
	Oneof uint32 `protobuf:"varint,20,opt,name=oneof,proto3,oneof"`
}

func (*A_Oneof) isA_Sum() {}

type B struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	X string `protobuf:"bytes,1,opt,name=x,proto3" json:"x,omitempty"`
}

func (x *B) Reset() {
	*x = B{}
	if protoimpl.UnsafeEnabled {
		mi := &file_testpb_testschema_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *B) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*B) ProtoMessage() {}

// Deprecated: Use B.ProtoReflect.Descriptor instead.
func (*B) Descriptor() ([]byte, []int) {
	return file_testpb_testschema_proto_rawDescGZIP(), []int{1}
}

func (x *B) GetX() string {
	if x != nil {
		return x.X
	}
	return ""
}

type C struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	X  string `protobuf:"bytes,2,opt,name=x,proto3" json:"x,omitempty"`
}

func (x *C) Reset() {
	*x = C{}
	if protoimpl.UnsafeEnabled {
		mi := &file_testpb_testschema_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *C) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*C) ProtoMessage() {}

// Deprecated: Use C.ProtoReflect.Descriptor instead.
func (*C) Descriptor() ([]byte, []int) {
	return file_testpb_testschema_proto_rawDescGZIP(), []int{2}
}

func (x *C) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *C) GetX() string {
	if x != nil {
		return x.X
	}
	return ""
}

var File_testpb_testschema_proto protoreflect.FileDescriptor

var file_testpb_testschema_proto_rawDesc = []byte{
	0x0a, 0x17, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x63, 0x68,
	0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x74, 0x65, 0x73, 0x74, 0x70,
	0x62, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x1d, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x6f, 0x72, 0x6d, 0x2f, 0x76,
	0x31, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x6f, 0x72, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xd5, 0x04, 0x0a, 0x01, 0x41, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x33, 0x32, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x03, 0x75, 0x33, 0x32, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x36, 0x34,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x75, 0x36, 0x34, 0x12, 0x10, 0x0a, 0x03, 0x73,
	0x74, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x74, 0x72, 0x12, 0x0e, 0x0a,
	0x02, 0x62, 0x7a, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x62, 0x7a, 0x12, 0x2a, 0x0a,
	0x02, 0x74, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02, 0x74, 0x73, 0x12, 0x2b, 0x0a, 0x03, 0x64, 0x75, 0x72,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x03, 0x64, 0x75, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x69, 0x33, 0x32, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x03, 0x69, 0x33, 0x32, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x33, 0x32, 0x18,
	0x08, 0x20, 0x01, 0x28, 0x11, 0x52, 0x03, 0x73, 0x33, 0x32, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x66,
	0x33, 0x32, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0f, 0x52, 0x04, 0x73, 0x66, 0x33, 0x32, 0x12, 0x10,
	0x0a, 0x03, 0x69, 0x36, 0x34, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x69, 0x36, 0x34,
	0x12, 0x10, 0x0a, 0x03, 0x73, 0x36, 0x34, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x12, 0x52, 0x03, 0x73,
	0x36, 0x34, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x66, 0x36, 0x34, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x10,
	0x52, 0x04, 0x73, 0x66, 0x36, 0x34, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x33, 0x32, 0x18, 0x0d, 0x20,
	0x01, 0x28, 0x07, 0x52, 0x03, 0x66, 0x33, 0x32, 0x12, 0x10, 0x0a, 0x03, 0x66, 0x36, 0x34, 0x18,
	0x0e, 0x20, 0x01, 0x28, 0x06, 0x52, 0x03, 0x66, 0x36, 0x34, 0x12, 0x0c, 0x0a, 0x01, 0x62, 0x18,
	0x0f, 0x20, 0x01, 0x28, 0x08, 0x52, 0x01, 0x62, 0x12, 0x1a, 0x0a, 0x01, 0x65, 0x18, 0x10, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x45, 0x6e, 0x75,
	0x6d, 0x52, 0x01, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x18, 0x11, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x08, 0x72, 0x65, 0x70, 0x65, 0x61, 0x74, 0x65, 0x64,
	0x12, 0x24, 0x0a, 0x03, 0x6d, 0x61, 0x70, 0x18, 0x12, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e,
	0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x41, 0x2e, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72,
	0x79, 0x52, 0x03, 0x6d, 0x61, 0x70, 0x12, 0x1b, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x13, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x09, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62, 0x2e, 0x42, 0x52, 0x03,
	0x6d, 0x73, 0x67, 0x12, 0x16, 0x0a, 0x05, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x18, 0x14, 0x20, 0x01,
	0x28, 0x0d, 0x48, 0x00, 0x52, 0x05, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x1a, 0x36, 0x0a, 0x08, 0x4d,
	0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a,
	0x02, 0x38, 0x01, 0x3a, 0x3d, 0xf2, 0x9e, 0xd3, 0x8e, 0x03, 0x37, 0x0a, 0x0d, 0x0a, 0x0b, 0x75,
	0x33, 0x32, 0x2c, 0x75, 0x36, 0x34, 0x2c, 0x73, 0x74, 0x72, 0x12, 0x0b, 0x0a, 0x07, 0x75, 0x36,
	0x34, 0x2c, 0x73, 0x74, 0x72, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x73, 0x74, 0x72, 0x2c, 0x75,
	0x33, 0x32, 0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x62, 0x7a, 0x2c, 0x73, 0x74, 0x72, 0x10, 0x03,
	0x18, 0x01, 0x42, 0x05, 0x0a, 0x03, 0x73, 0x75, 0x6d, 0x22, 0x1b, 0x0a, 0x01, 0x42, 0x12, 0x0c,
	0x0a, 0x01, 0x78, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x01, 0x78, 0x3a, 0x08, 0xfa, 0x9e,
	0xd3, 0x8e, 0x03, 0x02, 0x08, 0x02, 0x22, 0x33, 0x0a, 0x01, 0x43, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x12, 0x0c, 0x0a, 0x01, 0x78,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x01, 0x78, 0x3a, 0x10, 0xf2, 0x9e, 0xd3, 0x8e, 0x03,
	0x0a, 0x0a, 0x06, 0x0a, 0x02, 0x69, 0x64, 0x10, 0x01, 0x18, 0x03, 0x2a, 0x64, 0x0a, 0x04, 0x45,
	0x6e, 0x75, 0x6d, 0x12, 0x14, 0x0a, 0x10, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x55, 0x4e, 0x53, 0x50,
	0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x45, 0x4e, 0x55,
	0x4d, 0x5f, 0x4f, 0x4e, 0x45, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x45, 0x4e, 0x55, 0x4d, 0x5f,
	0x54, 0x57, 0x4f, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x46, 0x49,
	0x56, 0x45, 0x10, 0x05, 0x12, 0x1b, 0x0a, 0x0e, 0x45, 0x4e, 0x55, 0x4d, 0x5f, 0x4e, 0x45, 0x47,
	0x5f, 0x54, 0x48, 0x52, 0x45, 0x45, 0x10, 0xfd, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0x01, 0x42, 0x87, 0x01, 0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x70, 0x62,
	0x42, 0x0f, 0x54, 0x65, 0x73, 0x74, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x50, 0x01, 0x5a, 0x30, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2f, 0x63, 0x6f, 0x73, 0x6d, 0x6f, 0x73, 0x2d, 0x73, 0x64,
	0x6b, 0x2f, 0x6f, 0x72, 0x6d, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x74,
	0x65, 0x73, 0x74, 0x70, 0x62, 0xa2, 0x02, 0x03, 0x54, 0x58, 0x58, 0xaa, 0x02, 0x06, 0x54, 0x65,
	0x73, 0x74, 0x70, 0x62, 0xca, 0x02, 0x06, 0x54, 0x65, 0x73, 0x74, 0x70, 0x62, 0xe2, 0x02, 0x12,
	0x54, 0x65, 0x73, 0x74, 0x70, 0x62, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x06, 0x54, 0x65, 0x73, 0x74, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_testpb_testschema_proto_rawDescOnce sync.Once
	file_testpb_testschema_proto_rawDescData = file_testpb_testschema_proto_rawDesc
)

func file_testpb_testschema_proto_rawDescGZIP() []byte {
	file_testpb_testschema_proto_rawDescOnce.Do(func() {
		file_testpb_testschema_proto_rawDescData = protoimpl.X.CompressGZIP(file_testpb_testschema_proto_rawDescData)
	})
	return file_testpb_testschema_proto_rawDescData
}

var file_testpb_testschema_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_testpb_testschema_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_testpb_testschema_proto_goTypes = []interface{}{
	(Enum)(0),                     // 0: testpb.Enum
	(*A)(nil),                     // 1: testpb.A
	(*B)(nil),                     // 2: testpb.B
	(*C)(nil),                     // 3: testpb.C
	nil,                           // 4: testpb.A.MapEntry
	(*timestamppb.Timestamp)(nil), // 5: google.protobuf.Timestamp
	(*durationpb.Duration)(nil),   // 6: google.protobuf.Duration
}
var file_testpb_testschema_proto_depIdxs = []int32{
	5, // 0: testpb.A.ts:type_name -> google.protobuf.Timestamp
	6, // 1: testpb.A.dur:type_name -> google.protobuf.Duration
	0, // 2: testpb.A.e:type_name -> testpb.Enum
	4, // 3: testpb.A.map:type_name -> testpb.A.MapEntry
	2, // 4: testpb.A.msg:type_name -> testpb.B
	5, // [5:5] is the sub-list for method output_type
	5, // [5:5] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_testpb_testschema_proto_init() }
func file_testpb_testschema_proto_init() {
	if File_testpb_testschema_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_testpb_testschema_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*A); i {
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
		file_testpb_testschema_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*B); i {
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
		file_testpb_testschema_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*C); i {
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
	file_testpb_testschema_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*A_Oneof)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_testpb_testschema_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_testpb_testschema_proto_goTypes,
		DependencyIndexes: file_testpb_testschema_proto_depIdxs,
		EnumInfos:         file_testpb_testschema_proto_enumTypes,
		MessageInfos:      file_testpb_testschema_proto_msgTypes,
	}.Build()
	File_testpb_testschema_proto = out.File
	file_testpb_testschema_proto_rawDesc = nil
	file_testpb_testschema_proto_goTypes = nil
	file_testpb_testschema_proto_depIdxs = nil
}
