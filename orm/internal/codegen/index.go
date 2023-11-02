//nolint:unused // ignore unused code linting
package codegen

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const indexKey = "IndexKey"

func (t tableGen) genIndexKeys() {
	// interface that all keys must adhere to
	t.P("type ", t.indexKeyInterfaceName(), " interface {")
	t.P("id() uint32")
	t.P("values() []interface{}")
	t.P(t.param(t.indexKeyInterfaceName()), "()")
	t.P("}")
	t.P()

	// start with primary key..
	t.P("// primary key starting index..")
	t.genIndex(t.table.PrimaryKey.Fields, 0, true)
	for _, idx := range t.table.Index {
		t.genIndex(idx.Fields, idx.Id, false)
	}
}

func (t tableGen) genIterator() {
	t.P("type ", t.iteratorName(), " struct {")
	t.P(ormTablePkg.Ident("Iterator"))
	t.P("}")
	t.P()
	t.genValueFunc()
	t.P()
}

func (t tableGen) genValueFunc() {
	varName := t.param(t.msg.GoIdent.GoName)
	t.P("func (i ", t.iteratorName(), ") Value() (*", t.QualifiedGoIdent(t.msg.GoIdent), ", error) {")
	t.P("var ", varName, " ", t.QualifiedGoIdent(t.msg.GoIdent))
	t.P("err := i.UnmarshalMessage(&", varName, ")")
	t.P("return &", varName, ", err")
	t.P("}")
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

func (t tableGen) indexKeyInterfaceName() string {
	return t.msg.GoIdent.GoName + indexKey
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
	return t.msg.GoIdent.GoName + joinedNames + indexKey
}

func (t tableGen) indexStructName(fields []string) string {
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = strcase.ToCamel(field)
	}
	joinedNames := strings.Join(names, "")
	return t.msg.GoIdent.GoName + joinedNames + indexKey
}

func (t tableGen) genIndex(fields string, id uint32, isPrimaryKey bool) {
	fieldsSlc := strings.Split(fields, ",")
	idxKeyName := t.indexStructName(fieldsSlc)

	if isPrimaryKey {
		t.P("type ", t.msg.GoIdent.GoName, "PrimaryKey = ", idxKeyName)
		t.P()
	}

	t.P("type ", idxKeyName, " struct {")
	t.P("vs []interface{}")
	t.P("}")

	t.genIndexInterfaceMethods(id, idxKeyName)

	for i := 1; i < len(fieldsSlc)+1; i++ {
		t.genWithMethods(idxKeyName, fieldsSlc[:i])
	}
}

func (t tableGen) genIndexInterfaceMethods(id uint32, indexStructName string) {
	funPrefix := fmt.Sprintf("func (x %s) ", indexStructName)
	t.P(funPrefix, "id() uint32 {return ", id, "}")
	t.P(funPrefix, "values() []interface{} {return x.vs}")
	t.P(funPrefix, t.param(t.indexKeyInterfaceName()), "() {}")
	t.P()
}

func (t tableGen) genWithMethods(indexStructName string, parts []string) {
	funcPrefix := fmt.Sprintf("func (this %s) ", indexStructName)
	camelParts := make([]string, len(parts))
	for i, part := range parts {
		camelParts[i] = strcase.ToCamel(part)
	}
	funcName := "With" + strings.Join(camelParts, "")

	t.P(funcPrefix, funcName, "(", t.fieldArgsFromStringSlice(parts), ") ", indexStructName, "{")
	t.P("this.vs = []interface{}{", strings.Join(parts, ","), "}")
	t.P("return this")
	t.P("}")
	t.P()
}
