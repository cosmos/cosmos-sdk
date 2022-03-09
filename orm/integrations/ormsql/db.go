package ormsql

import (
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoregistry"
	"reflect"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"gorm.io/gorm"
)

// DB wraps a gorm.DB instance with methods for indexing and querying
// ORM messages. Schema migration is handled automatically by gorm's
// auto-migration functionality. Indexers should use the ORM's
// EntryCodec functionality together with ADR-038 or ORM WriteHooks
// to listen for changes to the ORM state store, and then call the Index
// method to index those changes in the SQL database.
type DB struct {
	gormDb            *gorm.DB
	schema            *schema
	migratedMsgCodecs map[protoreflect.FullName]*messageCodec
}

type DBOptions struct {
	JsonMarshalOptions   protojson.MarshalOptions
	JsonUnmarshalOptions protojson.UnmarshalOptions
	MessageTypeResolver  protoregistry.MessageTypeResolver
}

// NewDB constructs a new DB instance.
func NewDB(gormDb *gorm.DB, options DBOptions) *DB {
	resolver := options.MessageTypeResolver
	if resolver == nil {
		resolver = protoregistry.GlobalTypes
	}

	return &DB{
		gormDb: gormDb,
		schema: &schema{
			jsonMarshalOptions:   options.JsonMarshalOptions,
			jsonUnmarshalOptions: options.JsonUnmarshalOptions,
			resolver:             resolver,
			messageCodecs:        map[protoreflect.FullName]*messageCodec{},
		},
		migratedMsgCodecs: map[protoreflect.FullName]*messageCodec{},
	}
}

func (d DB) save(message proto.Message) DB {
	cdc, err := d.getMessageCodec(message)
	if err != nil {
		d.gormDb.Error = err
		return d
	}
	val, err := cdc.encode(message.ProtoReflect())
	if err != nil {
		d.gormDb.Error = err
		return d
	}
	d.gormDb = d.gormDb.Table(cdc.tableName).Save(val.Interface())
	return d
}

var protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()

func (d DB) Where(query interface{}, args ...interface{}) DB {
	if protoMsg, ok := query.(proto.Message); ok {
		cdc, err := d.getMessageCodec(protoMsg)
		if err != nil {
			d.gormDb.Error = err
			return d
		}

		val, err := cdc.encode(protoMsg.ProtoReflect())
		if err != nil {
			d.gormDb.Error = err
			return d
		}
		query = val.Interface()
		d.gormDb = d.gormDb.Table(cdc.tableName)
	}
	d.gormDb = d.gormDb.Where(query, args)
	return d
}

func (d DB) Find(dest interface{}, args ...interface{}) DB {
	typ := reflect.TypeOf(dest).Elem()
	if typ.Kind() != reflect.Slice {
		d.gormDb.Error = fmt.Errorf("expected a slice, got %T", dest)
		return d
	}

	elem := typ.Elem()
	if !elem.AssignableTo(protoMessageType) {
		d.gormDb.Error = fmt.Errorf("expected a proto.Message slice type, got %T", dest)
		return d
	}

	msg := reflect.Zero(elem).Interface().(proto.Message)
	cdc, err := d.getMessageCodec(msg)
	if err != nil {
		d.gormDb.Error = err
		return d
	}

	structSliceType := reflect.SliceOf(cdc.structType)
	structSlicePtr := reflect.New(structSliceType)
	d.gormDb = d.gormDb.Table(cdc.tableName).Find(structSlicePtr.Interface(), args...)
	if d.gormDb.Error != nil {
		return d
	}
	structSlice := structSlicePtr.Elem()
	n := structSlice.Len()
	destVal := reflect.ValueOf(dest)
	resSlice := reflect.MakeSlice(typ, n, n)
	destVal.Elem().Set(resSlice)
	for i := 0; i < n; i++ {
		msg := cdc.msgType.New()
		err = cdc.decode(structSlice.Index(i), msg)
		if err != nil {
			d.gormDb.Error = err
			return d
		}
		resSlice.Index(i).Set(reflect.ValueOf(msg.Interface()))
	}
	return d
}

func (d DB) First(message proto.Message) DB {
	msgCdc, err := d.schema.messageCodecForType(message.ProtoReflect().Type())
	if err != nil {
		d.gormDb.Error = err
		return d
	}

	ptr := reflect.New(msgCdc.structType)
	d.gormDb = d.gormDb.Table(msgCdc.tableName).First(ptr.Interface())
	if d.gormDb.Error != nil {
		return d
	}

	d.gormDb.Error = msgCdc.decode(ptr.Elem(), message.ProtoReflect())
	return d
}

func (d DB) Error() error {
	return d.gormDb.Error
}

func (d DB) getMessageCodec(message proto.Message) (*messageCodec, error) {
	if cdc, ok := d.migratedMsgCodecs[message.ProtoReflect().Descriptor().FullName()]; ok {
		return cdc, nil
	}

	cdc, err := d.schema.getMessageCodec(message)
	if err != nil {
		return nil, err
	}

	// use a new instance because message may be nil
	newMsg := cdc.msgType.New()

	val, err := cdc.encode(newMsg)
	if err != nil {
		return nil, err
	}

	err = d.gormDb.Table(cdc.tableName).AutoMigrate(val.Interface())
	return cdc, err
}

// Index inserts, updates or deletes the entry with the provided primary key and
// value based on whether the entry exists or not and the deleted flag.
// It is designed to be used with either ADR-038 and the ORM EntryCodec or
// ORM WriteHooks.
func (d DB) Index(primaryKey []protoreflect.Value, value proto.Message, deleted bool) error {
	if !deleted {
		d.save(value)
		return d.gormDb.Error
	} else {
		cdc, err := d.getMessageCodec(value)
		if err != nil {
			return err
		}

		cond, err := cdc.deletionClause(primaryKey)
		if err != nil {
			return err
		}

		d.gormDb.Table(cdc.tableName).Where(cond).Delete(value)
		return d.gormDb.Error
	}
}
