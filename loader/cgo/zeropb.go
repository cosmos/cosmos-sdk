package cgo

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const BUFFER_SIZE = 65536
const MAX_EXTENT = BUFFER_SIZE - 2

func ZeroPBMarshal(message proto.Message) ([]byte, error) {
	zd, err := globalRegistry.messageDescriptor(message.ProtoReflect().Descriptor())
	if err != nil {
		return nil, err
	}

	out := newBufferContext()
	if err := zd.marshal(message.ProtoReflect(), out); err != nil {
		return nil, err
	}

	return out.root, nil

}

func ZeroPBUnmarshal(in []byte, message proto.Message) error {
	zd, err := globalRegistry.messageDescriptor(message.ProtoReflect().Descriptor())
	if err != nil {
		return err
	}

	ctx := bufferContext{root: in}

	return zd.unmarshal(ctx, message.ProtoReflect())
}

var globalRegistry = &registry{}

type registry struct {
	messages map[protoreflect.MessageDescriptor]*zeropbMessageDescriptor
}

type zeropbMessageDescriptor struct {
	protoreflect.MessageDescriptor
	registry *registry

	offsets map[protoreflect.FieldNumber]int
	size    int
	align   int
}

func (r *registry) messageDescriptor(md protoreflect.MessageDescriptor) (*zeropbMessageDescriptor, error) {
	if r.messages == nil {
		r.messages = map[protoreflect.MessageDescriptor]*zeropbMessageDescriptor{}
	}
	if zd, ok := r.messages[md]; ok {
		return zd, nil
	}

	// collect and sort fields
	fds := md.Fields()
	numFields := fds.Len()
	fields := make([]protoreflect.FieldDescriptor, numFields)
	for i := 0; i < numFields; i++ {
		fields[i] = fds.Get(i)
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Number() < fields[j].Number()
	})

	offsets := map[protoreflect.FieldNumber]int{}

	align := 1
	lastNum := protoreflect.FieldNumber(0)
	offset := 0
	for _, field := range fields {
		num := field.Number()
		if num != lastNum+1 {
			return nil, fmt.Errorf("field numbers must be declared in consecutive order, field number %d is missing", lastNum+1)
		}

		fieldSize, fieldAlign, err := r.fieldSizeAlign(field)
		if err != nil {
			return nil, err
		}

		if fieldAlign > align {
			align = fieldAlign
		}

		offset = nextAlignedOffset(offset, fieldAlign)
		offsets[field.Number()] = offset
		offset += fieldSize
		lastNum = num
	}

	zd := &zeropbMessageDescriptor{
		MessageDescriptor: md,
		registry:          r,
		offsets:           offsets,
		size:              nextAlignedOffset(offset, align),
		align:             align,
	}
	r.messages[md] = zd
	return zd, nil
}

func (r *registry) fieldSizeAlign(field protoreflect.FieldDescriptor) (int, int, error) {
	if field.IsList() {
		return 4, 2, nil
	}

	if field.HasOptionalKeyword() {
		return 0, 0, fmt.Errorf("optional fields are not handled yet")
	}

	switch field.Kind() {
	case protoreflect.BoolKind:
		return 1, 1, nil
	case protoreflect.Int32Kind, protoreflect.Uint32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind, protoreflect.Fixed32Kind:
		return 4, 4, nil
	case protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind, protoreflect.DoubleKind, protoreflect.Fixed64Kind:
		return 8, 8, nil
	case protoreflect.StringKind:
		return 4, 2, nil
	case protoreflect.BytesKind:
		return 4, 2, nil
	case protoreflect.EnumKind:
		return 4, 4, nil
	case protoreflect.MessageKind:
		md, err := r.messageDescriptor(field.Message())
		if err != nil {
			return 0, 0, err
		}

		return md.size, md.align, nil
	default:
		return 0, 0, fmt.Errorf("unhandled field kind: %v", field.Kind())
	}
}

func nextAlignedOffset(offset, align int) int {
	return (offset + align - 1) &^ (align - 1)
}

