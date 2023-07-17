package ormkv

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"cosmossdk.io/orm/internal/stablejson"
)

// Entry defines a logical representation of a kv-store entry for ORM instances.
type Entry interface {
	fmt.Stringer

	// GetTableName returns the table-name (equivalent to the fully-qualified
	// proto message name) this entry corresponds to.
	GetTableName() protoreflect.FullName

	// to allow new methods to be added without breakage, this interface
	// shouldn't be implemented outside this package,
	// see https://go.dev/blog/module-compatibility
	doNotImplement()
}

// PrimaryKeyEntry represents a logically decoded primary-key entry.
type PrimaryKeyEntry struct {
	// TableName is the table this entry represents.
	TableName protoreflect.FullName

	// Key represents the primary key values.
	Key []protoreflect.Value

	// Value represents the message stored under the primary key.
	Value proto.Message
}

func (p *PrimaryKeyEntry) GetTableName() protoreflect.FullName {
	return p.TableName
}

func (p *PrimaryKeyEntry) String() string {
	if p.Value == nil {
		return fmt.Sprintf("PK %s %s -> _", p.TableName, fmtValues(p.Key))
	}

	valBz, err := stablejson.Marshal(p.Value)
	valStr := string(valBz)
	if err != nil {
		valStr = fmt.Sprintf("ERR %v", err)
	}
	return fmt.Sprintf("PK %s %s -> %s", p.TableName, fmtValues(p.Key), valStr)
}

func fmtValues(values []protoreflect.Value) string {
	if len(values) == 0 {
		return "_"
	}

	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("%v", v.Interface())
	}

	return strings.Join(parts, "/")
}

func (p *PrimaryKeyEntry) doNotImplement() {}

// IndexKeyEntry represents a logically decoded index entry.
type IndexKeyEntry struct {
	// TableName is the table this entry represents.
	TableName protoreflect.FullName

	// Fields are the index fields this entry represents.
	Fields []protoreflect.Name

	// IsUnique indicates whether this index is unique or not.
	IsUnique bool

	// IndexValues represent the index values.
	IndexValues []protoreflect.Value

	// PrimaryKey represents the primary key values, it is empty if this is a
	// prefix key
	PrimaryKey []protoreflect.Value
}

func (i *IndexKeyEntry) GetTableName() protoreflect.FullName {
	return i.TableName
}

func (i *IndexKeyEntry) doNotImplement() {}

func (i *IndexKeyEntry) string() string {
	return fmt.Sprintf("%s %s : %s -> %s", i.TableName, fmtFields(i.Fields), fmtValues(i.IndexValues), fmtValues(i.PrimaryKey))
}

func fmtFields(fields []protoreflect.Name) string {
	strs := make([]string, len(fields))
	for i, field := range fields {
		strs[i] = string(field)
	}
	return strings.Join(strs, "/")
}

func (i *IndexKeyEntry) String() string {
	if i.IsUnique {
		return fmt.Sprintf("UNIQ %s", i.string())
	}

	return fmt.Sprintf("IDX %s", i.string())
}

// SeqEntry represents a sequence for tables with auto-incrementing primary keys.
type SeqEntry struct {
	// TableName is the table this entry represents.
	TableName protoreflect.FullName

	// Value is the uint64 value stored for this sequence.
	Value uint64
}

func (s *SeqEntry) GetTableName() protoreflect.FullName {
	return s.TableName
}

func (s *SeqEntry) doNotImplement() {}

func (s *SeqEntry) String() string {
	return fmt.Sprintf("SEQ %s %d", s.TableName, s.Value)
}

var _, _, _ Entry = &PrimaryKeyEntry{}, &IndexKeyEntry{}, &SeqEntry{}
