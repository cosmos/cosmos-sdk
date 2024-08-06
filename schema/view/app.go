package view

type AppData interface {
	// AppState returns the app state.
	AppState() AppState

	// BlockNum returns the latest block number.
	BlockNum() uint64
}

// AppState defines an interface for things that represent application state in schema format.
type AppState interface {
	// GetModule returns the module state for the given module name.
	GetModule(moduleName string) (ModuleState, bool)

	// Modules iterates over all the module state instances in the app.
	Modules(f func(moduleName string, modState ModuleState) bool)

	// NumModules returns the number of modules in the app.
	NumModules() int
}
