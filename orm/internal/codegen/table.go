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
	if t.table.PrimaryKey.AutoIncrement {
		t.P("InsertReturningID(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") (uint64, error)")
	}
	t.P("Update(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Save(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Delete(ctx ", contextPkg.Ident("Context"), ", ", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Has(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (found bool, err error)")
	t.P("Get(ctx ", contextPkg.Ident("Context"), ", ", t.fieldsArgs(t.primaryKeyFields.Names()), ") (*", t.QualifiedGoIdent(t.msg.GoIdent), ", error)")

	_, _, pkDeleteSig := t.uniqueIndexSig(t.table.PrimaryKey.Fields)
	t.P(pkDeleteSig)

	for _, idx := range t.uniqueIndexes {
		t.genUniqueIndexSig(idx)
	}
	t.P("List(ctx ", contextPkg.Ident("Context"), ", prefixKey ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") ", "(", t.iteratorName(), ", error)")
	t.P("ListRange(ctx ", contextPkg.Ident("Context"), ", from, to ", t.indexKeyInterfaceName(), ", opts ...", ormListPkg.Ident("Option"), ") ", "(", t.iteratorName(), ", error)")
	t.P()
	t.P("doNotImplement()")
	t.P("}")
	t.P()
}

// returns the has, get and delete (in that order) function type signatures 
// for unique indexes.
func (t tableGen) uniqueIndexSig(idxFields string) (string, string, string) {
	fieldsSlc := strings.Split(idxFields, ",")
	camelFields := t.fieldsToCamelCase(idxFields)

	hasFuncName := "HasBy" + camelFields
	getFuncName := "GetBy" + camelFields
	deleteFuncName := "DeleteBy" + camelFields
	args := t.fieldArgsFromStringSlice(fieldsSlc)

	hasFuncSig := fmt.Sprintf("%s (ctx context.Context, %s) (found bool, err error)", hasFuncName, args)
	getFuncSig := fmt.Sprintf("%s (ctx context.Context, %s) (*%s, error)", getFuncName, args, t.msg.GoIdent.GoName)
	deleteFuncSig := fmt.Sprintf("%s (ctx context.Context, %s) error", deleteFuncName, args)
	return hasFuncSig, getFuncSig, deleteFuncSig
}

func (t tableGen) genUniqueIndexSig(idx *ormv1alpha1.SecondaryIndexDescriptor) {
	hasSig, getSig, deleteSig := t.uniqueIndexSig(idx.Fields)
	t.P(hasSig)
	t.P(getSig)
	t.P(deleteSig)
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
	t.P("type ", t.messageStoreReceiverName(t.msg), " struct {")
	if t.table.PrimaryKey.AutoIncrement {
		t.P("table ", ormTablePkg.Ident("AutoIncrementTable"))
	} else {
		t.P("table ", ormTablePkg.Ident("Table"))
	}
	t.P("}")
	t.storeStructName()
}

func (t tableGen) genStoreImpl() {
	receiverVar := "this"
	receiver := fmt.Sprintf("func (%s %s) ", receiverVar, t.messageStoreReceiverName(t.msg))
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
		t.P(receiver, "InsertReturningID(ctx ", contextPkg.Ident("Context"), ", ", varName, " *", varTypeName, ") (uint64, error) {")
		t.P("return ", receiverVar, ".table.InsertReturningID(ctx, ", varName, ")")
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
	t.P("if !found {")
	t.P("return nil, err")
	t.P("}")
	t.P("return &", varName, ", err")
	t.P("}")
	t.P()

	t.deleteByImpl(receiverVar, 0, t.table.PrimaryKey.Fields)

	for _, idx := range t.uniqueIndexes {
		fields := strings.Split(idx.Fields, ",")
		hasName, getName, _ := t.uniqueIndexSig(idx.Fields)

		// has
		t.P("func (", receiverVar, " ", t.messageStoreReceiverName(t.msg), ") ", hasName, "{")
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
		t.P("func (", receiverVar, " ", t.messageStoreReceiverName(t.msg), ") ", getName, "{")
		t.P("var ", varName, " ", varTypeName)
		t.P("found, err := ", receiverVar, ".table.GetIndexByID(", idx.Id, ").(",
			ormTablePkg.Ident("UniqueIndex"), ").Get(ctx, &", varName, ",")
		for _, field := range fields {
			t.P(field, ",")
		}
		t.P(")")
		t.P("if !found {")
		t.P("return nil, err")
		t.P("}")
		t.P("return &", varName, ", nil")
		t.P("}")
		t.P()

		// delete
		t.deleteByImpl(receiverVar, idx.Id, idx.Fields)
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

	t.P(receiver, "doNotImplement() {}")
	t.P()
}

func (t tableGen) deleteByImpl(receiverVar string, idxId uint32, fields string) {
	_, _, deleteName := t.uniqueIndexSig(fields)
	fieldNames := strings.Split(fields, ",")
	t.P("func (", receiverVar, " ", t.messageStoreReceiverName(t.msg), ") ", deleteName, "{")
	t.P("return ", receiverVar, ".table.GetIndexByID(", idxId, ").(",
		ormTablePkg.Ident("UniqueIndex"), ").DeleteByKey(ctx,")
	for _, field := range fieldNames {
		t.P(field, ",")
	}
	t.P(")")
	t.P("}")
	t.P()
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
	if t.table.PrimaryKey.AutoIncrement {
		t.P(
			"return ", t.messageStoreReceiverName(t.msg), "{table.(",
			ormTablePkg.Ident("AutoIncrementTable"), ")}, nil",
		)
	} else {
		t.P("return ", t.messageStoreReceiverName(t.msg), "{table}, nil")
	}
	t.P("}")
}
