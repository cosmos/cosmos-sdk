package ormtable

import (
	"encoding/json"
	"io"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/cosmos/cosmos-sdk/orm/model/kvstore"
)

type Singleton struct {
	*TableImpl
}

func (t Singleton) DefaultJSON() json.RawMessage {
	msg := t.MessageType().New().Interface()
	bz, err := t.jsonMarshalOptions().Marshal(msg)
	if err != nil {
		return json.RawMessage("{}")
	}
	return bz
}

func (t Singleton) ValidateJSON(reader io.Reader) error {
	bz, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	msg := t.MessageType().New().Interface()
	err = protojson.Unmarshal(bz, msg)
	if err != nil {
		return err
	}

	if t.customJSONValidator != nil {
		return t.customJSONValidator(msg)
	} else {
		return DefaultJSONValidator(msg)
	}
}

func (t Singleton) ImportJSON(store kvstore.IndexCommitmentStore, reader io.Reader) error {
	bz, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	msg := t.MessageType().New().Interface()
	err = protojson.Unmarshal(bz, msg)
	if err != nil {
		return err
	}

	return t.Save(store, msg, SAVE_MODE_DEFAULT)
}

func (t Singleton) ExportJSON(store kvstore.IndexCommitmentReadStore, writer io.Writer) error {
	msg := t.MessageType().New().Interface()
	found, err := t.Get(store, nil, msg)
	if err != nil {
		return err
	}

	var bz []byte
	if !found {
		bz = t.DefaultJSON()
	} else {
		bz, err = t.jsonMarshalOptions().Marshal(msg)
		if err != nil {
			return err
		}
	}

	_, err = writer.Write(bz)
	return err
}

func (t Singleton) jsonMarshalOptions() protojson.MarshalOptions {
	return protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "",
		UseProtoNames:   true,
		EmitUnpopulated: true,
		Resolver:        t.typeResolver,
	}
}
