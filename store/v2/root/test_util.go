package root

import (
	"errors"

	dbm "github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/memdb"
)

type dbDeleteVersionFails struct{ dbm.DBConnection }
type dbRWCommitFails struct{ *memdb.MemDB }
type dbRWCrudFails struct{ dbm.DBConnection }
type dbSaveVersionFails struct{ *memdb.MemDB }
type dbVersionsIs struct {
	dbm.DBConnection
	vset dbm.VersionSet
}
type dbVersionsFails struct{ dbm.DBConnection }
type rwCommitFails struct{ dbm.DBReadWriter }
type rwCrudFails struct{ dbm.DBReadWriter }

func (dbVersionsFails) Versions() (dbm.VersionSet, error) { return nil, errors.New("dbVersionsFails") }
func (db dbVersionsIs) Versions() (dbm.VersionSet, error) { return db.vset, nil }
func (db dbRWCrudFails) ReadWriter() dbm.DBReadWriter {
	return rwCrudFails{db.DBConnection.ReadWriter()}
}
func (dbSaveVersionFails) SaveVersion(uint64) error     { return errors.New("dbSaveVersionFails") }
func (dbDeleteVersionFails) DeleteVersion(uint64) error { return errors.New("dbDeleteVersionFails") }
func (tx rwCommitFails) Commit() error {
	tx.Discard()
	return errors.New("rwCommitFails")
}
func (db dbRWCommitFails) ReadWriter() dbm.DBReadWriter {
	return rwCommitFails{db.MemDB.ReadWriter()}
}

func (rwCrudFails) Get([]byte) ([]byte, error) { return nil, errors.New("rwCrudFails.Get") }
func (rwCrudFails) Has([]byte) (bool, error)   { return false, errors.New("rwCrudFails.Has") }
func (rwCrudFails) Set([]byte, []byte) error   { return errors.New("rwCrudFails.Set") }
func (rwCrudFails) Delete([]byte) error        { return errors.New("rwCrudFails.Delete") }
