package container

// Build runs the provided invoker function with values provided by the provided
// options. It is the single entry point for building and running a dependency
// injection container. Invoker should be a function taking one or more
// dependencies from the container, optionally returning an error.
//
// Ex:
//  Build(func (x int) error { println(x) }, Provide(func() int { return 1 }))
func Build(containerOption Option, outputs ...interface{}) error {
	loc := LocationFromCaller(1)
	return build(loc, nil, containerOption, outputs...)
}

// BuildDebug is a version of Build which takes an optional DebugOption for
// logging and visualization.
func BuildDebug(debugOpt DebugOption, option Option, outputs ...interface{}) error {
	loc := LocationFromCaller(1)
	return build(loc, debugOpt, option, outputs...)
}

func build(loc Location, debugOpt DebugOption, option Option, outputs ...interface{}) error {
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
	err = option.apply(ctr)
	if err != nil {
		cfg.logf("Failed registering providers because of: %+v", err)
		return err
	}
	cfg.dedentLogger()

	return ctr.build(loc, outputs...)
}
