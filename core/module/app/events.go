package app

import (
	"context"
)

type EventListener interface {
	Module

	RegisterEventListeners(EventListenerRegistrar)
}

type EventListenerRegistrar interface {
	OnEvent(eventType interface{}, listener func(ctx context.Context, event interface{}))
}
