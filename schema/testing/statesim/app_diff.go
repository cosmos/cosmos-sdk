package statesim

import (
	"fmt"

	"cosmossdk.io/schema/view"
)

// DiffAppStates compares the app state of two objects that implement AppState and returns a string with a diff if they
// are different or the empty string if they are the same.
func DiffAppStates(expected, actual view.AppState) string {
	res := ""

	if expected.NumModules() != actual.NumModules() {
		res += fmt.Sprintf("MODULE COUNT ERROR: expected %d, got %d\n", expected.NumModules(), actual.NumModules())
	}

	expected.Modules(func(moduleName string, expectedMod view.ModuleState) bool {
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
