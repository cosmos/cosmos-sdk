package codegen

import (
	"fmt"
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
	t.genIterator()
	t.genIndexKeys()
	t.genStruct()
	t.genStoreImpl()
	t.genStoreImplGuard()
	t.genConstructor()
}

func (t tableGen) genStoreInterface() {
	t.P("type ", t.messageStoreInterfaceName(t.msg), " interface {")
	t.P("Insert(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Update(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Save(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Delete(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error)")
	t.P("Get(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (*", t.QualifiedGoIdent(t.msg.GoIdent), ", error)")
	t.P("List(ctx ", contextPkg.Ident("Context"), ", prefixKey ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") ", "(", t.iteratorName(), ", error)")
	t.P("ListRange(ctx ", contextPkg.Ident("Context"), ", from, to ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") ", "(", t.iteratorName(), ", error)")
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
	t.P("Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error)")
	return ""
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
	t.P("id() uint32")
	t.P("values() []interface{}")
	t.P(t.param(t.indexKeyInterfaceName()), "()")
	t.P("}")
	t.P()

	for _, index := range t.ormTable.Indexes() {
		indexCodec := index.(ormkv.IndexCodec)
		idxKeyName := t.indexKeyName(indexCodec.GetFieldNames())
		t.genIndexKey(idxKeyName)
		t.genIndexMethods(idxKeyName)
		t.genIndexInterfaceGuard(idxKeyName)
		t.genWithMethods(idxKeyName, index, indexCodec.GetFieldNames())
	}
}

func (t tableGen) indexKeyInterfaceName() string {
	return t.msg.GoIdent.GoName + "IndexKey"
}

func (t tableGen) genIndexKey(idxKeyName string) {
	t.P("type ", idxKeyName, " struct {")
	t.P("vs []interface{}")
	t.P("}")
	t.P()
}

func (t tableGen) indexKeyParts(names []protoreflect.Name) string {
	cnames := make([]string, len(names))
	for i, name := range names {
		cnames[i] = strcase.ToCamel(string(name))
	}
	return strings.Join(cnames, "")
}

func (t tableGen) indexKeyName(names []protoreflect.Name) string {
	cnames := make([]string, len(names))
	for i, name := range names {
		cnames[i] = strcase.ToCamel(string(name))
	}
	joinedNames := strings.Join(cnames, "")
	return t.msg.GoIdent.GoName + joinedNames + "IndexKey"
}

func (t tableGen) genIndexMethods(idxKeyName string) {
	receiverFunc := fmt.Sprintf("func (x %s) ", idxKeyName)
	t.P(receiverFunc, "id() uint32 { return ", t.table.Id, " /* primary key */ }")
	t.P(receiverFunc, "values() []interface{} { return x.vs }")
	t.P(receiverFunc, t.param(t.indexKeyInterfaceName()), "() {}")
	t.P()
}

func (t tableGen) genIndexInterfaceGuard(idxKeyName string) {
	t.P("var _ ", t.indexKeyInterfaceName(), " = ", idxKeyName, "{}")
	t.P()
}

func (t tableGen) genWithMethods(idxKeyName string, idx ormtable.Index, idxParts []protoreflect.Name) {
	receiverFunc := fmt.Sprintf("func (x %s) ", idxKeyName)
	for i := 0; i < len(idxParts)-1; i++ {
		t.P(receiverFunc, "With", strcase.ToCamel(string(idxParts[i])), "(", t.fieldArg(idxParts[i]), ") ", idxKeyName, " {")
		t.P("x.vs = []interface{}{", string(idxParts[i]), "}")
		t.P("return x")
		t.P("}")
	}

	strParts := make([]string, len(idxParts))
	for i, part := range idxParts {
		strParts[i] = string(part)
	}

	strParams := strings.Join(strParts, ",")

	t.P(receiverFunc, "With", t.indexKeyParts(idxParts), "(", t.fieldsArgs(idxParts), ") ", idxKeyName, "{")
	t.P("x.vs = []interface{}{", strParams, "}")
	t.P("return x")
	t.P("}")
}

func (t tableGen) genStruct() {
	t.P("type ", t.messageStoreReceiverName(t.msg), " struct {")
	t.P("table ", tablePkg.Ident("Table"))
	t.P("}")
	t.storeStructName()
}

func (t tableGen) genStoreImpl() {
	receiver := fmt.Sprintf("func (x %s) ", t.messageStoreReceiverName(t.msg))
	varName := t.param(t.msg.GoIdent.GoName)
	varTypeName := t.QualifiedGoIdent(t.msg.GoIdent)

	// these methods all have the same impl sans their names. so we can just loop and replace.
	methods := []string{"Insert", "Update", "Save", "Delete"}
	for _, method := range methods {
		t.P(receiver, method, "(ctx ", contextPkg.Ident("Context"), ", ", varName, " *", varTypeName, ") error {")
		t.P("return x.table.", method, "(ctx, ", varName, ")")
		t.P("}")
	}

	// Has
	t.P(receiver, "Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error) {")
	t.P("return x.table.PrimaryKey().Has(ctx, ", t.primaryKeyFields.String(), ")")
	t.P("}")

	// Get
	t.P(receiver, "Get(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (*", varTypeName, ", error) {")
	t.P("var ", varName, " ", varTypeName)
	t.P("found, err := x.table.PrimaryKey().Get(ctx, &", varName, ", ", t.primaryKeyFields.String(), ")")
	t.P("if !found {")
	t.P("return nil, err")
	t.P("}")
	t.P("return &", varName, ", err")
	t.P("}")

	// List
	t.P(receiver, "List(ctx ", contextPkg.Ident("Context"), ", prefixKey ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") (", t.iteratorName(), ", error) {")
	t.P("opts = append(opts, ", ormListPkg.Ident("Prefix"), "(prefixKey.values()))")
	t.P("it, err := x.table.GetIndexByID(prefixKey.id()).Iterator(ctx, opts...)")
	t.P("return ", t.iteratorName(), "{it}, err")
	t.P("}")

	// ListRange
	t.P(receiver, "ListRange(ctx ", contextPkg.Ident("Context"), ", from, to ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") (", t.iteratorName(), ", error) {")
	t.P("opts = append(opts, ", ormListPkg.Ident("Start"), "(from.values()), ", ormListPkg.Ident("End"), "(to))")
	t.P("it, err := x.table.GetIndexByID(from.id()).Iterator(ctx, opts...)")
	t.P("return ", t.iteratorName(), "{it}, err")
	t.P("}")
}

func (t tableGen) genStoreImplGuard() {
	t.P("var _ ", t.messageStoreInterfaceName(t.msg), " = ", t.messageStoreReceiverName(t.msg), "{}")
}

func (t tableGen) genConstructor() {
	iface := t.messageStoreInterfaceName(t.msg)
	t.P("func New", iface, "(db ", ormdbPkg.Ident("ModuleDB"), ") (", iface, ", error) {")
	t.P("table := db.GetTable(&", t.msg.GoIdent.GoName, "{})")
	t.P("if table == nil {")
	t.P("return nil,", ormErrPkg.Ident("TableNotFound.Wrap"), "(string((&", t.msg.GoIdent.GoName, "{}).ProtoReflect().Descriptor().FullName()))")
	t.P("}")
	t.P("return ", t.messageStoreReceiverName(t.msg), "{table}, nil")
	t.P("}")
}