func (z *zeropbMessageDescriptor) marshal(msg protoreflect.Message, out bufferContext) error {
	var err error
	out, err = out.alloc(z.size, z.align)
	if err != nil {
		return err
	}

	fds := msg.Descriptor().Fields()
	numFields := fds.Len()
	for i := 0; i < numFields; i++ {
		field := fds.Get(i)
		if !msg.Has(field) {
			continue
		}

		if err := z.marshalField(field, msg.Get(field), out.seek(z.offsets[field.Number()])); err != nil {
			return err
		}
	}
	return nil
}

func (z *zeropbMessageDescriptor) marshalField(field protoreflect.FieldDescriptor, get protoreflect.Value, out bufferContext) error {
	switch {
	case field.IsList():
		return z.marshalList(field, get.List(), out)
	case field.Kind() == protoreflect.MessageKind:
		return z.marshalMessage(get.Message(), out)
	default:
		return z.marshalScalar(field, get, out)
	}
}

func (z *zeropbMessageDescriptor) marshalList(field protoreflect.FieldDescriptor, list protoreflect.List, out bufferContext) error {
	elSize, elAlign, err := z.registry.fieldSizeAlign(field)
	if err != nil {
		return err
	}

	n := list.Len()
	if n == 0 {
		return nil
	}

	if n > MAX_EXTENT {
		return fmt.Errorf("list too large, must be less than %d elements, have %d", MAX_EXTENT, n)
	}

	for left := n; left > 0; {
		segmentCount := 256
		if left > segmentCount {
			segmentCount = left
		}

		segmentSize := elSize * segmentCount
		segmentOut, err := out.alloc(segmentSize, elAlign)
		if err != nil {
			return err
		}

		// write first segment pointer
		if left == n {
			binary.LittleEndian.PutUint16(out.bytes(), uint16(segmentOut.offset))
			binary.LittleEndian.PutUint16(out.bytes()[2:], uint16(n))
		}

		err = z.marshalListSegment(field, list, segmentOut, left, left+segmentCount)
		if err != nil {
			return err
		}
		left -= segmentCount
	}
	return nil
}

func (z *zeropbMessageDescriptor) marshalListSegment(field protoreflect.FieldDescriptor, list protoreflect.List, out bufferContext, start, end int) error {
	panic("TODO")
}

func (z *zeropbMessageDescriptor) marshalMessage(message protoreflect.Message, out bufferContext) error {
	zd, err := z.registry.messageDescriptor(message.Descriptor())
	if err != nil {
		return err
	}

	return zd.marshal(message, out)
}

func (z *zeropbMessageDescriptor) marshalScalar(field protoreflect.FieldDescriptor, value protoreflect.Value, out bufferContext) error {
	bytes := out.bytes()
	switch field.Kind() {
	case protoreflect.BoolKind:
		if value.Bool() {
			bytes[0] = 1
		} else {
			bytes[0] = 0
		}
	case protoreflect.Int32Kind, protoreflect.Uint32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		binary.LittleEndian.PutUint32(bytes, uint32(value.Int()))
	case protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		binary.LittleEndian.PutUint64(bytes, uint64(value.Int()))
	case protoreflect.FloatKind:
		binary.LittleEndian.PutUint32(bytes, math.Float32bits(float32(value.Float())))
	case protoreflect.DoubleKind:
		binary.LittleEndian.PutUint64(bytes, math.Float64bits(value.Float()))
	case protoreflect.StringKind:
		return z.marshalBytes([]byte(value.String()), out)
	case protoreflect.BytesKind:
		return z.marshalBytes(value.Bytes(), out)
	case protoreflect.EnumKind:
		binary.LittleEndian.PutUint32(bytes, uint32(value.Enum()))
	default:
		return fmt.Errorf("unhandled field kind: %v", field.Kind())
	}
	return nil
}

func (z *zeropbMessageDescriptor) marshalBytes(bz []byte, out bufferContext) error {
	n := len(bz)
	newOut, err := out.alloc(n, 1)
	if err != nil {
		return err
	}

	copy(newOut.bytes(), bz)
	binary.LittleEndian.PutUint16(out.bytes(), uint16(newOut.offset))
	binary.LittleEndian.PutUint16(out.bytes()[2:], uint16(n))
	return nil
}

