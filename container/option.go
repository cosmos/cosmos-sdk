package container

import "fmt"

type Option interface {
	applyOption(*container)
}

type option struct {
	f func(*container)
}

func (o option) applyOption(ctr *container) {
	o.f(ctr)
}

func Provide(constructors ...interface{}) Option {
	return option{
		func(container *container) {
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
	return option{func(container *container) {
		container.err = err
	}}
}

func Options(opts ...Option) Option {
	return option{
		func(container *container) {
			for _, opt := range opts {
				if opt == nil {
					container.err = fmt.Errorf("unexpected nil option")
					return
				}
				opt.applyOption(container)
				if container.err != nil {
					return
				}
			}
		},
	}
}
