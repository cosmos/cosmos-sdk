package codegen

func (t tableGen) genQueries() {
	if !t.genQueryServer {
		return
	}

	name := t.msg.Desc.Name()

	// gen Get method
	t.P("func (x ", t.storeStructName(), ") Get", name, "(ctx ", contextPkg.Ident("Context"), ", req *Get", name, "Request) (*Get", name, "Response, error) {")
	t.P("res, err := x.", t.param(t.msg.GoIdent.GoName), ".Get(ctx, ", "")
	for _, name := range t.primaryKeyFields.Names() {
		for _, field := range t.msg.Fields {
			if field.Desc.Name() == name {
				t.P("req.", field.GoName, ",")
			}
		}
	}
	t.P(")")
	t.P("if err != nil { return nil, err }")
	t.P("return &Get", name, "Response{Value: res}, nil")
	t.P("}")
	t.P()

	// gen List method
	t.P("func (x ", t.storeStructName(), ") List", name, "(ctx ", contextPkg.Ident("Context"), ", req *List", name, "Request) (*List", name, "Response, error) {")
	t.P("return nil, nil")
	t.P("}")
}
