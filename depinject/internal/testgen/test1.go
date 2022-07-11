//go:build depinject

package testgen

var appConfig = ScenarioConfig

//depinject:appConfig
func Build(a ModuleA, b ModuleB) ([]Handler, error) {
	panic("depinject")
}
