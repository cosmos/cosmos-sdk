package indexerbase

type EntityUpdate interface {
	TypeName() string
	Key() []any
	Value() []any // ideally want a way to filter out unchanged fields
}

type EntityDelete interface {
	TypeName() string
	Key() []any
	Prune() bool
}
