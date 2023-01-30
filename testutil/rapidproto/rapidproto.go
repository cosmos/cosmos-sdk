package rapidproto

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"gotest.tools/v3/assert"
	"pgregory.net/rapid"
)

func MessageGenerator[T proto.Message](x T, options GeneratorOptions) *rapid.Generator[T] {
	msgType := x.ProtoReflect().Type()
	return rapid.Custom(func(t *rapid.T) T {
		msg := msgType.New()

		options.setFields(t, msg, 0)

		return msg.Interface().(T)
	})
}

type GeneratorOptions struct {
	AnyTypeURLs []string
	Resolver    protoregistry.MessageTypeResolver
}

const depthLimit = 10

func (opts GeneratorOptions) setFields(t *rapid.T, msg protoreflect.Message, depth int) bool {
	// to avoid stack overflow we limit the depth of nested messages
	if depth > depthLimit {
		return false
	}

	descriptor := msg.Descriptor()
	fullName := descriptor.FullName()
	switch fullName {
	case timestampFullName:
		opts.genTimestamp(t, msg)
		return true
	case durationFullName:
		opts.genDuration(t, msg)
		return true
	case anyFullName:
		return opts.genAny(t, msg, depth)
	case fieldMaskFullName:
		opts.genFieldMask(t, msg)
		return true
	default:
		fields := descriptor.Fields()
		n := fields.Len()
		for i := 0; i < n; i++ {
			field := fields.Get(i)
			if !rapid.Bool().Draw(t, fmt.Sprintf("gen-%s", field.Name())) {
				continue
			}

			opts.setFieldValue(t, msg, field, depth)
		}
		return true
	}
}

const (
	timestampFullName protoreflect.FullName = "google.protobuf.Timestamp"
	durationFullName                        = "google.protobuf.Duration"
	anyFullName                             = "google.protobuf.Any"
	fieldMaskFullName                       = "google.protobuf.FieldMask"
)

func (opts GeneratorOptions) setFieldValue(t *rapid.T, msg protoreflect.Message, field protoreflect.FieldDescriptor, depth int) {
	name := string(field.Name())
	kind := field.Kind()

	if field.IsList() {

		list := msg.Mutable(field).List()
		n := rapid.IntRange(0, 10).Draw(t, fmt.Sprintf("%sN", name))
		for i := 0; i < n; i++ {
			if kind == protoreflect.MessageKind || kind == protoreflect.GroupKind {
				if !opts.setFields(t, list.AppendMutable().Message(), depth+1) {
					list.Truncate(i)
				}
			} else {
				list.Append(opts.genScalarFieldValue(t, field, fmt.Sprintf("%s%d", name, i)))
			}
		}

	} else if field.IsMap() {

		m := msg.Mutable(field).Map()
		n := rapid.IntRange(0, 10).Draw(t, fmt.Sprintf("%sN", name))
		for i := 0; i < n; i++ {
			keyField := field.MapKey()
			valueField := field.MapValue()
			valueKind := valueField.Kind()
			key := opts.genScalarFieldValue(t, keyField, fmt.Sprintf("%s%d-key", name, i))
			if valueKind == protoreflect.MessageKind || valueKind == protoreflect.GroupKind {
				if !opts.setFields(t, m.Mutable(key.MapKey()).Message(), depth+1) {
					m.Clear(key.MapKey())
				}
			} else {
				value := opts.genScalarFieldValue(t, valueField, fmt.Sprintf("%s%d-key", name, i))
				m.Set(key.MapKey(), value)
			}
		}

	} else {

		if kind == protoreflect.MessageKind || kind == protoreflect.GroupKind {
			if !opts.setFields(t, msg.Mutable(field).Message(), depth+1) {
				msg.Clear(field)
			}
		} else {
			msg.Set(field, opts.genScalarFieldValue(t, field, name))
		}

	}
}

