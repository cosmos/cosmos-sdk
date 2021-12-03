package memkv

import (
	"fmt"

	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
)

type Debugger interface {
	Log(string)
	Decode(storeName string, key, value []byte) string
}

type TestStore struct {
	store   kv.Store
	decoder Debugger
	name    string
}

func (t TestStore) Get(key []byte) ([]byte, error) {
	val, err := t.store.Get(key)
	if err != nil {
		if t.decoder != nil {
			t.decoder.Log(fmt.Sprintf("ERR on GET %s: %v", t.decoder.Decode(t.name, key, nil), err))
		}
		return nil, err
	}
	if t.decoder != nil {
		t.decoder.Log(fmt.Sprintf("GET %s", t.decoder.Decode(t.name, key, val)))
	}
	return val, nil
}

func (t TestStore) Has(key []byte) (bool, error) {
	has, err := t.store.Has(key)
	if err != nil {
		if t.decoder != nil {
			t.decoder.Log(fmt.Sprintf("ERR on HAS %s: %v", t.decoder.Decode(t.name, key, nil), err))
		}
		return has, err
	}
	if t.decoder != nil {
		t.decoder.Log(fmt.Sprintf("HAS %s", t.decoder.Decode(t.name, key, nil)))
	}
	return has, nil
}

func (t TestStore) Iterator(start, end []byte) (kv.Iterator, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) ReverseIterator(start, end []byte) (kv.Iterator, error) {
	//TODO implement me
	panic("implement me")
}

func (t TestStore) Set(key, value []byte) error {
	if t.decoder != nil {
		t.decoder.Log(fmt.Sprintf("SET %s", t.decoder.Decode(t.name, key, value)))
	}
	err := t.store.Set(key, value)
	if err != nil {
		if t.decoder != nil {
			t.decoder.Log(fmt.Sprintf("ERR on SET %s: %v", t.decoder.Decode(t.name, key, value), err))
		}
		return err
	}
	return nil
}

func (t TestStore) Delete(key []byte) error {
	if t.decoder != nil {
		t.decoder.Log(fmt.Sprintf("DEL %s", t.decoder.Decode(t.name, key, nil)))
	}
	err := t.store.Delete(key)
	if err != nil {
		if t.decoder != nil {
			t.decoder.Log(fmt.Sprintf("ERR on SET %s: %v", t.decoder.Decode(t.name, key, nil), err))
		}
		return err
	}
	return nil
}

var _ kv.Store = &TestStore{}

type IndexCommitmentStore struct {
	commitment dbm.DB
	index      dbm.DB
}

func NewIndexCommitmentStore() *IndexCommitmentStore {
	return &IndexCommitmentStore{
		commitment: dbm.NewMemDB(),
		index:      dbm.NewMemDB(),
	}
}

var _ kv.IndexCommitmentStore = &IndexCommitmentStore{}

func (i IndexCommitmentStore) ReadCommitmentStore() kv.ReadStore {
	return i.commitment
}

func (i IndexCommitmentStore) ReadIndexStore() kv.ReadStore {
	return i.index
}

func (i IndexCommitmentStore) CommitmentStore() kv.Store {
	return i.commitment
}

func (i IndexCommitmentStore) IndexStore() kv.Store {
	return i.index
}
