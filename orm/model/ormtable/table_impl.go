package ormtable

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// tableImpl implements Table.
type tableImpl struct {
	*PrimaryKeyIndex
	indexers              []indexer
	indexes               []Index
	indexesByFields       map[FieldNames]concreteIndex
	uniqueIndexesByFields map[FieldNames]UniqueIndex
	entryCodecsById       map[uint32]ormkv.EntryCodec
	tablePrefix           []byte
	tableId               uint32
	typeResolver          TypeResolver
	customJSONValidator   func(message proto.Message) error
}

func (t tableImpl) Save(store kvstore.Backend, message proto.Message, mode SaveMode) error {
	writer := newBatchIndexCommitmentWriter(store)
	defer writer.Close()
	return t.doSave(writer, message, mode)
}

func (t tableImpl) doSave(writer *batchIndexCommitmentWriter, message proto.Message, mode SaveMode) error {
	mref := message.ProtoReflect()
	pkValues, pk, err := t.EncodeKeyFromMessage(mref)
	if err != nil {
		return err
	}

	existing := mref.New().Interface()
	haveExisting, err := t.GetByKeyBytes(writer, pk, pkValues, existing)
	if err != nil {
		return err
	}

	if haveExisting {
		if mode == SAVE_MODE_INSERT {
			return sdkerrors.Wrapf(ormerrors.PrimaryKeyConstraintViolation, "%q:%+v", mref.Descriptor().FullName(), pkValues)
		}

		if hooks := writer.ORMHooks(); hooks != nil {
			err = hooks.OnUpdate(existing, message)
			if err != nil {
				return err
			}
		}
	} else {
		if mode == SAVE_MODE_UPDATE {
			return ormerrors.NotFoundOnUpdate.Wrapf("%q", mref.Descriptor().FullName())
		}

		if hooks := writer.ORMHooks(); hooks != nil {
			err = hooks.OnInsert(message)
			if err != nil {
				return err
			}
		}
	}

	// temporarily clear primary key
	t.ClearValues(mref)

	// store object
	bz, err := proto.MarshalOptions{Deterministic: true}.Marshal(message)
	err = writer.CommitmentStore().Set(pk, bz)
	if err != nil {
		return err
	}

	// set primary key again
	t.SetKeyValues(mref, pkValues)

	// set indexes
	indexStoreWriter := writer.IndexStore()
	if !haveExisting {
		for _, idx := range t.indexers {
			err = idx.onInsert(indexStoreWriter, mref)
			if err != nil {
				return err
			}

		}
	} else {
		existingMref := existing.ProtoReflect()
		for _, idx := range t.indexers {
			err = idx.onUpdate(indexStoreWriter, mref, existingMref)
			if err != nil {
				return err
			}
		}
	}

	return writer.Write()
}

func (t tableImpl) Delete(store kvstore.Backend, primaryKey []protoreflect.Value) error {
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

	if hooks := store.ORMHooks(); hooks != nil {
		err = hooks.OnDelete(msg)
		if err != nil {
			return err
		}
	}

	// delete object
	writer := newBatchIndexCommitmentWriter(store)
	defer writer.Close()
	err = writer.CommitmentStore().Delete(pk)
	if err != nil {
		return err
	}

	// clear indexes
	mref := msg.ProtoReflect()
	indexStoreWriter := writer.IndexStore()
	for _, idx := range t.indexers {
		err := idx.onDelete(indexStoreWriter, mref)
		if err != nil {
			return err
		}
	}

	return writer.Write()
}

func (t tableImpl) DeleteMessage(store kvstore.Backend, message proto.Message) error {
	pk := t.PrimaryKeyCodec.GetKeyValues(message.ProtoReflect())
	return t.Delete(store, pk)
}

func (t tableImpl) GetIndex(fields FieldNames) Index {
	return t.indexesByFields[fields]
}

func (t tableImpl) GetUniqueIndex(fields FieldNames) UniqueIndex {
	return t.uniqueIndexesByFields[fields]
}

func (t tableImpl) Indexes() []Index {
	return t.indexes
}

func (t tableImpl) DefaultJSON() json.RawMessage {
	return json.RawMessage("[]")
}

