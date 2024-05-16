// Package branch contains the core branch service interface.
package container

import "context"

type CacheService interface {
	OpenContainer(ctx context.Context) Service
}

type Service interface {
	Get(prefix []byte) (value any, ok bool)
	Remove(prefix []byte)
	Set(prefix []byte, value any)
}

