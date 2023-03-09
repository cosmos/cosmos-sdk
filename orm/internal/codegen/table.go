package codegen

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	ormv1 "cosmossdk.io/api/cosmos/orm/v1"

	"github.com/cosmos/cosmos-sdk/orm/internal/fieldnames"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

type tableGen struct {
	fileGen
	msg              *protogen.Message
	table            *ormv1.TableDescriptor
	primaryKeyFields fieldnames.FieldNames
	fields           map[protoreflect.Name]*protogen.Field
	uniqueIndexes    []*ormv1.SecondaryIndexDescriptor
	ormTable         ormtable.Table
}

const notFoundDocs = " returns nil and an error which responds true to ormerrors.IsNotFound() if the record was not found."

func newTableGen(fileGen fileGen, msg *protogen.Message, table *ormv1.TableDescriptor) (*tableGen, error) {
	t := &tableGen{fileGen: fileGen, msg: msg, table: table, fields: map[protoreflect.Name]*protogen.Field{}}
	t.primaryKeyFields = fieldnames.CommaSeparatedFieldNames(table.PrimaryKey.Fields)
	for _, field := range msg.Fields {
		t.fields[field.Desc.Name()] = field
	}
	uniqIndexes := make([]*ormv1.SecondaryIndexDescriptor, 0)
	for _, idx := range t.table.Index {
		if idx.Unique {
			uniqIndexes = append(uniqIndexes, idx)
		}
	}
	t.uniqueIndexes = uniqIndexes
	var err error
	t.ormTable, err = ormtable.Build(ormtable.Options{
		MessageType:     dynamicpb.NewMessageType(msg.Desc),
		TableDescriptor: table,
	})
	return t, err
}

func (t tableGen) gen() {
	t.getTableInterface()
	t.genIterator()
	t.genIndexKeys()
	t.genStruct()
	t.genTableImpl()
	t.genTableImplGuard()
	t.genConstructor()
}

func (t tableGen) getTableInterface() {
	t.P("type ", t.messageTableInterfaceName(t.msg), " interface {")
	t.P("Insert(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	if t.table.PrimaryKey.AutoIncrement {
		t.P("InsertReturning", fieldsToCamelCase(t.table.PrimaryKey.Fields), "(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") (uint64, error)")
		t.P("LastInsertedSequence(ctx ", contextPkg.Ident("Context"), ") (uint64, error)")
	}
	t.P("Update(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Save(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Delete(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error)")
	t.P("// Get", notFoundDocs)
	t.P("Get(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (*", t.QualifiedGoIdent(t.msg.GoIdent), ", error)")

	for _, idx := range t.uniqueIndexes {
		t.genUniqueIndexSig(idx)
	}
	t.P("List(ctx ", contextPkg.Ident("Context"), ", prefixKey ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") ", "(", t.iteratorName(), ", error)")
	t.P("ListRange(ctx ", contextPkg.Ident("Context"), ", from, to ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") ", "(", t.iteratorName(), ", error)")
	t.P("DeleteBy(ctx ", contextPkg.Ident("Context"), ", prefixKey ", t.indexKeyInterfaceName(), ") error")
	t.P("DeleteRange(ctx ", contextPkg.Ident("Context"), ", from, to ", t.indexKeyInterfaceName(), ") error")
	t.P()
	t.P("doNotImplement()")
	t.P("}")
	t.P()
}

// returns the has and get (in that order) function signature for unique indexes.
func (t tableGen) uniqueIndexSig(idxFields string) (string, string, string) {
	fieldsSlc := strings.Split(idxFields, ",")
	camelFields := fieldsToCamelCase(idxFields)

	hasFuncName := "HasBy" + camelFields
	getFuncName := "GetBy" + camelFields
	args := t.fieldArgsFromStringSlice(fieldsSlc)

	hasFuncSig := fmt.Sprintf("%s (ctx context.Context, %s) (found bool, err error)", hasFuncName, args)
	getFuncSig := fmt.Sprintf("%s (ctx context.Context, %s) (*%s, error)", getFuncName, args, t.msg.GoIdent.GoName)
	return hasFuncSig, getFuncSig, getFuncName
}

