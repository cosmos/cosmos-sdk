package container

// Run runs the provided invoker function with values provided by the provided
// options. It is the single entry point for building and running a dependency
// injection container. Invoker should be a function taking one or more
// dependencies from the container, optionally returning an error.
//
// Ex:
//  Run(func (x int) error { println(x) }, Provide(func() int { return 1 }))
func Run(invoker interface{}, opts ...Option) error {
	opt := Options(opts...)

	cfg := newConfig()
	err := opt.applyConfig(cfg)
	if err != nil {
		return err
	}

	cfg.logf("Registering providers")
	cfg.indentLogger()
	ctr, _ := newContainer(cfg)
	defer ctr.generateGraph() // always generate graph on exit
	err = opt.applyContainer(ctr)
	if err != nil {
		cfg.logf("Failed registering providers because of: %+v", err)
		return err
	}
	cfg.dedentLogger()

	err = ctr.run(invoker)
	if err != nil {
		cfg.logf("Failed resolving dependencies because of: %+v", err)
	}
	return err
}
