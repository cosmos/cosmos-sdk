package testkv

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/orm/pkg/kv"
	"github.com/dgraph-io/badger/v3"
	"log"
)

type KV struct {
	db *badger.DB
}

func New() *KV {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		panic(err)
	}
	return &KV{db: db}
}

func (b *KV) Set(_ context.Context, k, v []byte) {
	log.Printf("setting:\n\t key: %v\n\tkeyString: %s\n\tvalue: %v", k, k, v)
	txn := b.db.NewTransaction(true)
	err := txn.Set(k, v)
	if err != nil {
		panic(err)
	}
	err = txn.Commit()
	if err != nil {
		panic(err)
	}
	txn.Discard()
}

func (b *KV) Get(_ context.Context, k []byte) (v []byte) {
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			v = val
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	switch {
	case err == nil:
		return v
	case errors.Is(err, badger.ErrKeyNotFound):
		return nil
	default:
		panic(err)
	}
}

func (b *KV) Has(ctx context.Context, k []byte) bool {
	v := b.Get(ctx, k)
	return v != nil
}

func (b *KV) Delete(_ context.Context, k []byte) {
	txn := b.db.NewTransaction(true)
	err := txn.Delete(k)
	if err != nil {
		panic(err)
	}
	err = txn.Commit()
	if err != nil {
		panic(err)
	}
}

func (b *KV) Iterate(ctx context.Context, start, end []byte) kv.Iterator {
	txn := b.db.NewTransaction(true)
	iter := txn.NewKeyIterator(start, badger.IteratorOptions{
		PrefetchSize:   1,
		PrefetchValues: false,
		Reverse:        false,
		AllVersions:    false,
		InternalAccess: false,
	})
	iter.Rewind()
	return BadgerIterator{

		start: start,
		end:   end,
		iter:  iter,
	}
}

func (b *KV) IteratePrefix(ctx context.Context, prefix []byte) kv.Iterator {
	log.Printf("iterating prefix:\n\tstring: %s\n\t%v", prefix, prefix)
	txn := b.db.NewTransaction(true)
	iter := txn.NewIterator(badger.DefaultIteratorOptions)
	iter.Seek(prefix)
	return BadgerPrefixIterator{
		iter:      iter,
		prefix:    prefix,
		pfxLength: len(prefix),
	}
}

type BadgerPrefixIterator struct {
	ctx       context.Context
	iter      *badger.Iterator
	prefix    []byte
	pfxLength int
}

func (b BadgerPrefixIterator) Next() {
	if !b.Valid() {
		panic(fmt.Errorf("kv: iterator is not valid"))
	}
	b.iter.Next()
}

func (b BadgerPrefixIterator) Key() []byte {
	return b.iter.Item().Key()[b.pfxLength:]
}

func (b BadgerPrefixIterator) Value() []byte {
	v, err := b.iter.Item().ValueCopy(nil)
	if err != nil {
		panic(err)
	}
	return v
}

func (b BadgerPrefixIterator) Valid() bool {
	return b.iter.ValidForPrefix(b.prefix)
}

func (b BadgerPrefixIterator) Close() {
	b.iter.Close()
}

func (b BadgerPrefixIterator) Context() context.Context {
	return b.ctx
}

type BadgerIterator struct {
	ctx   context.Context
	start []byte
	end   []byte
	iter  *badger.Iterator
}

func (b BadgerIterator) Next() {
	if !b.Valid() {
		panic(fmt.Errorf("kv: iterator is not valid"))
	}
	b.iter.Next()
}

func (b BadgerIterator) Key() []byte {
	return b.iter.Item().Key()
}

func (b BadgerIterator) Value() []byte {
	v, err := b.iter.Item().ValueCopy(nil) // TODO maybe don't copy
	if err != nil {
		panic(err)
	}
	return v
}

func (b BadgerIterator) Valid() bool {
	if !b.iter.Valid() {
		return false
	}
	if len(b.end) != 0 {
		if bytes.Compare(b.iter.Item().Key(), b.end) == 1 {
			return false
		}
	}
	return true
}

func (b BadgerIterator) Close() {
	b.iter.Close()
}

func (b BadgerIterator) Context() context.Context {
	return b.ctx
}
