package types

type Plugin func(ags AccountGetterSetter,
	caller *Account,
	input []byte,
	gas *int64) (result []byte, err error)
