package ormtable

import (
	"encoding/json"
	"fmt"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

// autoIncrementTable is a Table implementation for tables with an
// auto-incrementing uint64 primary key.
type autoIncrementTable struct {
	*tableImpl
	autoIncField protoreflect.FieldDescriptor
	seqCodec     *ormkv.SeqCodec
}

func (t *autoIncrementTable) Save(store kvstore.IndexCommitmentStore, message proto.Message, mode SaveMode) error {
	messageRef := message.ProtoReflect()
	val := messageRef.Get(t.autoIncField).Uint()
	writer := store.NewWriter()
	defer writer.Close()

	if val == 0 {
		if mode == SAVE_MODE_UPDATE {
			return ormerrors.PrimaryKeyInvalidOnUpdate
		}

		mode = SAVE_MODE_INSERT
		key, err := t.nextSeqValue(writer.IndexStoreWriter())
		if err != nil {
			return err
		}

		messageRef.Set(t.autoIncField, protoreflect.ValueOfUint64(key))
	} else {
		if mode == SAVE_MODE_INSERT {
			return ormerrors.AutoIncrementKeyAlreadySet
		}

		mode = SAVE_MODE_UPDATE
	}

	hooks, _ := store.(Hooks)
	return t.tableImpl.doSave(writer, hooks, message, mode)
}

func (t *autoIncrementTable) curSeqValue(kv kvstore.Reader) (uint64, error) {
	bz, err := kv.Get(t.seqCodec.Prefix())
	if err != nil {
		return 0, err
	}

	return t.seqCodec.DecodeValue(bz)
}

func (t *autoIncrementTable) nextSeqValue(kv kvstore.Writer) (uint64, error) {
	seq, err := t.curSeqValue(kv)
	if err != nil {
		return 0, err
	}

	seq++
	return seq, t.setSeqValue(kv, seq)
}

func (t *autoIncrementTable) setSeqValue(kv kvstore.Writer, seq uint64) error {
	return kv.Set(t.seqCodec.Prefix(), t.seqCodec.EncodeValue(seq))
}

func (t autoIncrementTable) EncodeEntry(entry ormkv.Entry) (k, v []byte, err error) {
	if _, ok := entry.(*ormkv.SeqEntry); ok {
		return t.seqCodec.EncodeEntry(entry)
	}
	return t.tableImpl.EncodeEntry(entry)
}

func (t autoIncrementTable) ValidateJSON(reader io.Reader) error {
	return t.decodeAutoIncJson(nil, reader, func(message proto.Message, maxID uint64) error {
		messageRef := message.ProtoReflect()
		id := messageRef.Get(t.autoIncField).Uint()
		if id > maxID {
			return fmt.Errorf("invalid ID %d, expected a value <= %d", id, maxID)
		}

		if t.customJSONValidator != nil {
			return t.customJSONValidator(message)
		} else {
			return DefaultJSONValidator(message)
		}
	})
}

func (t autoIncrementTable) ImportJSON(store kvstore.IndexCommitmentStore, reader io.Reader) error {
	return t.decodeAutoIncJson(store, reader, func(message proto.Message, maxID uint64) error {
		messageRef := message.ProtoReflect()
		id := messageRef.Get(t.autoIncField).Uint()
		if id == 0 {
			// we don't have an ID in the JSON, so we call Save to insert and
			// generate one
			return t.Save(store, message, SAVE_MODE_INSERT)
		} else {
			if id > maxID {
				return fmt.Errorf("invalid ID %d, expected a value <= %d", id, maxID)
			}
			// we do have an ID and calling Save will fail because it expects
			// either no ID or SAVE_MODE_UPDATE. So instead we drop one level
			// down and insert using tableImpl which doesn't know about
			// auto-incrementing IDs
			return t.tableImpl.Save(store, message, SAVE_MODE_INSERT)
		}
	})
}

func (t autoIncrementTable) decodeAutoIncJson(store kvstore.IndexCommitmentStore, reader io.Reader, onMsg func(message proto.Message, maxID uint64) error) error {
	decoder, err := t.startDecodeJson(reader)
	if err != nil {
		return err
	}

	var seq uint64

	return t.doDecodeJson(decoder,
		func(message json.RawMessage) bool {
			err = json.Unmarshal(message, &seq)
			if err == nil {
				// writer is nil during validation
				if store != nil {
					writer := store.NewWriter()
					defer writer.Close()
					err = t.setSeqValue(writer.IndexStoreWriter(), seq)
					if err != nil {
						panic(err)
					}
					err = writer.Write()
					if err != nil {
						panic(err)
					}
				}
				return true
			}
			return false
		},
		func(message proto.Message) error {
			return onMsg(message, seq)
		})
}

func (t autoIncrementTable) ExportJSON(store kvstore.IndexCommitmentReadStore, writer io.Writer) error {
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	seq, err := t.curSeqValue(store.IndexStoreReader())
	if err != nil {
		return err
	}

	bz, err := json.Marshal(seq)
	if err != nil {
		return err
	}
	_, err = writer.Write(bz)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(",\n"))
	if err != nil {
		return err
	}

	return t.doExportJSON(store, writer)
}
