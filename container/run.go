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
	cfg, err := newDebugConfig()
	if err != nil {
		return err
	}

	// debug cleanup
	defer func() {
		for _, f := range cfg.cleanup {
			f()
		}
	}()

	err = run(cfg, invoker, debugOpt, opts...)
	if err != nil {
		if cfg.onError != nil {
			err2 := cfg.onError.applyConfig(cfg)
			if err2 != nil {
				return err2
			}
		}
		return err
	} else {
		if cfg.onSuccess != nil {
			err2 := cfg.onSuccess.applyConfig(cfg)
			if err2 != nil {
				return err2
			}
		}
		return nil
	}
}

func run(cfg *debugConfig, invoker interface{}, debugOpt DebugOption, opts ...Option) error {
	opt := Options(opts...)

	defer cfg.generateGraph() // always generate graph on exit

	if debugOpt != nil {
		err := debugOpt.applyConfig(cfg)
		if err != nil {
			return err
		}
	}

	cfg.logf("Registering providers")
	cfg.indentLogger()
	ctr := newContainer(cfg)
	err := opt.apply(ctr)
	if err != nil {
		cfg.logf("Failed registering providers because of: %+v", err)
		return err
	}
	cfg.dedentLogger()

	return ctr.run(invoker)
}
