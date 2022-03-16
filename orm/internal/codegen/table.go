package codegen

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	ormv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/orm/v1alpha1"
	"github.com/cosmos/cosmos-sdk/orm/internal/fieldnames"
	"github.com/cosmos/cosmos-sdk/orm/model/ormtable"
)

type tableGen struct {
	fileGen
	msg              *protogen.Message
	table            *ormv1alpha1.TableDescriptor
	primaryKeyFields fieldnames.FieldNames
	fields           map[protoreflect.Name]*protogen.Field
	uniqueIndexes    []*ormv1alpha1.SecondaryIndexDescriptor
	ormTable         ormtable.Table
}

const notFoundDocs = " returns nil and an error which responds true to ormerrors.IsNotFound() if the record was not found."

func newTableGen(fileGen fileGen, msg *protogen.Message, table *ormv1alpha1.TableDescriptor) (*tableGen, error) {
	t := &tableGen{fileGen: fileGen, msg: msg, table: table, fields: map[protoreflect.Name]*protogen.Field{}}
	t.primaryKeyFields = fieldnames.CommaSeparatedFieldNames(table.PrimaryKey.Fields)
	for _, field := range msg.Fields {
		t.fields[field.Desc.Name()] = field
	}
	uniqIndexes := make([]*ormv1alpha1.SecondaryIndexDescriptor, 0)
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
	if t.table.PrimaryKey.AutoIncrement {
		t.P(ormTablePkg.Ident("GenericAutoIncrementTable"), "[*", t.QualifiedGoIdent(t.msg.GoIdent), ",", t.indexKeyInterfaceName(), "]")
	} else {
		t.P(ormTablePkg.Ident("GenericTable"), "[*", t.QualifiedGoIdent(t.msg.GoIdent), ",", t.indexKeyInterfaceName(), "]")
	}
	t.P("Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error)")
	t.P("// Get", notFoundDocs)
	t.P("Get(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (*", t.QualifiedGoIdent(t.msg.GoIdent), ", error)")
	for _, idx := range t.uniqueIndexes {
		t.genUniqueIndexSig(idx)
	}
	t.P()
	t.P("doNotImplement()")
	t.P("}")
	t.P()
}

// returns the has and get (in that order) function signature for unique indexes.
func (t tableGen) uniqueIndexSig(idxFields string) (string, string, string) {
	fieldsSlc := strings.Split(idxFields, ",")
	camelFields := t.fieldsToCamelCase(idxFields)

	hasFuncName := "HasBy" + camelFields
	getFuncName := "GetBy" + camelFields
	args := t.fieldArgsFromStringSlice(fieldsSlc)

	hasFuncSig := fmt.Sprintf("%s (ctx context.Context, %s) (found bool, err error)", hasFuncName, args)
	getFuncSig := fmt.Sprintf("%s (ctx context.Context, %s) (*%s, error)", getFuncName, args, t.msg.GoIdent.GoName)
	return hasFuncSig, getFuncSig, getFuncName
}

func (t tableGen) genUniqueIndexSig(idx *ormv1alpha1.SecondaryIndexDescriptor) {
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
		t.P(ormTablePkg.Ident("GenericAutoIncrementTable"), "[*", t.QualifiedGoIdent(t.msg.GoIdent), ",", t.indexKeyInterfaceName(), "]")
	} else {
		t.P(ormTablePkg.Ident("GenericTable"), "[*", t.QualifiedGoIdent(t.msg.GoIdent), ",", t.indexKeyInterfaceName(), "]")
	}
	t.P("}")
	t.storeStructName()
}

func (t tableGen) genTableImpl() {
	receiverVar := "this"
	receiver := fmt.Sprintf("func (%s %s) ", receiverVar, t.messageTableReceiverName(t.msg))
	varName := t.param(t.msg.GoIdent.GoName)
	varTypeName := t.QualifiedGoIdent(t.msg.GoIdent)

	// Has
	t.P(receiver, "Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error) {")
	t.P("return ", receiverVar, ".DynamicTable().PrimaryKey().Has(ctx, ", t.primaryKeyFields.String(), ")")
	t.P("}")
	t.P()

	// Get
	t.P(receiver, "Get(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (*", varTypeName, ", error) {")
	t.P("var ", varName, " ", varTypeName)
	t.P("found, err := ", receiverVar, ".DynamicTable().PrimaryKey().Get(ctx, &", varName, ", ", t.primaryKeyFields.String(), ")")
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
		t.P("return ", receiverVar, ".DynamicTable().GetIndexByID(", idx.Id, ").(",
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
		t.P("found, err := ", receiverVar, ".DynamicTable().GetIndexByID(", idx.Id, ").(",
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
		t.P("return ", t.messageTableReceiverName(t.msg), "{", ormTablePkg.Ident("NewGenericAutoIncrementTable"), "[*", t.QualifiedGoIdent(t.msg.GoIdent), ",", t.indexKeyInterfaceName(), "](table.(", ormTablePkg.Ident("AutoIncrementTable"), "))}, nil")
	} else {
		t.P("return ", t.messageTableReceiverName(t.msg), "{", ormTablePkg.Ident("NewGenericTable"), "[*", t.QualifiedGoIdent(t.msg.GoIdent), ",", t.indexKeyInterfaceName(), "](table)}, nil")
	}
	t.P("}")
}
