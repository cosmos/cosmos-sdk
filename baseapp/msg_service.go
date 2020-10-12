package baseapp

// NewMsgServiceRouter creates a new MsgServiceRouter.
func NewMsgServiceRouter() *GRPCQueryRouter {
	return &GRPCQueryRouter{
		routes: map[string]GRPCQueryHandler{},
	}
}
