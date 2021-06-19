package container

import "go.uber.org/dig"

type Container struct {
	container *dig.Container
	err       error
}

func New(opts ...Option) *Container {
	ctr := &Container{
		container: dig.New(),
	}

	Options(opts...).applyOption(ctr)

	return ctr
}

func (c Container) Invoke(f interface{}) error {
	return c.container.Invoke(f)
}
