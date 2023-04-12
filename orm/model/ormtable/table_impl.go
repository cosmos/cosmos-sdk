package ormtable

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"math"

	"github.com/cosmos/cosmos-sdk/orm/internal/fieldnames"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/orm/encoding/encodeutil"
	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
)

// tableImpl implements Table.
type tableImpl struct {
	*primaryKeyIndex
	indexes               []Index
	indexesByFields       map[fieldnames.FieldNames]concreteIndex
	uniqueIndexesByFields map[fieldnames.FieldNames]UniqueIndex
	indexesById           map[uint32]Index
	entryCodecsById       map[uint32]ormkv.EntryCodec
	tablePrefix           []byte
	tableId               uint32
	typeResolver          TypeResolver
	customJSONValidator   func(message proto.Message) error
}

func (t *tableImpl) GetTable(message proto.Message) Table {
	if message.ProtoReflect().Descriptor().FullName() == t.MessageType().Descriptor().FullName() {
		return t
	}
	return nil
}

func (t tableImpl) PrimaryKey() UniqueIndex {
	return t.primaryKeyIndex
}

func (t tableImpl) GetIndexByID(id uint32) Index {
	return t.indexesById[id]
}

func (t tableImpl) Save(ctx context.Context, message proto.Message) error {
	backend, err := t.getWriteBackend(ctx)
	if err != nil {
		return err
	}

	return t.save(ctx, backend, message, saveModeDefault)
}

func (t tableImpl) Insert(ctx context.Context, message proto.Message) error {
	backend, err := t.getWriteBackend(ctx)
	if err != nil {
		return err
	}

	return t.save(ctx, backend, message, saveModeInsert)
}

func (t tableImpl) Update(ctx context.Context, message proto.Message) error {
	backend, err := t.getWriteBackend(ctx)
	if err != nil {
		return err
	}

	return t.save(ctx, backend, message, saveModeUpdate)
}

func (t tableImpl) save(ctx context.Context, backend Backend, message proto.Message, mode saveMode) error {
	writer := newBatchIndexCommitmentWriter(backend)
	defer writer.Close()
	return t.doSave(ctx, writer, message, mode)
}

func (t tableImpl) doSave(ctx context.Context, writer *batchIndexCommitmentWriter, message proto.Message, mode saveMode) error {
	mref := message.ProtoReflect()
	pkValues, pk, err := t.EncodeKeyFromMessage(mref)
	if err != nil {
		return err
	}

	existing := mref.New().Interface()
	haveExisting, err := t.getByKeyBytes(writer, pk, pkValues, existing)
	if err != nil {
		return err
	}

	if haveExisting {
		if mode == saveModeInsert {
			return ormerrors.AlreadyExists.Wrapf("%q:%+v", mref.Descriptor().FullName(), pkValues)
		}

		if validateHooks := writer.ValidateHooks(); validateHooks != nil {
			err = validateHooks.ValidateUpdate(ctx, existing, message)
			if err != nil {
				return err
			}
		}
	} else {
		if mode == saveModeUpdate {
			return ormerrors.NotFound.Wrapf("%q", mref.Descriptor().FullName())
		}

		if validateHooks := writer.ValidateHooks(); validateHooks != nil {
			err = validateHooks.ValidateInsert(ctx, message)
			if err != nil {
				return err
			}
		}
	}

	// temporarily clear primary key
	t.ClearValues(mref)

	// store object
	bz, err := proto.MarshalOptions{Deterministic: true}.Marshal(message)
	if err != nil {
		return err
	}
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
		if writeHooks := writer.WriteHooks(); writeHooks != nil {
			writer.enqueueHook(func() {
				writeHooks.OnInsert(ctx, message)
			})
		}
	} else {
		existingMref := existing.ProtoReflect()
		for _, idx := range t.indexers {
			err = idx.onUpdate(indexStoreWriter, mref, existingMref)
			if err != nil {
				return err
			}
		}
		if writeHooks := writer.WriteHooks(); writeHooks != nil {
			writer.enqueueHook(func() {
				writeHooks.OnUpdate(ctx, existing, message)
			})
		}
	}

	return writer.Write()
}

func (t tableImpl) Delete(ctx context.Context, message proto.Message) error {
	pk := t.PrimaryKeyCodec.GetKeyValues(message.ProtoReflect())
	return t.doDelete(ctx, pk)
}

func (t tableImpl) GetIndex(fields string) Index {
	return t.indexesByFields[fieldnames.CommaSeparatedFieldNames(fields)]
}

func (t tableImpl) GetUniqueIndex(fields string) UniqueIndex {
	return t.uniqueIndexesByFields[fieldnames.CommaSeparatedFieldNames(fields)]
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
		var rawJSON json.RawMessage
		err := decoder.Decode(&rawJSON)
		if err != nil {
			return ormerrors.JSONImportError.Wrapf("%s", err)
		}

		if first {
			first = false
			if onFirst != nil {
				if onFirst(rawJSON) {
					// if onFirst handled this, skip decoding into a proto message
					continue
				}
			}
		}

		msg := t.MessageType().New().Interface()
		err = unmarshalOptions.Unmarshal(rawJSON, msg)
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
		}

		return DefaultJSONValidator(message)
	})
}

func (t tableImpl) ImportJSON(ctx context.Context, reader io.Reader) error {
	backend, err := t.getWriteBackend(ctx)
	if err != nil {
		return err
	}

	return t.decodeJson(reader, func(message proto.Message) error {
		return t.save(ctx, backend, message, saveModeDefault)
	})
}

func (t tableImpl) ExportJSON(context context.Context, writer io.Writer) error {
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	return t.doExportJSON(context, writer, true)
}

func (t tableImpl) doExportJSON(ctx context.Context, writer io.Writer, start bool) error {
	marshalOptions := protojson.MarshalOptions{
		UseProtoNames: true,
		Resolver:      t.typeResolver,
	}

	var err error
	it, _ := t.List(ctx, nil)
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
	err := encodeutil.SkipPrefix(r, t.tablePrefix)
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
		idx, ok := t.indexesByFields[fieldnames.FieldsFromNames(entry.Fields)]
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

func (t tableImpl) Has(ctx context.Context, message proto.Message) (found bool, err error) {
	backend, err := t.getBackend(ctx)
	if err != nil {
		return false, err
	}

	keyValues := t.primaryKeyIndex.PrimaryKeyCodec.GetKeyValues(message.ProtoReflect())
	return t.primaryKeyIndex.has(backend, keyValues)
}

// Get retrieves the message if one exists for the primary key fields
// set on the message. Other fields besides the primary key fields will not
// be used for retrieval.
func (t tableImpl) Get(ctx context.Context, message proto.Message) (found bool, err error) {
	backend, err := t.getBackend(ctx)
	if err != nil {
		return false, err
	}

	keyValues := t.primaryKeyIndex.PrimaryKeyCodec.GetKeyValues(message.ProtoReflect())
	return t.primaryKeyIndex.get(backend, message, keyValues)
}

var (
	_ Table  = &tableImpl{}
	_ Schema = &tableImpl{}
)

type saveMode int

const (
	saveModeDefault saveMode = iota
	saveModeInsert
	saveModeUpdate
)
