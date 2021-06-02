package app

type HandlerVisitor interface {
	VisitAppHandler(info HandlerInfo, handler interface{})
}

type HandlerInfo struct {
	Name string
}
