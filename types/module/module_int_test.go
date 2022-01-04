package module

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestModuleIntSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

type TestSuite struct {
	suite.Suite
}

func (s TestSuite) TestAssertNoForgottenModules() {
	m := Manager{
		Modules: map[string]AppModule{"a": nil, "b": nil},
	}
	tcs := []struct {
		name     string
		positive bool
		modules  []string
	}{
		{"same modules", true, []string{"a", "b"}},
		{"more modules", true, []string{"a", "b", "c"}},
	}

	for _, tc := range tcs {
		if tc.positive {
			m.assertNoForgottenModules("x", tc.modules)
		} else {
			s.Panics(func() { m.assertNoForgottenModules("x", tc.modules) })
		}
	}
}
