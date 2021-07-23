package container

import "fmt"

// Run runs the provided invoker function with values provided by the provided
// options. It is the single entry point for building and running a dependency
// injection container. Invoker should be a function taking one or more
// dependencies from the container, optionally returning an error.
//
// Ex:
//  Run(func (x int) error { println(x) }, Provide(func() int { return 1 }))
func Run(invoker interface{}, opts ...Option) error {
	opt := Options(opts...)

	cfg := &config{}
	err := opt.applyConfig(cfg)
	if cfg.err != nil {
		return err
	}

	ctr := &container{}
	err = opt.applyContainer(ctr)
	if cfg.err != nil {
		return err
	}

	return fmt.Errorf("TODO call invoker")
}
