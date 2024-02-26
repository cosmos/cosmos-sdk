package router

import gogoproto "github.com/cosmos/gogoproto/proto"

// Service is the interface that wraps the basic methods for a router service.
type Service interface {
	Handler(msg gogoproto.Message) Service
	HandlerByTypeURL(typeURL string) Service
}
