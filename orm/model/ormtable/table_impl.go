package ormtable

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"io"
	"math"

	"google.golang.org/protobuf/encoding/protojson"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

type TableImpl struct {
	*PrimaryKeyIndex
	indexers              []Indexer
	indexes               []Index
	indexesByFields       map[FieldNames]concreteIndex
	uniqueIndexesByFields map[FieldNames]UniqueIndex
	indexesById           map[uint32]concreteIndex
	tablePrefix           []byte
	typeResolver          TypeResolver
	customImportValidator func(message proto.Message) error
}

type TypeResolver interface {
	protoregistry.MessageTypeResolver
	protoregistry.ExtensionTypeResolver
}

func (t TableImpl) Save(store kvstore.IndexCommitmentStore, message proto.Message, mode SaveMode) error {
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
			return sdkerrors.Wrapf(ormerrors.PrimaryKeyConstraintViolation, "%q", mref.Descriptor().FullName())
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

func (t TableImpl) Delete(store kvstore.IndexCommitmentStore, primaryKey []protoreflect.Value) error {
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

func (t TableImpl) GetIndex(fields FieldNames) Index {
	return t.indexesByFields[fields]
}

func (t TableImpl) GetUniqueIndex(fields FieldNames) UniqueIndex {
	return t.uniqueIndexesByFields[fields]
}

func (t TableImpl) Indexes() []Index {
	return t.indexes
}

func (t TableImpl) DefaultJSON() json.RawMessage {
	return json.RawMessage("[]")
}

func (t TableImpl) decodeJson(reader io.Reader, onMsg func(message proto.Message) error) error {
	decoder := json.NewDecoder(reader)
	token, err := decoder.Token()
	if err != nil {
		return err
	}

	if token != json.Delim('[') {
		return ormerrors.JSONImportError.Wrapf("expected [ got %s", token)
	}

	for decoder.More() {
		var rawJson json.RawMessage
		err = decoder.Decode(&rawJson)
		if err != nil {
			return ormerrors.JSONImportError.Wrapf("%s", err)
		}

		msg := t.MessageType().New().Interface()
		err = protojson.UnmarshalOptions{Resolver: t.typeResolver}.Unmarshal(rawJson, msg)
		if err != nil {
			return err
		}

		err = onMsg(msg)
		if err != nil {
			return err
		}
	}

	token, err = decoder.Token()
	if err != nil {
		return err
	}

	if token != json.Delim(']') {
		return ormerrors.JSONImportError.Wrapf("expected ] got %s", token)
	}

	return nil
}

func DefaultImportValidator(message proto.Message) error {
	if v, ok := message.(interface{ ValidateBasic() error }); ok {
		err := v.ValidateBasic()
		if err != nil {
			return err
		}
	}

	if v, ok := message.(interface{ Validate() error }); ok {
		err := v.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (t TableImpl) ValidateJSON(reader io.Reader) error {
	return t.decodeJson(reader, func(message proto.Message) error {
		if t.customImportValidator != nil {
			return t.customImportValidator(message)
		} else {
			return DefaultImportValidator(message)
		}
	})
}

func (t TableImpl) ImportJSON(store kvstore.IndexCommitmentStore, reader io.Reader) error {
	return t.decodeJson(reader, func(message proto.Message) error {
		return t.Save(store, message, SAVE_MODE_DEFAULT)
	})
}

func (t TableImpl) ExportJSON(store kvstore.IndexCommitmentReadStore, writer io.Writer) error {
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	it, _ := t.PrefixIterator(store, nil, IteratorOptions{})
	start := true
	for {
		found, err := it.Next()
		if err != nil {
			return err
		}

		if !found {
			_, err = writer.Write([]byte("]"))
			return err
		} else if !start {
			_, err = writer.Write([]byte(",\n"))
			if err != nil {
				return err
			}
		}
		start = false

		msg := t.MessageType().New().Interface()
		err = it.UnmarshalMessage(msg)
		if err != nil {
			return err
		}

		bz, err := protojson.Marshal(msg)
		if err != nil {
			return err
		}

		_, err = writer.Write(bz)
		if err != nil {
			return err
		}

	}
}

func (t TableImpl) DecodeKV(k, v []byte) (ormkv.Entry, error) {
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
