package root

import (
	"bytes"
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
type rwCrudFails struct {
	dbm.DBReadWriter
	onKey []byte
}

func (dbVersionsFails) Versions() (dbm.VersionSet, error) { return nil, errors.New("dbVersionsFails") }
func (db dbVersionsIs) Versions() (dbm.VersionSet, error) { return db.vset, nil }
func (db dbRWCrudFails) ReadWriter() dbm.DBReadWriter {
	return rwCrudFails{db.DBConnection.ReadWriter(), nil}
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

func (rw rwCrudFails) Get(k []byte) ([]byte, error) {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return nil, errors.New("rwCrudFails.Get")
	}
	return rw.DBReadWriter.Get(k)
}
func (rw rwCrudFails) Has(k []byte) (bool, error) {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return false, errors.New("rwCrudFails.Has")
	}
	return rw.DBReadWriter.Has(k)
}
func (rw rwCrudFails) Set(k []byte, v []byte) error {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return errors.New("rwCrudFails.Set")
	}
	return rw.DBReadWriter.Set(k, v)
}
func (rw rwCrudFails) Delete(k []byte) error {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return errors.New("rwCrudFails.Delete")
	}
	return rw.DBReadWriter.Delete(k)
}
