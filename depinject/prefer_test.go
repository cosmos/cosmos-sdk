package depinject_test

import (
	"testing"

	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/assert"

	"github.com/cosmos/cosmos-sdk/depinject"
)

func TestPrefer(t *testing.T) {
	gocuke.NewRunner(t, &preferSuite{}).
		Path("features/prefer.feature").
		Step(`we try to resolve a "Duck" in global scope`, (*preferSuite).WeTryToResolveADuckInGlobalScope).
		Step(`module (\w+) wants a "Duck"`, (*preferSuite).ModuleWantsADuck).
		Run()
}

type Duck interface {
	quack()
}

type Mallard struct{}
type Canvasback struct{}
type Marbled struct{}

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

type preferSuite struct {
	gocuke.TestingT // this gets injected by gocuke

	configs []depinject.Config
	pond    *Pond
	err     error
}

func (s preferSuite) AnInterfaceDuck() {
	// we don't need to do anything because this is defined at the type level
}

func (s preferSuite) TwoImplementationsMallardAndCanvasback() {
	// we don't need to do anything because this is defined at the type level
}

func (s *preferSuite) IsProvided(a string) {
	switch a {
	case "Mallard":
		s.addConfig(depinject.Provide(func() Mallard { return Mallard{} }))
	case "Canvasback":
		s.addConfig(depinject.Provide(func() Canvasback { return Canvasback{} }))
	case "Marbled":
		s.addConfig(depinject.Provide(func() Marbled { return Marbled{} }))
	default:
		s.Fatalf("unexpected duck type %s", a)
	}
}

func (s *preferSuite) addConfig(config depinject.Config) {
	s.configs = append(s.configs, config)
}

func (s *preferSuite) WeTryToResolveADuckInGlobalScope() {
	s.addConfig(depinject.Provide(func(duck Duck) DuckWrapper {
		return DuckWrapper{Module: "", Duck: duck}
	}))
}

func (s *preferSuite) resolvePond() *Pond {
	if s.pond != nil {
		return s.pond
	}

	s.err = depinject.Inject(depinject.Configs(s.configs...), s.pond)
	return s.pond
}

func (s *preferSuite) IsResolvedInGlobalScope(a string) {
	pond := s.resolvePond()
	for _, _ = range pond.Ducks {
		// TODO check that the duck with no module name (global)
		// is the expected type of duck
	}
}

func (s *preferSuite) ThereIsAError(a string) {
	assert.ErrorContains(s, s.err, a)
}

func (s *preferSuite) ThereIsAGlobalPreferenceForA(a string, b string) {
}

func (s *preferSuite) ThereIsAPreferenceForAInModule(a string, b string, c string) {
}

func (s *preferSuite) ModuleWantsADuck(module string) {
}

func (s *preferSuite) ModuleResolvesA(module string, duckType string) {
	pond := s.resolvePond()
	for _, _ = range pond.Ducks {
		// TODO check that the duck with module name (global)
		// is the expected type of duck
	}
}
