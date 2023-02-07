package depinject_test

import (
	"testing"

	"github.com/regen-network/gocuke"
	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
)

func TestInvoke(t *testing.T) {
	gocuke.NewRunner(t, &InvokeSuite{}).
		Path("features/invoke.feature").
		Step("an int provider returning 5", (*InvokeSuite).AnIntProviderReturning5).
		Step(`a string pointer provider pointing to "foo"`, (*InvokeSuite).AStringPointerProviderPointingToFoo).
		Run()
}

type InvokeSuite struct {
	gocuke.TestingT
	configs []depinject.Config
	i       int
	sp      *string
}

func (s *InvokeSuite) AnInvokerRequestingAnIntAndStringPointer() {
	s.configs = append(s.configs,
		depinject.Supply(s),
		depinject.Invoke((*InvokeSuite).IntStringPointerInvoker),
	)
}

func (s *InvokeSuite) IntStringPointerInvoker(i int, sp *string) {
	s.i = i
	s.sp = sp
}

func (s *InvokeSuite) TheContainerIsBuilt() {
	assert.NilError(s, depinject.Inject(depinject.Configs(s.configs...)))
}

func (s *InvokeSuite) TheInvokerWillGetTheIntParameterSetTo(a int64) {
	assert.Equal(s, int(a), s.i)
}

func (s *InvokeSuite) TheInvokerWillGetTheStringPointerParameterSetToNil() {
	if s.sp != nil {
		s.Fatalf("expected a nil string pointer, got %s", *s.sp)
	}
}

func IntProvider5() int { return 5 }

func (s *InvokeSuite) AnIntProviderReturning5() {
	s.configs = append(s.configs, depinject.Provide(IntProvider5))
}

func StringPtrProviderFoo() *string {
	x := "foo"
	return &x
}

func (s *InvokeSuite) AStringPointerProviderPointingToFoo() {
	s.configs = append(s.configs, depinject.Provide(StringPtrProviderFoo))
}

func (s *InvokeSuite) TheInvokerWillGetTheStringPointerParameterSetTo(a string) {
	if s.sp == nil {
		s.Fatalf("expected a non-nil string pointer")
	}
	assert.Equal(s, a, *s.sp)
}

func (s *InvokeSuite) AnInvokerRequestingAnIntAndStringPointerRunInModule(a string) {
	s.configs = append(s.configs,
		depinject.Supply(s),
		depinject.InvokeInModule(a, (*InvokeSuite).IntStringPointerInvoker),
	)
}

func ProvideLenModuleKey(key depinject.ModuleKey) int {
	return len(key.Name())
}

func (s *InvokeSuite) AModulescopedIntProviderWhichReturnsTheLengthOfTheModuleName() {
	s.configs = append(s.configs, depinject.Provide(ProvideLenModuleKey))
}
