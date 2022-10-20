package appmodule

import (
	"cosmossdk.io/depinject"
)

// Handler describes an ABCI app module handler. It can be injected into a
// depinject container as a one-per-module type (in the pointer variant).
type Handler struct {
	EventListeners []EventListener

	UpgradeHandlers []UpgradeHandler
}

func (h *Handler) IsOnePerModuleType() {}

var _ depinject.OnePerModuleType = &Handler{}
