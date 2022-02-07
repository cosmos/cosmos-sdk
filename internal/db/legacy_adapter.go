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
type tmdbAdapter struct {
	dbm.DBReadWriter
	db dbm.DBConnection
}
type tmdbBatchAdapter struct {
	*tmdbAdapter
	written bool
}

var (
	_ tmdb.DB = (*tmdbAdapter)(nil)
)

// ConnectionAsTmdb returns a tmdb.DB which wraps a DBConnection.
func ConnectionAsTmdb(db dbm.DBConnection) *tmdbAdapter { return &tmdbAdapter{db.ReadWriter(), db} }

func (d *tmdbAdapter) Close() error { return d.db.Close() }

func (d *tmdbAdapter) sync() error {
	err := d.DBReadWriter.Commit()
	if err != nil {
		return err
	}
	d.DBReadWriter = d.db.ReadWriter()
	return nil
}
func (d *tmdbAdapter) DeleteSync(k []byte) error {
	err := d.DBReadWriter.Delete(k)
	if err != nil {
		return err
	}
	return d.sync()
}
func (d *tmdbAdapter) SetSync(k, v []byte) error {
	err := d.DBReadWriter.Set(k, v)
	if err != nil {
		return err
	}
	return d.sync()
}

func (d *tmdbAdapter) Iterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.DBReadWriter.Iterator(s, e)
	if err != nil {
		return nil, err
	}
	return DBToStoreIterator(it), nil
}
func (d *tmdbAdapter) ReverseIterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.DBReadWriter.ReverseIterator(s, e)
	if err != nil {
		return nil, err
	}
	return DBToStoreIterator(it), nil
}

// NewBatch returns a tmdb.Batch which wraps a DBWriter.
func (d *tmdbAdapter) NewBatch() tmdb.Batch {
	return &tmdbBatchAdapter{d, false}
}
func (d *tmdbAdapter) Print() error             { return nil }
func (d *tmdbAdapter) Stats() map[string]string { return nil }

func (d *tmdbBatchAdapter) Set(k, v []byte) error {
	if d.written {
		return errors.New("Batch already written")
	}
	return d.tmdbAdapter.Set(k, v)
}
func (d *tmdbBatchAdapter) Delete(k []byte) error {
	if d.written {
		return errors.New("Batch already written")
	}
	return d.tmdbAdapter.Delete(k)
}
func (d *tmdbBatchAdapter) WriteSync() error {
	if d.written {
		return errors.New("Batch already written")
	}
	d.written = true
	return d.sync()
}
func (d *tmdbBatchAdapter) Write() error {
	if d.written {
		return errors.New("Batch already written")
	}
	return d.WriteSync()
}
func (d *tmdbBatchAdapter) Close() error { return nil }
