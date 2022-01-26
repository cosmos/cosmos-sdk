package codegen

func (t tableGen) genIterator() {
	t.P("type ", t.iteratorName(), " struct {")
	t.P(tablePkg.Ident("Iterator"))
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
