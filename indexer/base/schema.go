package indexerbase

type Table interface {
	TypeName() string
	Fields() []Field
	PrimaryKey() []string
}

type Field struct {
	Name string
	Type string
}

type Schema interface {
	Tables() []Table
}
