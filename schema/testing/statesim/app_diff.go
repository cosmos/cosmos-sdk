package statesim

import "fmt"

// AppState defines an interface for things that represent application state in schema format.
type AppState interface {
	// GetModule returns the module state for the given module name.
	GetModule(moduleName string) (ModuleState, bool)

	// Modules iterates over all the module state instances in the app.
	Modules(f func(moduleName string, modState ModuleState) bool)

	// NumModules returns the number of modules in the app.
	NumModules() int
}

// DiffAppStates compares the app state of two objects that implement AppState and returns a string with a diff if they
// are different or the empty string if they are the same.
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
