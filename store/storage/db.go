package storage

import (
	"sync"

	"cosmossdk.io/store/v2"
)

// Database defines the state storage backend.
type Database struct {
	lock sync.RWMutex
	db   store.Database
}
