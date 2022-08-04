package db

import (
	_ "github.com/cosmos/cosmos-sdk/db/internal/backends"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	"github.com/cosmos/cosmos-sdk/db/types"
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
	BadgerDBBackend = types.BadgerDBBackend

	NewDB              = types.NewDB
	ReaderAsReadWriter = types.ReaderAsReadWriter
	NewVersionManager  = types.NewVersionManager

	NewMemDB = memdb.NewDB
)
