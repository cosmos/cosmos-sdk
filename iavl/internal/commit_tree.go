package internal

import (
	"log/slog"
	"sync"
	"sync/atomic"
)

type CommitTree struct {
	latest     atomic.Pointer[NodePointer]
	root       *NodePointer
	writeMutex sync.Mutex
	logger     *slog.Logger
}
