package ormkv

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
)

type Entry interface {
	fmt.Stringer
	GetTableName() protoreflect.FullName

	doNotImplement()
}

type PrimaryKeyEntry struct {
	Key   []protoreflect.Value
	Value proto.Message
}

func (p *PrimaryKeyEntry) GetTableName() protoreflect.FullName {
	return p.Value.ProtoReflect().Descriptor().FullName()
}

func (p *PrimaryKeyEntry) String() string {
	msg := p.Value
	name := msg.ProtoReflect().Descriptor().FullName()
	msgBz, err := protojson.Marshal(msg)
	msgStr := string(msgBz)
	if err != nil {
		msgStr = fmt.Sprintf("%s:%+v", name, msg)
	}
	return fmt.Sprintf("PK:%s:%s:%s", name, fmtValues(p.Key), msgStr)
}

func fmtValues(values []protoreflect.Value) string {
	var xs []interface{}
	for _, v := range values {
		xs = append(xs, v.Interface())
	}
	list, err := structpb.NewList(xs)
	if err != nil {
		return fmt.Sprintf("%+v", values)
	}
	bz, err := protojson.Marshal(list)
	if err != nil {
		return fmt.Sprintf("%+v", values)
	}
	return string(bz)
}

func (p *PrimaryKeyEntry) doNotImplement() {}

type IndexKeyEntry struct {
	TableName   protoreflect.FullName
	Fields      Fields
	IsUnique    bool
	IndexValues []protoreflect.Value
	// PrimaryKey is empty if this is a prefix key
	PrimaryKey []protoreflect.Value
}

func (i *IndexKeyEntry) GetTableName() protoreflect.FullName {
	return i.TableName
}

func (i *IndexKeyEntry) GetFields() Fields {
	return i.Fields
}

func (i *IndexKeyEntry) doNotImplement() {}

func (i *IndexKeyEntry) string() string {
	return fmt.Sprintf("%s%s:%s:%s", i.TableName, i.Fields, fmtValues(i.IndexValues), fmtValues(i.PrimaryKey))
}

func (i *IndexKeyEntry) String() string {
	if i.IsUnique {
		return fmt.Sprintf("UNIQ:%s", i.string())
	} else {

		return fmt.Sprintf("IDX:%s", i.string())
	}
}

type SeqEntry struct {
	TableName protoreflect.FullName
	Value     uint64
}

func (s *SeqEntry) GetTableName() protoreflect.FullName {
	return s.TableName
}

func (s *SeqEntry) doNotImplement() {}

func (s *SeqEntry) String() string {
	return fmt.Sprintf("SEQ:%s:%d", s.TableName, s.Value)
}

var _, _, _ Entry = &PrimaryKeyEntry{}, &IndexKeyEntry{}, &SeqEntry{}
