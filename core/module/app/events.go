package app

import (
	"context"
)

type EventListener interface {
	Handler

	RegisterEventListeners(EventListenerRegistrar)
}

type EventListenerRegistrar interface {
	OnEvent(eventType interface{}, listener func(ctx context.Context, event interface{}))
}
