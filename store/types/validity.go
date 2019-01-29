package types

func AssertValidKey(key []byte) {
	if key == nil {
		panic("key is nil")
	}
}

func AssertValidValue(value []byte) {
	if value == nil {
		panic("value is nil")
	}
}
