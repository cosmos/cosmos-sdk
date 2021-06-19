package container

type Option interface {
	applyOption(*Container)
}

type option struct {
	f func(*Container)
}

func (o option) applyOption(ctr *Container) {
	o.f(ctr)
}

func Provide(constructors ...interface{}) Option {
	return option{
		func(container *Container) {
			for _, c := range constructors {
				err := container.container.Provide(c)
				if err != nil {
					container.err = err
					return
				}
			}
		},
	}
}

func Error(err error) Option {
	return option{func(container *Container) {
		container.err = err
	}}
}

func Options(opts ...Option) Option {
	return option{
		func(container *Container) {
			for _, opt := range opts {
				opt.applyOption(container)
				if container.err != nil {
					return
				}
			}
		},
	}
}
