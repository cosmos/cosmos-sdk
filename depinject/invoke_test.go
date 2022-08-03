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
		Step("an int provider returning 5", (*invokeSuite).AnIntProviderReturning5).
		Step(`a string pointer provider pointing to "foo"`, (*invokeSuite).AStringPointerProviderPointingToFoo).
		Run()
}

type invokeSuite struct {
	gocuke.TestingT
	configs []depinject.Config
	i       int
	sp      *string
}

func (s *invokeSuite) AnInvokerRequestingAnIntAndStringPointer() {
	s.configs = append(s.configs,
		depinject.Supply(s),
		depinject.Invoke((*invokeSuite).IntStringPointerInvoker),
	)
}

func (s *invokeSuite) IntStringPointerInvoker(i int, sp *string) {
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

func IntProvider5() int { return 5 }

func (s *invokeSuite) AnIntProviderReturning5() {
	s.configs = append(s.configs, depinject.Provide(IntProvider5))
}

func StringPtrProviderFoo() *string {
	x := "foo"
	return &x
}

func (s *invokeSuite) AStringPointerProviderPointingToFoo() {
	s.configs = append(s.configs, depinject.Provide(StringPtrProviderFoo))
}

func (s *invokeSuite) TheInvokerWillGetTheStringPointerParameterSetTo(a string) {
	if s.sp == nil {
		s.Fatalf("expected a non-nil string pointer")
	}
	assert.Equal(s, a, *s.sp)
}

func (s *invokeSuite) AnInvokerRequestingAnIntAndStringPointerRunInModule(a string) {
	s.configs = append(s.configs,
		depinject.Supply(s),
		depinject.InvokeInModule(a, (*invokeSuite).IntStringPointerInvoker),
	)
}

func ProvideLenModuleKey(key depinject.ModuleKey) int {
	return len(key.Name())
}

func (s *invokeSuite) AModulescopedIntProviderWhichReturnsTheLengthOfTheModuleName() {
	s.configs = append(s.configs, depinject.Provide(ProvideLenModuleKey))
}
