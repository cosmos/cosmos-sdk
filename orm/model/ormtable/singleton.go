package ormtable

import (
	"context"
	"encoding/json"
	"io"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// singleton implements a Table instance for singletons.
type singleton struct {
	*tableImpl
}

func (t singleton) DefaultJSON() json.RawMessage {
	msg := t.MessageType().New().Interface()
	bz, err := t.jsonMarshalOptions().Marshal(msg)
	if err != nil {
		return json.RawMessage("{}")
	}
	return bz
}

func (t singleton) ValidateJSON(reader io.Reader) error {
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
	}

	return DefaultJSONValidator(msg)
}

func (t singleton) ImportJSON(ctx context.Context, reader io.Reader) error {
	backend, err := t.getWriteBackend(ctx)
	if err != nil {
		return err
	}

	bz, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	msg := t.MessageType().New().Interface()
	err = protojson.Unmarshal(bz, msg)
	if err != nil {
		return err
	}

	return t.save(ctx, backend, msg, saveModeDefault)
}

func (t singleton) ExportJSON(ctx context.Context, writer io.Writer) error {
	msg := t.MessageType().New().Interface()
	found, err := t.Get(ctx, msg)
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

func (t singleton) jsonMarshalOptions() protojson.MarshalOptions {
	return protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "",
		UseProtoNames:   true,
		EmitUnpopulated: true,
		Resolver:        t.typeResolver,
	}
}

func (t *singleton) GetTable(message proto.Message) Table {
	if message.ProtoReflect().Descriptor().FullName() == t.MessageType().Descriptor().FullName() {
		return t
	}
	return nil
}
