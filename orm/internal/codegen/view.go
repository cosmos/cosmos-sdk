package codegen

func (t tableGen) genReaderInterface() {
	t.P("type ", t.messageReaderInterfaceName(t.msg), " interface {")
	t.P(t.getSig())
	t.P(t.listSig())
	t.P("}")
	t.P()
}
