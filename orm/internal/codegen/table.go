package codegen

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"

	"github.com/iancoleman/strcase"

	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"

	"google.golang.org/protobuf/reflect/protoreflect"

	"google.golang.org/protobuf/compiler/protogen"

	ormv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/internal/fieldnames"
)

type tableGen struct {
	fileGen
	msg              *protogen.Message
	table            *ormv1alpha1.TableDescriptor
	primaryKeyFields fieldnames.FieldNames
	fields           map[protoreflect.Name]*protogen.Field
	ormTable         ormtable.Table
}

func newTableGen(fileGen fileGen, msg *protogen.Message, table *ormv1alpha1.TableDescriptor) *tableGen {
	t := &tableGen{fileGen: fileGen, msg: msg, table: table, fields: map[protoreflect.Name]*protogen.Field{}}
	t.primaryKeyFields = fieldnames.CommaSeparatedFieldNames(table.PrimaryKey.Fields)
	for _, field := range msg.Fields {
		t.fields[field.Desc.Name()] = field
	}
	var err error
	t.ormTable, err = ormtable.Build(ormtable.Options{
		MessageType:     dynamicpb.NewMessageType(msg.Desc),
		TableDescriptor: table,
	})
	if err != nil {
		panic(err)
	}
	return t
}

func (t tableGen) gen() {
	t.genStoreInterface()
	t.genReaderInterface()
	t.genIterator()
	t.genIndexKeys()
}

func (t tableGen) genStoreInterface() {
	t.P("type ", t.messageStoreInterfaceName(t.msg), " interface {")
	t.P(t.messageReaderInterfaceName(t.msg))
	t.P()
	t.P("Create", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Update", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Save", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Delete", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("}")
	t.P()
}

func (t tableGen) iteratorName() string {
	return t.msg.GoIdent.GoName + "Iterator"
}

func (t tableGen) getSig() string {
	res := "Get" + t.msg.GoIdent.GoName + "("
	res += t.fieldsArgs(t.primaryKeyFields.Names())
	res += ") (*" + t.QualifiedGoIdent(t.msg.GoIdent) + ", error)"
	return res
}

func (t tableGen) hasSig() string {
	res := "Has" + t.msg.GoIdent.GoName + "("
	res += t.fieldsArgs(t.primaryKeyFields.Names())
	res += ") (found bool, err error)"
	return res
}

func (t tableGen) listSig() string {
	res := "List" + t.msg.GoIdent.GoName + "("
	res += t.indexKeyInterfaceName()
	res += ") ("
	res += t.iteratorName()
	res += ", error)"
	return res
}

func (t tableGen) fieldsArgs(names []protoreflect.Name) string {
	var params []string
	for _, name := range names {
		params = append(params, t.fieldArg(name))
	}
	return strings.Join(params, ",")
}

func (t tableGen) fieldArg(name protoreflect.Name) string {
	typ, pointer := t.GeneratedFile.FieldGoType(t.fields[name])
	if pointer {
		typ = "*" + typ
	}
	return string(name) + " " + typ
}

func (t tableGen) genIndexKeys() {
	t.P("type ", t.indexKeyInterfaceName(), " interface {")
	t.P(("id() uint32"))
	t.P("values() []", protoreflectPackage.Ident("Value"))
	t.P(t.param(t.indexKeyInterfaceName()), "()")
	t.P("}")
	t.P()

	for _, index := range t.ormTable.Indexes() {
		indexCodec := index.(ormkv.IndexCodec)
		t.genIndexKey(indexCodec.GetFieldNames())
	}
}

func (t tableGen) indexKeyInterfaceName() string {
	return t.msg.GoIdent.GoName + "IndexKey"
}

func (t tableGen) genIndexKey(names []protoreflect.Name) {
	t.P("type ", t.indexKeyName(names), " struct {")
	t.P("}")
	t.P()
}

func (t tableGen) indexKeyName(names []protoreflect.Name) string {
	cnames := make([]string, len(names))
	for i, name := range names {
		cnames[i] = strcase.ToCamel(string(name))
	}
	joinedNames := strings.Join(cnames, "")
	return t.msg.GoIdent.GoName + joinedNames + "IndexKey"
}