func (z *zeropbMessageDescriptor) unmarshal(in bufferContext, msg protoreflect.Message) error {
	in = in.align(z.align)
	if in.offset+z.size > len(in.root) {
		return fmt.Errorf("buffer too small")
	}

	fds := msg.Descriptor().Fields()
	numFields := fds.Len()
	for i := 0; i < numFields; i++ {
		field := fds.Get(i)
		if !msg.Has(field) {
			continue
		}

		if err := z.unmarshalField(msg, field, in.seek(z.offsets[field.Number()]), msg.Mutable(field)); err != nil {
			return err
		}
	}
	return nil
}

func (z *zeropbMessageDescriptor) unmarshalField(msg protoreflect.Message, field protoreflect.FieldDescriptor, in bufferContext, mutable protoreflect.Value) error {
	switch {
	case field.IsList():
		return z.unmarshalList(field, in, mutable)
	case field.Kind() == protoreflect.MessageKind:
		return z.unmarshalMessage(field, in, mutable)
	default:
		val, err := z.unmarshalScalar(field, in)
		if err != nil {
			return err
		}
		msg.Set(field, val)
		return nil
	}
}

func (z *zeropbMessageDescriptor) unmarshalList(field protoreflect.FieldDescriptor, bytes bufferContext, mutable protoreflect.Value) error {
	panic("TODO")
}

func (z *zeropbMessageDescriptor) unmarshalMessage(field protoreflect.FieldDescriptor, bytes bufferContext, mutable protoreflect.Value) error {
	zd, err := z.registry.messageDescriptor(field.Message())
	if err != nil {
		return err
	}

	newMsg := mutable.Message()
	return zd.unmarshal(bytes, newMsg)
}

func (z *zeropbMessageDescriptor) unmarshalScalar(field protoreflect.FieldDescriptor, in bufferContext) (protoreflect.Value, error) {
	bytes := in.bytes()
	switch field.Kind() {
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(bytes[0] != 0), nil
	case protoreflect.Int32Kind, protoreflect.Uint32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfInt32(int32(binary.LittleEndian.Uint32(bytes))), nil
	case protoreflect.Int64Kind, protoreflect.Uint64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfInt64(int64(binary.LittleEndian.Uint64(bytes))), nil
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(math.Float32frombits(binary.LittleEndian.Uint32(bytes))), nil
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(math.Float64frombits(binary.LittleEndian.Uint64(bytes))), nil
	case protoreflect.StringKind:
		bz, err := z.umarshalBytes(in)
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfString(string(bz)), nil
	case protoreflect.BytesKind:
		bz, err := z.umarshalBytes(in)
		if err != nil {
			return protoreflect.Value{}, err
		}

		return protoreflect.ValueOfBytes(bz), nil
	case protoreflect.EnumKind:
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(binary.LittleEndian.Uint32(bytes))), nil
	default:
		return protoreflect.Value{}, fmt.Errorf("unhandled field kind: %v", field.Kind())
	}
}

func (z *zeropbMessageDescriptor) umarshalBytes(in bufferContext) ([]byte, error) {
	offset := int(binary.LittleEndian.Uint16(in.bytes()))
	n := int(binary.LittleEndian.Uint16(in.bytes()[2:]))
	return in.root[offset : offset+n], nil
}

type bufferContext struct {
	root   []byte
	offset int
}

func newBufferContext() bufferContext {
	return bufferContext{root: make([]byte, BUFFER_SIZE)}
}

func (c bufferContext) extent() int {
	return int(binary.LittleEndian.Uint16(c.root[MAX_EXTENT:]))
}

func (c bufferContext) setExtent(extent int) {
	binary.LittleEndian.PutUint16(c.root[MAX_EXTENT:], uint16(extent))
}

func (c bufferContext) alloc(size, align int) (bufferContext, error) {
	offset := nextAlignedOffset(c.offset, align)
	if offset+size > c.extent() {
		return bufferContext{}, fmt.Errorf("out of buffer space")
	}
	c.setExtent(offset + size)
	return bufferContext{root: c.root, offset: offset}, nil
}

func (c bufferContext) seek(offset int) bufferContext {
	c.offset += offset
	return c
}

func (c bufferContext) bytes() []byte {
	return c.root[c.offset:]
}

func (c bufferContext) align(align int) bufferContext {
	c.offset = nextAlignedOffset(c.offset, align)
	return c
}
