package statesim

import "io"

type AppState interface {
	GetModule(moduleName string) (ModuleState, bool)
	Modules(f func(moduleName string, modState ModuleState) bool)
}

func CompareAppStates(expected, actual AppState, diffWriter io.Writer) bool {
	panic("unimplemented")
}
