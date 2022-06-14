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
	dbm.ReadWriter
	Connection dbm.Connection
}
type tmdbBatchAdapter struct {
	*TmdbAdapter
	closed bool
}

var (
	_ tmdb.DB = (*TmdbAdapter)(nil)
)

// ConnectionAsTmdb returns a tmdb.DB which wraps a Connection.
func ConnectionAsTmdb(db dbm.Connection) *TmdbAdapter { return &TmdbAdapter{db.ReadWriter(), db} }

func (d *TmdbAdapter) Close() error   { d.CloseTx(); return d.Connection.Close() }
func (d *TmdbAdapter) CloseTx() error { return d.ReadWriter.Discard() }

func (d *TmdbAdapter) sync() error {
	err := d.ReadWriter.Commit()
	if err != nil {
		return err
	}
	d.ReadWriter = d.Connection.ReadWriter()
	return nil
}
func (d *TmdbAdapter) DeleteSync(k []byte) error {
	err := d.ReadWriter.Delete(k)
	if err != nil {
		return err
	}
	return d.sync()
}
func (d *TmdbAdapter) SetSync(k, v []byte) error {
	err := d.ReadWriter.Set(k, v)
	if err != nil {
		return err
	}
	return d.sync()
}

func (d *TmdbAdapter) Commit() (uint64, error) {
	err := d.ReadWriter.Commit()
	if err != nil {
		return 0, err
	}
	v, err := d.Connection.SaveNextVersion()
	if err != nil {
		return 0, err
	}
	d.ReadWriter = d.Connection.ReadWriter()
	return v, err
}

func (d *TmdbAdapter) Iterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.ReadWriter.Iterator(s, e)
	if err != nil {
		return nil, err
	}
	return ToStoreIterator(it), nil
}
func (d *TmdbAdapter) ReverseIterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.ReadWriter.ReverseIterator(s, e)
	if err != nil {
		return nil, err
	}
	return ToStoreIterator(it), nil
}

// NewBatch returns a tmdb.Batch which wraps a Writer.
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
