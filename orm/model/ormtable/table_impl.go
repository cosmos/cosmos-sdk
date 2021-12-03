package ormtable

import (
	"bytes"
	"encoding/binary"
	"math"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/orm/backend/kv"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/ormindex"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

type TableImpl struct {
	*ormindex.PrimaryKeyIndex
	indexers              []ormindex.Indexer
	indexes               []ormindex.Index
	indexesByFields       map[FieldNames]ormindex.Index
	uniqueIndexesByFields map[FieldNames]ormindex.UniqueIndex
	indexesById           map[uint32]ormindex.Index
	tablePrefix           []byte
	typeResolver          TypeResolver
}

type TypeResolver interface {
	protoregistry.MessageTypeResolver
	protoregistry.ExtensionTypeResolver
}

func (t TableImpl) Save(store kv.IndexCommitmentStore, message proto.Message, mode SaveMode) error {
	mref := message.ProtoReflect()
	pkValues, pk, err := t.EncodeKeyFromMessage(mref)
	if err != nil {
		return err
	}

	existing := mref.New().Interface()
	haveExisting, err := t.GetByKeyBytes(store, pk, pkValues, existing)
	if err != nil {
		return err
	}

	if haveExisting {
		if mode == SAVE_MODE_CREATE {
			return ormerrors.PrimaryKeyConstraintViolation.Wrapf("%q", mref.Descriptor().FullName())
		}
	} else {
		if mode == SAVE_MODE_UPDATE {
			return ormerrors.NotFoundOnUpdate.Wrapf("%q", mref.Descriptor().FullName())
		}
	}

	// temporarily clear primary key
	t.ClearValues(mref)

	// store object
	bz, err := proto.MarshalOptions{Deterministic: true}.Marshal(message)
	err = store.CommitmentStore().Set(pk, bz)
	if err != nil {
		return err
	}

	// set primary key again

	t.SetKeyValues(mref, pkValues)

	// set indexes
	indexStore := store.IndexStore()
	if !haveExisting {
		for _, idx := range t.indexers {
			err = idx.OnCreate(indexStore, mref)
			if err != nil {
				return err
			}

		}
	} else {
		existingMref := existing.ProtoReflect()
		for _, idx := range t.indexers {
			err = idx.OnUpdate(indexStore, mref, existingMref)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t TableImpl) Delete(store kv.IndexCommitmentStore, primaryKey []protoreflect.Value) error {
	pk, err := t.EncodeKey(primaryKey)
	if err != nil {
		return err
	}

	msg := t.MessageType().New().Interface()
	found, err := t.GetByKeyBytes(store, pk, primaryKey, msg)
	if err != nil {
		return err
	}

	if !found {
		return nil
	}

	// delete object
	err = store.CommitmentStore().Delete(pk)
	if err != nil {
		return err
	}

	// clear indexes
	mref := msg.ProtoReflect()
	indexStore := store.IndexStore()
	for _, idx := range t.indexers {
		err := idx.OnDelete(indexStore, mref)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t TableImpl) GetIndex(fields FieldNames) ormindex.Index {
	return t.indexesByFields[fields]
}

func (t TableImpl) GetUniqueIndex(fields FieldNames) ormindex.UniqueIndex {
	return t.uniqueIndexesByFields[fields]
}

func (t TableImpl) Indexes() []ormindex.Index {
	return t.indexes
}

func (t TableImpl) DecodeEntry(k, v []byte) (ormkv.Entry, error) {
	r := bytes.NewReader(k)
	if bytes.HasPrefix(k, t.tablePrefix) {
		err := ormkv.SkipPrefix(r, t.tablePrefix)
		if err != nil {
			return nil, err
		}

		id, err := binary.ReadUvarint(r)
		if err != nil {
			return nil, err
		}

		if id == 0 {
			return t.PrimaryKeyCodec.DecodeEntry(k, v)
		}

		if id > math.MaxUint32 {
			return nil, ormerrors.UnexpectedDecodePrefix.Wrapf("uint32 varint id out of range %d", id)
		}

		idx, ok := t.indexesById[uint32(id)]
		if !ok {
			return nil, ormerrors.UnexpectedDecodePrefix.Wrapf("can't find field with id %d", id)
		}

		return idx.DecodeEntry(k, v)
	} else {
		return nil, ormerrors.UnexpectedDecodePrefix
	}
}

func (t TableImpl) EncodeEntry(entry ormkv.Entry) (k, v []byte, err error) {
	switch entry := entry.(type) {
	case *ormkv.PrimaryKeyEntry:
		return t.PrimaryKeyCodec.EncodeEntry(entry)
	case *ormkv.IndexKeyEntry:
		idx, ok := t.indexesByFields[FieldsFromNames(entry.Fields)]
		if !ok {
			return nil, nil, ormerrors.BadDecodeEntry.Wrapf("can't find index with fields %s", entry.Fields)
		}

		return idx.EncodeEntry(entry)
	default:
		return nil, nil, ormerrors.BadDecodeEntry.Wrapf("%s", entry)
	}
}

var _ Table = &TableImpl{}
