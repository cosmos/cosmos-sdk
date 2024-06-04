package indexerbase

type EntityUpdate interface {
	TypeName() string
	PrimaryKey() []any
	IterateChanges(func(name string, value any) bool)
}

type EntityDelete interface {
	TypeName() string
	PrimaryKey() []any
	Prune() bool
}
