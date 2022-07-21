// Adapters used to wrap objects supporting cosmos-sdk/db interface so that they support
// the tm-db interface.
//
// This serves as a transitional step in introducing the new DB interface while maintaining
// compatibility with existing code that expects the old interface.
package db

import (
	"errors"

	dbm "github.com/cosmos/cosmos-sdk/db/types"

	tmdb "github.com/tendermint/tm-db"
)

// TmdbTxnAdapter adapter wraps a single ReadWriter.
// Calling *Sync methods performs a commit and closes the transaction, invalidating it.
type TmdbTxnAdapter struct {
	dbm.ReadWriter
}

// TmdbConnAdapter wraps a DBConnection and a current transaction.
// When calling a *Sync method, a commit is performed and the transaction refreshed.
type TmdbConnAdapter struct {
	dbm.ReadWriter
	Connection dbm.Connection
}
type tmdbBatchAdapter struct {
	*TmdbConnAdapter
	closed bool
}

var (
	_ tmdb.DB = (*TmdbTxnAdapter)(nil)
	_ tmdb.DB = (*TmdbConnAdapter)(nil)
)

// ReadWriterAsTmdb returns a tmdb.DB which wraps a ReadWriter.
// Calling *Sync methods performs a commit and closes the transaction.
func ReadWriterAsTmdb(rw dbm.ReadWriter) TmdbTxnAdapter { return TmdbTxnAdapter{rw} }

// ConnectionAsTmdb returns a tmdb.DB which wraps a DBConnection.
func ConnectionAsTmdb(db dbm.Connection) *TmdbConnAdapter {
	return &TmdbConnAdapter{db.ReadWriter(), db}
}

func (d TmdbTxnAdapter) DeleteSync(k []byte) error {
	err := d.ReadWriter.Delete(k)
	if err != nil {
		return err
	}
	return d.Commit()
}
func (d TmdbTxnAdapter) SetSync(k, v []byte) error {
	err := d.ReadWriter.Set(k, v)
	if err != nil {
		return err
	}
	return d.Commit()
}

func (d TmdbTxnAdapter) Iterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.ReadWriter.Iterator(s, e)
	if err != nil {
		return nil, err
	}
	return ToStoreIterator(it), nil
}
func (d TmdbTxnAdapter) ReverseIterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.ReadWriter.ReverseIterator(s, e)
	if err != nil {
		return nil, err
	}
	return ToStoreIterator(it), nil
}

func (d TmdbTxnAdapter) Close() error             { return d.ReadWriter.Discard() }
func (d TmdbTxnAdapter) NewBatch() tmdb.Batch     { return d }
func (d TmdbTxnAdapter) Print() error             { return nil }
func (d TmdbTxnAdapter) Stats() map[string]string { return nil }

func (d TmdbTxnAdapter) Write() error     { return d.Commit() }
func (d TmdbTxnAdapter) WriteSync() error { return d.Commit() }

// TmdbConnAdapter:

func (d *TmdbConnAdapter) Close() error   { d.CloseTx(); return d.Connection.Close() }
func (d *TmdbConnAdapter) CloseTx() error { return d.ReadWriter.Discard() }

func (d *TmdbConnAdapter) sync() error {
	err := d.ReadWriter.Commit()
	if err != nil {
		return err
	}
	d.ReadWriter = d.Connection.ReadWriter()
	return nil
}
func (d *TmdbConnAdapter) DeleteSync(k []byte) error {
	err := d.ReadWriter.Delete(k)
	if err != nil {
		return err
	}
	return d.sync()
}
func (d *TmdbConnAdapter) SetSync(k, v []byte) error {
	err := d.ReadWriter.Set(k, v)
	if err != nil {
		return err
	}
	return d.sync()
}

func (d *TmdbConnAdapter) Commit() (uint64, error) {
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

func (d *TmdbConnAdapter) Iterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.ReadWriter.Iterator(s, e)
	if err != nil {
		return nil, err
	}
	return ToStoreIterator(it), nil
}
func (d *TmdbConnAdapter) ReverseIterator(s, e []byte) (tmdb.Iterator, error) {
	it, err := d.ReadWriter.ReverseIterator(s, e)
	if err != nil {
		return nil, err
	}
	return ToStoreIterator(it), nil
}

// NewBatch returns a tmdb.Batch which wraps a DBWriter.
func (d *TmdbConnAdapter) NewBatch() tmdb.Batch {
	return &tmdbBatchAdapter{d, false}
}
func (d *TmdbConnAdapter) Print() error             { return nil }
func (d *TmdbConnAdapter) Stats() map[string]string { return nil }

var errClosed = errors.New("batch is closed")

func (d *tmdbBatchAdapter) Set(k, v []byte) error {
	if d.closed {
		return errClosed
	}
	return d.TmdbConnAdapter.Set(k, v)
}
func (d *tmdbBatchAdapter) Delete(k []byte) error {
	if d.closed {
		return errClosed
	}
	return d.TmdbConnAdapter.Delete(k)
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
