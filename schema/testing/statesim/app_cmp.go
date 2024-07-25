package statesim

import "fmt"

type AppState interface {
	GetModule(moduleName string) (ModuleState, bool)
	Modules(f func(moduleName string, modState ModuleState) bool)
	NumModules() int
}

func DiffAppStates(expected, actual AppState) string {
	res := ""

	if expected.NumModules() != actual.NumModules() {
		res += fmt.Sprintf("MODULE COUNT ERROR: expected %d, got %d\n", expected.NumModules(), actual.NumModules())
	}

	expected.Modules(func(moduleName string, expectedMod ModuleState) bool {
		actualMod, found := actual.GetModule(moduleName)
		if !found {
			res += fmt.Sprintf("Module %s: NOT FOUND\n", moduleName)
			return true
		}

		diff := DiffModuleStates(expectedMod, actualMod)
		if diff != "" {
			res += "Module " + moduleName + "\n"
			res += indentAllLines(diff)
		}

		return true
	})

	return res
}
