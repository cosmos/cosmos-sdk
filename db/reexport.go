package db

import (
	"github.com/cosmos/cosmos-sdk/db/types"

	_ "github.com/cosmos/cosmos-sdk/db/internal/backends"
)

type (
	Connection      = types.Connection
	Reader          = types.Reader
	Writer          = types.Writer
	ReadWriter      = types.ReadWriter
	Iterator        = types.Iterator
	VersionSet      = types.VersionSet
	VersionIterator = types.VersionIterator
	BackendType     = types.BackendType
)

var (
	ErrVersionDoesNotExist = types.ErrVersionDoesNotExist

	MemDBBackend    = types.MemDBBackend
	RocksDBBackend  = types.RocksDBBackend
	BadgerDBBackend = types.BadgerDBBackend

	NewDB              = types.NewDB
	ReaderAsReadWriter = types.ReaderAsReadWriter
	NewVersionManager  = types.NewVersionManager
)
