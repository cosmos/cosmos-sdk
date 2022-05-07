package container

// Run runs the provided invoker function with values provided by the provided
// options. It is the single entry point for building and running a dependency
// injection container. Invoker should be a function taking one or more
// dependencies from the container, optionally returning an error.
//
// Ex:
//  Run(func (x int) error { println(x) }, Provide(func() int { return 1 }))
//
// Run uses the debug mode provided by AutoDebug which means there will be
// verbose debugging information if there is an error and nothing upon success.
// Use RunDebug to configure behavior with more control.
func Run(invoker interface{}, opts ...Option) error {
	return RunDebug(invoker, AutoDebug(), opts...)
}

// RunDebug is a version of Run which takes an optional DebugOption for
// logging and visualization.
func RunDebug(invoker interface{}, debugOpt DebugOption, opts ...Option) error {
	opt := Options(opts...)

	cfg, err := newDebugConfig()
	if err != nil {
		return err
	}

	defer cfg.generateGraph() // always generate graph on exit

	if debugOpt != nil {
		err = debugOpt.applyConfig(cfg)
		if err != nil {
			return err
		}
	}

	cfg.logf("Registering providers")
	cfg.indentLogger()
	ctr := newContainer(cfg)
	err = opt.apply(ctr)
	if err != nil {
		cfg.logf("Failed registering providers because of: %+v", err)
		cfg.onError()
		return err
	}
	cfg.dedentLogger()

	err = ctr.run(invoker)
	if err != nil {
		cfg.onError()
	} else {
		cfg.onSuccess()
	}
	return err
}
