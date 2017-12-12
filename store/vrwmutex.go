package store

import (
	"sync"
	"sync/atomic"
)

// VRWMutex is a RWMutex with versions.
type VRWMutex interface {

	// Get the current version.  Blocks if locked.
	GetLockVersion() (lockVersion interface{})

	// Try to lock it for writing.
	// If the lockVersion is not the current version, it fails.
	// Only one writer will succeed for a given lockVersion.
	TryLock(lockVersion interface{}) bool

	// Release the write lock and update version to something new.
	Unlock()

	// Locks VRWMutex for reading.
	RLock()

	// RUnlock undoes a single RLock call.
	RUnlock()
}

// Implements VRWMutex.
type vrwMutex struct {

	// CONTRACT: reading/writing to `*written` should use `atomic.*`.
	// CONTRACT: replacing `written` with another *int32 should use `.mtx`.
	mtx     sync.RWMutex
	written *int32
}

func NewVRWMutex() *vrwMutex {
	return &vrwMutex{
		written: new(int32),
	}
}

func (mtx *vrwMutex) GetLockVersion() interface{} {
	mtx.mtx.RLock()
	defer mtx.mtx.RUnlock()

	return mtx.written
}

func (mtx *vrwMutex) TryLock(version interface{}) bool {
	if !mtx.trySetWritten() {
		return false
	}

	mtx.mtx.Lock()
	return true
}

func (mtx *vrwMutex) Unlock() {
	mtx.written = new(int32)
	mtx.mtx.Unlock()
}

func (mtx *vrwMutex) RLock() {
	mtx.mtx.RLock()
}

func (mtx *vrwMutex) RUnlock() {
	mtx.mtx.RUnlock()
}

func (mtx *vrwMutex) trySetWritten() bool {
	mtx.mtx.RLock()
	defer mtx.mtx.RUnlock()

	if version != mtx.written {
		return false // wrong write lock version
	}
	if !atomic.CompareAndSwapInt32(mtx.written, 0, 1) {
		return false // already written
	}

	return true
}