func (t tableGen) genUniqueIndexSig(idx *ormv1.SecondaryIndexDescriptor) {
	hasSig, getSig, getFuncName := t.uniqueIndexSig(idx.Fields)
	t.P(hasSig)
	t.P("// ", getFuncName, notFoundDocs)
	t.P(getSig)
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

func (t tableGen) fieldArgsFromStringSlice(names []string) string {
	args := make([]string, len(names))
	for i, name := range names {
		args[i] = t.fieldArg(protoreflect.Name(name))
	}
	return strings.Join(args, ",")
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

func (t tableGen) genStruct() {
	t.P("type ", t.messageTableReceiverName(t.msg), " struct {")
	if t.table.PrimaryKey.AutoIncrement {
		t.P("table ", ormTablePkg.Ident("AutoIncrementTable"))
	} else {
		t.P("table ", ormTablePkg.Ident("Table"))
	}
	t.P("}")
	t.storeStructName()
}

func (t tableGen) genTableImpl() {
	receiverVar := "this"
	receiver := fmt.Sprintf("func (%s %s) ", receiverVar, t.messageTableReceiverName(t.msg))
	varName := t.param(t.msg.GoIdent.GoName)
	varTypeName := t.QualifiedGoIdent(t.msg.GoIdent)

	// these methods all have the same impl sans their names. so we can just loop and replace.
	methods := []string{"Insert", "Update", "Save", "Delete"}
	for _, method := range methods {
		t.P(receiver, method, "(ctx ", contextPkg.Ident("Context"), ", ", varName, " *", varTypeName, ") error {")
		t.P("return ", receiverVar, ".table.", method, "(ctx, ", varName, ")")
		t.P("}")
		t.P()
	}

	if t.table.PrimaryKey.AutoIncrement {
		t.P(receiver, "InsertReturning", fieldsToCamelCase(t.table.PrimaryKey.Fields), "(ctx ", contextPkg.Ident("Context"), ", ", varName, " *", varTypeName, ") (uint64, error) {")
		t.P("return ", receiverVar, ".table.InsertReturningPKey(ctx, ", varName, ")")
		t.P("}")
		t.P()

		t.P(receiver, "LastInsertedSequence(ctx ", contextPkg.Ident("Context"), ") (uint64, error) {")
		t.P("return ", receiverVar, ".table.LastInsertedSequence(ctx)")
		t.P("}")
		t.P()
	}

	// Has
	t.P(receiver, "Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error) {")
	t.P("return ", receiverVar, ".table.PrimaryKey().Has(ctx, ", t.primaryKeyFields.String(), ")")
	t.P("}")
	t.P()

	// Get
	t.P(receiver, "Get(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (*", varTypeName, ", error) {")
	t.P("var ", varName, " ", varTypeName)
	t.P("found, err := ", receiverVar, ".table.PrimaryKey().Get(ctx, &", varName, ", ", t.primaryKeyFields.String(), ")")
	t.P("if err != nil {")
	t.P("return nil, err")
	t.P("}")
	t.P("if !found {")
	t.P("return nil, ", ormErrPkg.Ident("NotFound"))
	t.P("}")
	t.P("return &", varName, ", nil")
	t.P("}")
	t.P()

	for _, idx := range t.uniqueIndexes {
		fields := strings.Split(idx.Fields, ",")
		hasName, getName, _ := t.uniqueIndexSig(idx.Fields)

		// has
		t.P("func (", receiverVar, " ", t.messageTableReceiverName(t.msg), ") ", hasName, "{")
		t.P("return ", receiverVar, ".table.GetIndexByID(", idx.Id, ").(",
			ormTablePkg.Ident("UniqueIndex"), ").Has(ctx,")
		for _, field := range fields {
			t.P(field, ",")
		}
		t.P(")")
		t.P("}")
		t.P()

		// get
		varName := t.param(t.msg.GoIdent.GoName)
		varTypeName := t.msg.GoIdent.GoName
		t.P("func (", receiverVar, " ", t.messageTableReceiverName(t.msg), ") ", getName, "{")
		t.P("var ", varName, " ", varTypeName)
		t.P("found, err := ", receiverVar, ".table.GetIndexByID(", idx.Id, ").(",
			ormTablePkg.Ident("UniqueIndex"), ").Get(ctx, &", varName, ",")
		for _, field := range fields {
			t.P(field, ",")
		}
		t.P(")")
		t.P("if err != nil {")
		t.P("return nil, err")
		t.P("}")
		t.P("if !found {")
		t.P("return nil, ", ormErrPkg.Ident("NotFound"))
		t.P("}")
		t.P("return &", varName, ", nil")
		t.P("}")
		t.P()
	}

	// List
	t.P(receiver, "List(ctx ", contextPkg.Ident("Context"), ", prefixKey ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") (", t.iteratorName(), ", error) {")
	t.P("it, err := ", receiverVar, ".table.GetIndexByID(prefixKey.id()).List(ctx, prefixKey.values(), opts...)")
	t.P("return ", t.iteratorName(), "{it}, err")
	t.P("}")
	t.P()

	// ListRange
	t.P(receiver, "ListRange(ctx ", contextPkg.Ident("Context"), ", from, to ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") (", t.iteratorName(), ", error) {")
	t.P("it, err := ", receiverVar, ".table.GetIndexByID(from.id()).ListRange(ctx, from.values(), to.values(), opts...)")
	t.P("return ", t.iteratorName(), "{it}, err")
	t.P("}")
	t.P()

	// DeleteBy
	t.P(receiver, "DeleteBy(ctx ", contextPkg.Ident("Context"), ", prefixKey ", t.indexKeyInterfaceName(), ") error {")
	t.P("return ", receiverVar, ".table.GetIndexByID(prefixKey.id()).DeleteBy(ctx, prefixKey.values()...)")
	t.P("}")
	t.P()
	t.P()

	// DeleteRange
	t.P(receiver, "DeleteRange(ctx ", contextPkg.Ident("Context"), ", from, to ", t.indexKeyInterfaceName(), ") error {")
	t.P("return ", receiverVar, ".table.GetIndexByID(from.id()).DeleteRange(ctx, from.values(), to.values())")
	t.P("}")
	t.P()
	t.P()

	t.P(receiver, "doNotImplement() {}")
	t.P()
}

func (t tableGen) genTableImplGuard() {
	t.P("var _ ", t.messageTableInterfaceName(t.msg), " = ", t.messageTableReceiverName(t.msg), "{}")
}

func (t tableGen) genConstructor() {
	iface := t.messageTableInterfaceName(t.msg)
	t.P("func New", iface, "(db ", ormTablePkg.Ident("Schema"), ") (", iface, ", error) {")
	t.P("table := db.GetTable(&", t.msg.GoIdent.GoName, "{})")
	t.P("if table == nil {")
	t.P("return nil,", ormErrPkg.Ident("TableNotFound.Wrap"), "(string((&", t.msg.GoIdent.GoName, "{}).ProtoReflect().Descriptor().FullName()))")
	t.P("}")
	if t.table.PrimaryKey.AutoIncrement {
		t.P(
			"return ", t.messageTableReceiverName(t.msg), "{table.(",
			ormTablePkg.Ident("AutoIncrementTable"), ")}, nil",
		)
	} else {
		t.P("return ", t.messageTableReceiverName(t.msg), "{table}, nil")
	}
	t.P("}")
}
