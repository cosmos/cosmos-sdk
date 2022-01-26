package codegen

import (
	"fmt"

	"github.com/cosmos/cosmos-proto/generator"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	timestampMsgType  = (&timestamppb.Timestamp{}).ProtoReflect().Type()
	timestampFullName = timestampMsgType.Descriptor().FullName()
	durationMsgType   = (&durationpb.Duration{}).ProtoReflect().Type()
	durationFullName  = durationMsgType.Descriptor().FullName()
)

// fieldGoType only attempts to handle valid ORM field types
func fieldGoType(g *generator.GeneratedFile, field *protogen.Field) (goType string, err error) {
	pointer := field.Desc.HasPresence()
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "bool"
	case protoreflect.EnumKind:
		goType = g.QualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "uint64"
	case protoreflect.FloatKind:
		goType = "float32"
	case protoreflect.DoubleKind:
		goType = "float64"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.BytesKind:
		goType = "[]byte"
		pointer = false // rely on nullability of slices for presence
	case protoreflect.MessageKind:
		msgName := field.Desc.Message().FullName()
		switch msgName {
		case timestampFullName, durationFullName:
			return g.QualifiedGoIdent(field.Message.GoIdent), nil
		default:
			return "", fmt.Errorf("%s can't be used as an ORM index field", field.Message.Desc.FullName())
		}
	case protoreflect.GroupKind:
		return "", fmt.Errorf("groups can't be used as ORM index fields")
	}
	switch {
	case field.Desc.IsList():
		return "", fmt.Errorf("lists can't be used as ORM index fields")
	case field.Desc.IsMap():
		return "", fmt.Errorf("maps can't be used as ORM index fields")
	}
	if pointer {
		return "*" + goType, nil
	}
	return goType, nil
}
