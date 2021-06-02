package container

type Container struct {
}

var _ Registrar = &Container{}

func NewContainer() *Container {
	panic("TODO")
}

func (c *Container) Provide(fn interface{}) error {
	panic("implement me")
}

func (c *Container) Invoke(fn interface{}) error {
	panic("TODO")
}
