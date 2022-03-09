package module

type Option interface {
	todo()
}

func Provide(providers ...interface{}) Option {
	panic("TODO")
}