func (opts GeneratorOptions) genScalarFieldValue(t *rapid.T, field protoreflect.FieldDescriptor, name string) protoreflect.Value {
	switch field.Kind() {
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return protoreflect.ValueOfInt32(rapid.Int32().Draw(t, name))
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return protoreflect.ValueOfUint32(rapid.Uint32().Draw(t, name))
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return protoreflect.ValueOfInt64(rapid.Int64().Draw(t, name))
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return protoreflect.ValueOfUint64(rapid.Uint64().Draw(t, name))
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(rapid.Bool().Draw(t, name))
	case protoreflect.BytesKind:
		return protoreflect.ValueOfBytes(rapid.SliceOf(rapid.Byte()).Draw(t, name))
	case protoreflect.FloatKind:
		return protoreflect.ValueOfFloat32(rapid.Float32().Draw(t, name))
	case protoreflect.DoubleKind:
		return protoreflect.ValueOfFloat64(rapid.Float64().Draw(t, name))
	case protoreflect.EnumKind:
		enumValues := field.Enum().Values()
		val := rapid.Int32Range(0, int32(enumValues.Len()-1)).Draw(t, name)
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(val))
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(rapid.String().Draw(t, name))
	default:
		t.Fatalf("unexpected %v", field)
		return protoreflect.Value{}
	}
}

const (
	secondsName protoreflect.Name = "seconds"
	nanosName                     = "nanos"
)

func (opts GeneratorOptions) genTimestamp(t *rapid.T, msg protoreflect.Message) {
	seconds := rapid.Int64Range(-9999999999, 9999999999).Draw(t, "seconds")
	nanos := rapid.Int32Range(0, 999999999).Draw(t, "nanos")
	setSecondsNanosFields(t, msg, seconds, nanos)
}

func (opts GeneratorOptions) genDuration(t *rapid.T, msg protoreflect.Message) {
	seconds := rapid.Int64Range(0, 315576000000).Draw(t, "seconds")
	nanos := rapid.Int32Range(0, 999999999).Draw(t, "nanos")
	setSecondsNanosFields(t, msg, seconds, nanos)
}

func setSecondsNanosFields(t *rapid.T, message protoreflect.Message, seconds int64, nanos int32) {
	fields := message.Descriptor().Fields()

	secondsField := fields.ByName(secondsName)
	assert.Assert(t, secondsField != nil)
	message.Set(secondsField, protoreflect.ValueOfInt64(seconds))

	nanosField := fields.ByName(nanosName)
	assert.Assert(t, nanosField != nil)
	message.Set(nanosField, protoreflect.ValueOfInt32(nanos))

	return
}

const (
	typeUrlName = "type_url"
	valueName   = "value"
)

func (opts GeneratorOptions) genAny(t *rapid.T, msg protoreflect.Message, depth int) bool {
	if len(opts.AnyTypeURLs) == 0 {
		return false
	}

	fields := msg.Descriptor().Fields()

	typeUrl := rapid.SampledFrom(opts.AnyTypeURLs).Draw(t, "type_url")
	typ, err := opts.Resolver.FindMessageByURL(typeUrl)
	assert.NilError(t, err)

	typeUrlField := fields.ByName(typeUrlName)
	assert.Assert(t, typeUrlField != nil)
	msg.Set(typeUrlField, protoreflect.ValueOfString(typeUrl))

	valueMsg := typ.New()
	opts.setFields(t, valueMsg, depth+1)
	valueBz, err := proto.Marshal(valueMsg.Interface())
	assert.NilError(t, err)

	valueField := fields.ByName(valueName)
	assert.Assert(t, valueField != nil)
	msg.Set(valueField, protoreflect.ValueOfBytes(valueBz))

	return true
}

const (
	pathsName = "paths"
)

func (opts GeneratorOptions) genFieldMask(t *rapid.T, msg protoreflect.Message) {
	paths := rapid.SliceOfN(rapid.StringMatching("[a-z]+([.][a-z]+){0,2}"), 1, 5).Draw(t, "paths")
	pathsField := msg.Descriptor().Fields().ByName(pathsName)
	assert.Assert(t, pathsField != nil)
	pathsList := msg.NewField(pathsField).List()
	for _, path := range paths {
		pathsList.Append(protoreflect.ValueOfString(path))
	}
}
