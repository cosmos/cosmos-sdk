package db

import (
	tmdb "github.com/tendermint/tm-db"
)

// FIXME: hack, remove

// Pretend a new DBRW is a tm-db DB
type tmdbAdapterMunge struct {
	DBReadWriter
}

var _ tmdb.DB = tmdbAdapterMunge{}

func MungeTmdb(db DBReadWriter) tmdbAdapterMunge { return tmdbAdapterMunge{DBReadWriter: db} }

func (d tmdbAdapterMunge) Close() error { d.Discard(); return nil }

func (d tmdbAdapterMunge) DeleteSync(k []byte) error { return d.DBReadWriter.Delete(k) }
func (d tmdbAdapterMunge) SetSync(k, v []byte) error { return d.DBReadWriter.Set(k, v) }
func (d tmdbAdapterMunge) Iterator(s, e []byte) (tmdb.Iterator, error) {
	return d.DBReadWriter.Iterator(s, e)
}
func (d tmdbAdapterMunge) ReverseIterator(s, e []byte) (tmdb.Iterator, error) {
	return d.DBReadWriter.ReverseIterator(s, e)
}
func (d tmdbAdapterMunge) NewBatch() tmdb.Batch     { return nil }
func (d tmdbAdapterMunge) Print() error             { return nil }
func (d tmdbAdapterMunge) Stats() map[string]string { return nil }

// Pretend a tm-db DB is DBRW
type dbrwAdapterMunge struct {
	tmdb.DB
}

var _ DBReadWriter = dbrwAdapterMunge{}

func MungeDBRW(db tmdb.DB) dbrwAdapterMunge { return dbrwAdapterMunge{DB: db} }

func (d dbrwAdapterMunge) Commit() error { return nil }
func (d dbrwAdapterMunge) Discard()      { d.Close() }
func (d dbrwAdapterMunge) Iterator(s, e []byte) (Iterator, error) {
	return d.DB.Iterator(s, e)
}
func (d dbrwAdapterMunge) ReverseIterator(s, e []byte) (Iterator, error) {
	return d.DB.ReverseIterator(s, e)
}
