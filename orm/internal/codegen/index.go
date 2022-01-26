package codegen

func (t tableGen) genIterator() {
	t.P("type ", t.iteratorName(), " struct {")
	t.P(tablePkg.Ident("Iterator"))
	t.P("}")
	t.P()
}
