package codegen

import "google.golang.org/protobuf/compiler/protogen"

type tableGen struct {
	fileGen
	msg *protogen.Message
}

func (t tableGen) gen() {
	t.genStoreInterface()
	t.ge
}

func (t tableGen) genStoreInterface() {
	t.P("type ", t.messageStoreInterfaceName(t.msg), " interface {")
	t.P("Create", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Update", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Save", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("Delete", t.msg.GoIdent, "(", t.param(t.msg.GoIdent.GoName), " *", t.QualifiedGoIdent(t.msg.GoIdent), ") error")
	t.P("}")
	t.P()
}
