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

	cfg, err := newConfig()
	if err != nil {
		return err
	}

	defer cfg.generateGraph() // always generate graph on exit

	err = opt.applyConfig(cfg)
	if err != nil {
		return err
	}

	cfg.logf("Registering providers")
	cfg.indentLogger()
	ctr := newContainer(cfg)
	err = opt.applyContainer(ctr)
	if err != nil {
		cfg.logf("Failed registering providers because of: %+v", err)
		return err
	}
	cfg.dedentLogger()

	return ctr.run(invoker)
}
