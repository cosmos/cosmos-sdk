package store

func assertValidKey(key []byte) {
	if key == nil {
		panic("key is nil")
	}
}

func assertValidValue(value []byte) {
	if value == nil {
		panic("value is nil")
	}
}
