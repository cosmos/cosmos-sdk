//nolint:unused
package multi

import (
	"bytes"
	"errors"

	dbm "github.com/cosmos/cosmos-sdk/db"
)

type (
	dbDeleteVersionFails struct{ dbm.Connection }
	dbRWCommitFails      struct{ dbm.Connection }
	dbRWCrudFails        struct{ dbm.Connection }
	dbSaveVersionFails   struct{ dbm.Connection }
	dbRevertFails        struct {
		dbm.Connection
		// order of calls to fail on (eg. [1, 0] => first call fails; second succeeds)
		failOn []bool
	}
)

type dbVersionsIs struct {
	dbm.Connection
	vset dbm.VersionSet
}

type (
	dbVersionsFails struct{ dbm.Connection }
	rwCommitFails   struct{ dbm.ReadWriter }
	rwCrudFails     struct {
		dbm.ReadWriter
		onKey []byte
	}
)

func (dbVersionsFails) Versions() (dbm.VersionSet, error) { return nil, errors.New("dbVersionsFails") }
func (db dbVersionsIs) Versions() (dbm.VersionSet, error) { return db.vset, nil }
func (db dbRWCrudFails) ReadWriter() dbm.ReadWriter {
	return rwCrudFails{db.Connection.ReadWriter(), nil}
}
func (dbSaveVersionFails) SaveVersion(uint64) error { return errors.New("dbSaveVersionFails") }
func (db dbRevertFails) Revert() error {
	fail := false
	if len(db.failOn) > 0 {
		fail, db.failOn = db.failOn[0], db.failOn[1:] //nolint:staticcheck
	}
	if fail {
		return errors.New("dbRevertFails")
	}
	return db.Connection.Revert()
}
func (dbDeleteVersionFails) DeleteVersion(uint64) error { return errors.New("dbDeleteVersionFails") }
func (tx rwCommitFails) Commit() error {
	tx.Discard()
	return errors.New("rwCommitFails")
}

func (db dbRWCommitFails) ReadWriter() dbm.ReadWriter {
	return rwCommitFails{db.Connection.ReadWriter()}
}

func (rw rwCrudFails) Get(k []byte) ([]byte, error) {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return nil, errors.New("rwCrudFails.Get")
	}
	return rw.ReadWriter.Get(k)
}

func (rw rwCrudFails) Has(k []byte) (bool, error) {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return false, errors.New("rwCrudFails.Has")
	}
	return rw.ReadWriter.Has(k)
}

func (rw rwCrudFails) Set(k []byte, v []byte) error {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return errors.New("rwCrudFails.Set")
	}
	return rw.ReadWriter.Set(k, v)
}

func (rw rwCrudFails) Delete(k []byte) error {
	if rw.onKey == nil || bytes.Equal(rw.onKey, k) {
		return errors.New("rwCrudFails.Delete")
	}
	return rw.ReadWriter.Delete(k)
}
