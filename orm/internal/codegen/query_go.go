package codegen

func (t tableGen) genQueries() {
	if !t.genQueryServer {
		return
	}

	name := t.msg.Desc.Name()

	storeStructName := t.storeStructName()
	// gen Get method
	t.P("func (x ", storeStructName, "QueryServer) ", name, "(ctx ", contextPkg.Ident("Context"), ", req *", name, "Request) (*", name, "Response, error) {")
	t.P("res, err := x.", storeStructName, ".", t.param(t.msg.GoIdent.GoName), ".Get(ctx, ", "")
	for _, name := range t.primaryKeyFields.Names() {
		for _, field := range t.msg.Fields {
			if field.Desc.Name() == name {
				t.P("req.", field.GoName, ",")
			}
		}
	}
	t.P(")")
	t.P("if err != nil { return nil, err }")
	t.P("return &", name, "Response{Value: res}, nil")
	t.P("}")
	t.P()

	// gen List method
	//t.P("func (x ", t.storeStructName(), ") List", name, "(ctx ", contextPkg.Ident("Context"), ", req *List", name, "Request) (*List", name, "Response, error) {")
	//t.P("return nil, nil")
	//t.P("}")
}
