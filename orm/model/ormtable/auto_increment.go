package ormtable

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"

	"github.com/cosmos/cosmos-sdk/orm/encoding/ormkv"
)

type AutoIncrementTable struct {
	*TableImpl
	autoIncField protoreflect.FieldDescriptor
	seqCodec     *ormkv.SeqCodec
}

func (s *AutoIncrementTable) Save(store kvstore.IndexCommitmentStore, message proto.Message, mode SaveMode) error {
	messageRef := message.ProtoReflect()
	val := messageRef.Get(s.autoIncField).Uint()
	if val == 0 {
		if mode == SAVE_MODE_UPDATE {
			return ormerrors.PrimaryKeyInvalidOnUpdate
		}

		mode = SAVE_MODE_INSERT
		key, err := s.nextSeqValue(store.IndexStore())
		if err != nil {
			return err
		}

		messageRef.Set(s.autoIncField, protoreflect.ValueOfUint64(key))
	} else {
		if mode == SAVE_MODE_INSERT {
			return ormerrors.AutoIncrementKeyAlreadySet
		}

		mode = SAVE_MODE_UPDATE
	}
	return s.TableImpl.Save(store, message, mode)
}

func (s *AutoIncrementTable) curSeqValue(kv kvstore.ReadStore) (uint64, error) {
	bz, err := kv.Get(s.seqCodec.Prefix())
	if err != nil {
		return 0, err
	}

	return s.seqCodec.DecodeValue(bz)
}

func (s *AutoIncrementTable) nextSeqValue(kv kvstore.Store) (uint64, error) {
	seq, err := s.curSeqValue(kv)
	if err != nil {
		return 0, err
	}

	seq++
	err = kv.Set(s.seqCodec.Prefix(), s.seqCodec.EncodeValue(seq))
	return seq, err
}

func (t AutoIncrementTable) EncodeEntry(entry ormkv.Entry) (k, v []byte, err error) {
	if _, ok := entry.(*ormkv.SeqEntry); ok {
		return t.seqCodec.EncodeEntry(entry)
	}
	return t.TableImpl.EncodeEntry(entry)
}

func (t AutoIncrementTable) ValidateJSON(reader io.Reader) error {
	return t.decodeAutoIncJson(reader, func(message proto.Message, haveId bool) error {
		if t.customJSONValidator != nil {
			return t.customJSONValidator(message)
		} else {
			return DefaultJSONValidator(message)
		}
	})
}

func (t AutoIncrementTable) ImportJSON(store kvstore.IndexCommitmentStore, reader io.Reader) error {
	return t.decodeAutoIncJson(reader, func(message proto.Message, haveId bool) error {
		return t.Save(store, message, SAVE_MODE_DEFAULT)
	})
}

func (t TableImpl) decodeAutoIncJson(reader io.Reader, onMsg func(message proto.Message, haveId bool) error) error {
	decoder, err := t.startDecodeJson(reader)
	if err != nil {
		return err
	}

	// TODO handle start with ID

	return t.doDecodeJson(decoder, func(message proto.Message) error {
		return onMsg(message, false)
	})
}

func (t AutoIncrementTable) ExportJSON(store kvstore.IndexCommitmentReadStore, writer io.Writer) error {
	_, err := writer.Write([]byte("["))
	if err != nil {
		return err
	}

	seq, err := t.curSeqValue(store.ReadIndexStore())
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