func (t tableImpl) decodeJson(reader io.Reader, onMsg func(message proto.Message) error) error {
	decoder, err := t.startDecodeJson(reader)
	if err != nil {
		return err
	}

	return t.doDecodeJson(decoder, nil, onMsg)
}

func (t tableImpl) startDecodeJson(reader io.Reader) (*json.Decoder, error) {
	decoder := json.NewDecoder(reader)
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}

	if token != json.Delim('[') {
		return nil, ormerrors.JSONImportError.Wrapf("expected [ got %s", token)
	}

	return decoder, nil
}

// onFirst is called on the first RawMessage and used for auto-increment tables
// to decode the sequence in which case it should return true.
// onMsg is called on every decoded message
func (t tableImpl) doDecodeJson(decoder *json.Decoder, onFirst func(message json.RawMessage) bool, onMsg func(message proto.Message) error) error {
	unmarshalOptions := protojson.UnmarshalOptions{Resolver: t.typeResolver}

	first := true
	for decoder.More() {
		var rawJson json.RawMessage
		err := decoder.Decode(&rawJson)
		if err != nil {
			return ormerrors.JSONImportError.Wrapf("%s", err)
		}

		if first {
			first = false
			if onFirst != nil {
				if onFirst(rawJson) {
					// if onFirst handled this, skip decoding into a proto message
					continue
				}
			}
		}

		msg := t.MessageType().New().Interface()
		err = unmarshalOptions.Unmarshal(rawJson, msg)
		if err != nil {
			return err
		}

		err = onMsg(msg)
		if err != nil {
			return err
		}
	}

	token, err := decoder.Token()
	if err != nil {
		return err
	}

	if token != json.Delim(']') {
		return ormerrors.JSONImportError.Wrapf("expected ] got %s", token)
	}

	return nil
}

// DefaultJSONValidator is the default validator used when calling
// Table.ValidateJSON(). It will call methods with the signature `ValidateBasic() error`
// and/or `Validate() error` to validate the message.
func DefaultJSONValidator(message proto.Message) error {
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

func (t tableImpl) ValidateJSON(reader io.Reader) error {
	return t.decodeJson(reader, func(message proto.Message) error {
		if t.customJSONValidator != nil {
			return t.customJSONValidator(message)
		} else {
			return DefaultJSONValidator(message)
		}
	})
}

func (t tableImpl) ImportJSON(store kvstore.Backend, reader io.Reader) error {
	return t.decodeJson(reader, func(message proto.Message) error {
		return t.Save(store, message, SAVE_MODE_DEFAULT)
	})
}

func (t tableImpl) ExportJSON(store kvstore.ReadBackend, writer io.Writer) error {
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	return t.doExportJSON(store, writer)
}

func (t tableImpl) doExportJSON(store kvstore.ReadBackend, writer io.Writer) error {
	marshalOptions := protojson.MarshalOptions{
		UseProtoNames: true,
		Resolver:      t.typeResolver,
	}

	var err error
	it, _ := t.PrefixIterator(store, nil, IteratorOptions{})
	start := true
	for {
		found := it.Next()

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

		bz, err := marshalOptions.Marshal(msg)
		if err != nil {
			return err
		}

		_, err = writer.Write(bz)
		if err != nil {
			return err
		}

	}
}

func (t tableImpl) DecodeEntry(k, v []byte) (ormkv.Entry, error) {
	r := bytes.NewReader(k)
	err := ormkv.SkipPrefix(r, t.tablePrefix)
	if err != nil {
		return nil, err
	}

	id, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}

	if id > math.MaxUint32 {
		return nil, ormerrors.UnexpectedDecodePrefix.Wrapf("uint32 varint id out of range %d", id)
	}

	idx, ok := t.entryCodecsById[uint32(id)]
	if !ok {
		return nil, ormerrors.UnexpectedDecodePrefix.Wrapf("can't find field with id %d", id)
	}

	return idx.DecodeEntry(k, v)
}

func (t tableImpl) EncodeEntry(entry ormkv.Entry) (k, v []byte, err error) {
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

func (t tableImpl) ID() uint32 {
	return t.tableId
}

var _ Table = &tableImpl{}
