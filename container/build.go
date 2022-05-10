package container

// Build builds the container specified by containerOption and extracts the
// requested outputs from the container or returns an error. It is the single
// entry point for building and running a dependency injection container.
// Each of the values specified as outputs must be pointers to types that
// can be provided by the container.
//
// Ex:
//  var x int
//  Build(Provide(func() int { return 1 }), &x)
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
