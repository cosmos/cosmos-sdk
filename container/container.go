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

func (c *Container) ProvideWithScope(fn interface{}, scope Scope) error {
	panic("implement me")
}

func (c *Container) Invoke(fn interface{}) error {
	panic("TODO")
}

// InitializeAll attempts to initialize all registered providers in the container.
// It returns an error if a provider returns an error. If a given provider has
// dependencies which cannot be resolved, an error is not returned and instead
// that provider is not called.
func (c *Container) InitializeAll() error {
	panic("TODO")
}
