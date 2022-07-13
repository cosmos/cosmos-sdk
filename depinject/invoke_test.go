package depinject_test

import (
	"testing"

	"github.com/regen-network/gocuke"
	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
)

func TestInvoke(t *testing.T) {
	gocuke.NewRunner(t, &invokeSuite{}).
		Path("features/invoke.feature").
		Run()
}

type invokeSuite struct {
	gocuke.TestingT
	configs []depinject.Config
	i       int
	sp      *string
}

func (s *invokeSuite) AnInvokerRequestingAnIntAndStringPointer() {
	s.configs = append(s.configs, depinject.Invoke(s.intStringPointerInvoker))
}

func (s *invokeSuite) intStringPointerInvoker(i int, sp *string) {
	s.i = i
	s.sp = sp
}

func (s *invokeSuite) TheContainerIsBuilt() {
	assert.NilError(s, depinject.Inject(depinject.Configs(s.configs...)))
}

func (s *invokeSuite) TheInvokerWillGetTheIntParameterSetTo(a int64) {
	assert.Equal(s, int(a), s.i)
}

func (s *invokeSuite) TheInvokerWillGetTheStringPointerParameterSetToNil() {
	if s.sp != nil {
		s.Fatalf("expected a nil string pointer, got %s", *s.sp)
	}
}

func (s *invokeSuite) AnIntProviderReturning(a int64) {
	s.configs = append(s.configs, depinject.Provide(func() int { return int(a) }))
}

func (s *invokeSuite) AStringPointerProviderPointingTo(a string) {
	s.configs = append(s.configs, depinject.Provide(func() *string { return &a }))
}

func (s *invokeSuite) TheInvokerWillGetTheStringPointerParameterSetTo(a string) {
	if s.sp == nil {
		s.Fatalf("expected a non-nil string pointer")
	}
	assert.Equal(s, a, *s.sp)
}

func (s *invokeSuite) AnInvokerRequestingAnIntAndStringPointerRunInModule(a string) {
	s.configs = append(s.configs, depinject.InvokeInModule(a, s.intStringPointerInvoker))
}

func (s *invokeSuite) AModulescopedIntProviderWhichReturnsTheLengthOfTheModuleName() {
	s.configs = append(s.configs, depinject.Provide(func(key depinject.ModuleKey) int {
		return len(key.Name())
	}))
}
