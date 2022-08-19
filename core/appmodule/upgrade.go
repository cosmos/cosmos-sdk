package appmodule

import (
	"context"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type UpgradeHandler struct {
	FromModule protoreflect.FullName
	Handler    func(context.Context) error
}

func (h *Handler) RegisterUpgradeHandler(fromModule protoreflect.FullName, handler func(context.Context) error) {
	h.UpgradeHandlers = append(h.UpgradeHandlers, UpgradeHandler{
		FromModule: fromModule,
		Handler:    handler,
	})
}
