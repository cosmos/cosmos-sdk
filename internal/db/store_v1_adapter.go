package db

import (
	"errors"

	dbm "github.com/cosmos/cosmos-sdk/db"

	tmdb "github.com/tendermint/tm-db"
)

// An adapter used to wrap objects supporting cosmos-sdk/db interface so that they support
// the tm-db interface.
//
// This serves as a transitional step in introducing the new DB interface while maintaining
// compatibility with existing code that expects the old interface.
type TmdbAdapter struct {
	dbm.DBReadWriter
	db dbm.DBConnection
}
type tmdbBatchAdapter struct {
	*TmdbAdapter
	closed bool
}

var (
	_ tmdb.DB = (*TmdbAdapter)(nil)
)

// ConnectionAsTmdb returns a tmdb.DB which wraps a DBConnection.
func ConnectionAsTmdb(db dbm.DBConnection) *TmdbAdapter { return &TmdbAdapter{db.ReadWriter(), db} }

func (d *TmdbAdapter) Close() error   { d.CloseTx(); return d.db.Close() }
func (d *TmdbAdapter) CloseTx() error { return d.DBReadWriter.Discard() }

func (d *TmdbAdapter) sync() error {
	err := d.DBReadWriter.Commit()
	if err != nil {
		return err
	}
	d.DBReadWriter = d.db.ReadWriter()
	return nil
}
func (d *TmdbAdapter) DeleteSync(k []byte) error {
	err := d.DBReadWriter.Delete(k)
	if err != nil {
		return err
	}
	return d.sync()
}
func (d *TmdbAdapter) SetSync(k, v []byte) error {
	err := d.DBReadWriter.Set(k, v)
	if err != nil {
		return err
	}
	return d.sync()
}

func (d *TmdbAdapter) Commit() (uint64, error) {
	err := d.DBReadWriter.Commit()
	if err != nil {
		return 0, err
	}
	v, err := d.db.SaveNextVersion()
	if err != nil {
		return 0, err
	}
	d.DBReadWriter = d.db.ReadWriter()
	return v, err
}

func (d *TmdbAdapter) Iterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.DBReadWriter.Iterator(s, e)
	if err != nil {
		return nil, err
	}
	return DBToStoreIterator(it), nil
}
func (d *TmdbAdapter) ReverseIterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.DBReadWriter.ReverseIterator(s, e)
	if err != nil {
		return nil, err
	}
	return DBToStoreIterator(it), nil
}

// NewBatch returns a tmdb.Batch which wraps a DBWriter.
func (d *TmdbAdapter) NewBatch() tmdb.Batch {
	return &tmdbBatchAdapter{d, false}
}
func (d *TmdbAdapter) Print() error             { return nil }
func (d *TmdbAdapter) Stats() map[string]string { return nil }

var errClosed = errors.New("batch is closed")

func (d *tmdbBatchAdapter) Set(k, v []byte) error {
	if d.closed {
		return errClosed
	}
	return d.TmdbAdapter.Set(k, v)
}
func (d *tmdbBatchAdapter) Delete(k []byte) error {
	if d.closed {
		return errClosed
	}
	return d.TmdbAdapter.Delete(k)
}
func (d *tmdbBatchAdapter) WriteSync() error {
	if d.closed {
		return errClosed
	}
	d.closed = true
	return d.sync()
}
func (d *tmdbBatchAdapter) Write() error { return d.WriteSync() }
func (d *tmdbBatchAdapter) Close() error { d.closed = true; return nil }
