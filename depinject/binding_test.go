package depinject_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/assert"

	"cosmossdk.io/depinject"
)

func TestBindInterface(t *testing.T) {
	gocuke.NewRunner(t, &bindingSuite{}).
		Path("features/bindings.feature").
		Step(`we try to resolve a "Duck" in global scope`, (*bindingSuite).WeTryToResolveADuckInGlobalScope).
		Step(`module "(\w+)" wants a "Duck"`, (*bindingSuite).ModuleWantsADuck).
		Run()
}

type Duck interface {
	quack()
}

type (
	Mallard    struct{}
	Canvasback struct{}
	Marbled    struct{}
)

func (duck Mallard) quack()    {}
func (duck Canvasback) quack() {}
func (duck Marbled) quack()    {}

type DuckWrapper struct {
	Module string
	Duck   Duck
}

func (d DuckWrapper) IsManyPerContainerType() {}

type Pond struct {
	Ducks []DuckWrapper
}

type bindingSuite struct {
	gocuke.TestingT // this gets injected by gocuke

	configs []depinject.Config
	pond    *Pond
	err     error
}

func (s bindingSuite) AnInterfaceDuck() {
	// we don't need to do anything because this is defined at the type level
}

func (s bindingSuite) TwoImplementationsMallardAndCanvasback() {
	// we don't need to do anything because this is defined at the type level
}

func ProvideMallard() Mallard       { return Mallard{} }
func ProvideCanvasback() Canvasback { return Canvasback{} }
func ProvideMarbled() Marbled       { return Marbled{} }

func (s *bindingSuite) IsProvided(a string) {
	switch a {
	case "Mallard":
		s.addConfig(depinject.Provide(ProvideMallard))
	case "Canvasback":
		s.addConfig(depinject.Provide(ProvideCanvasback))
	case "Marbled":
		s.addConfig(depinject.Provide(ProvideMarbled))
	default:
		s.Fatalf("unexpected duck type %s", a)
	}
}

func (s *bindingSuite) addConfig(config depinject.Config) {
	s.configs = append(s.configs, config)
}

func ProvideDuckWrapper(duck Duck) DuckWrapper {
	return DuckWrapper{Module: "", Duck: duck}
}

func (s *bindingSuite) WeTryToResolveADuckInGlobalScope() {
	s.addConfig(depinject.Provide(ProvideDuckWrapper))
}

func ResolvePond(ducks []DuckWrapper) Pond { return Pond{Ducks: ducks} }

func (s *bindingSuite) resolvePond() *Pond {
	if s.pond != nil {
		return s.pond
	}

	s.addConfig(depinject.Provide(ResolvePond))
	var pond Pond
	s.err = depinject.Inject(depinject.Configs(s.configs...), &pond)
	s.pond = &pond
	return s.pond
}

func (s *bindingSuite) IsResolvedInGlobalScope(typeName string) {
	pond := s.resolvePond()
	found := false
	for _, dw := range pond.Ducks {
		if dw.Module == "" {
			require.Contains(s, reflect.TypeOf(dw.Duck).Name(), typeName)
			found = true
		}
	}
	assert.True(s, found)
}

func (s *bindingSuite) ThereIsAError(expectedErrorMsg string) {
	s.resolvePond()
	assert.ErrorContains(s, s.err, expectedErrorMsg)
}

func (s *bindingSuite) ThereIsNoError() {
	s.resolvePond()
	assert.NoError(s, s.err)
}

func fullTypeName(typeName string) string {
	return fmt.Sprintf("cosmossdk.io/depinject_test/depinject_test.%s", typeName)
}

func (s *bindingSuite) ThereIsAGlobalBindingForA(preferredType, interfaceType string) {
	s.addConfig(depinject.BindInterface(fullTypeName(interfaceType), fullTypeName(preferredType)))
}

func (s *bindingSuite) ThereIsABindingForAInModule(preferredType, interfaceType, moduleName string) {
	s.addConfig(depinject.BindInterfaceInModule(moduleName, fullTypeName(interfaceType), fullTypeName(preferredType)))
}

func ProvideModuleDuck(duck Duck, key depinject.OwnModuleKey) DuckWrapper {
	return DuckWrapper{Module: depinject.ModuleKey(key).Name(), Duck: duck}
}

func (s *bindingSuite) ModuleWantsADuck(module string) {
	s.addConfig(depinject.ProvideInModule(module, ProvideModuleDuck))
}

func (s *bindingSuite) ModuleResolvesA(module, duckType string) {
	pond := s.resolvePond()
	moduleFound := false
	for _, dw := range pond.Ducks {
		if dw.Module == module {
			assert.Contains(s, reflect.TypeOf(dw.Duck).Name(), duckType)
			moduleFound = true
		}
	}
	assert.True(s, moduleFound)
}
