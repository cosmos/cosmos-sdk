package container

import "go.uber.org/dig"

type container struct {
	container *dig.Container
	err       error
}

func Compose(invoker interface{}, opts ...Option) error {
	ctr := &container{
		container: dig.New(),
	}

	Options(opts...).applyOption(ctr)

	if ctr.err != nil {
		return ctr.err
	}

	return ctr.container.Invoke(invoker)
}

type In struct {
	dig.In
}

type Out struct {
	dig.Out
}
