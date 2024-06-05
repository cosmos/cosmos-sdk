package codec

type HasName interface {
	// Name returns the name of key in the schema if one is defined or the empty string.
	// Multipart keys should separate names with commas, i.e. "name1,name2".
	Name() string
}

// NameableKeyCodec is a KeyCodec that can be named.
type NameableKeyCodec[T any] interface {
	KeyCodec[T]

	// WithName returns the KeyCodec with the provided name.
	WithName(name string) NamedKeyCodec[T]
}

// NamedKeyCodec is a KeyCodec that has a name.
type NamedKeyCodec[T any] interface {
	KeyCodec[T]
	HasName
}

// NameableValueCodec is a ValueCodec that can be named.
type NameableValueCodec[T any] interface {
	ValueCodec[T]

	// WithName returns the ValueCodec with the provided name.
	WithName(name string) NamedValueCodec[T]
}

// NamedValueCodec is a ValueCodec that has a name.
type NamedValueCodec[T any] interface {
	ValueCodec[T]
	HasName
}
