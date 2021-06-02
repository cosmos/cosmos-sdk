package container

type Registrar interface {
	Provide(fn interface{}) error
}
