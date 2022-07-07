package appconfig_test

import (
	"bytes"
	"reflect"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/internal"
	"cosmossdk.io/core/internal/testpb"
	_ "cosmossdk.io/core/internal/testpb"
	"github.com/cosmos/cosmos-sdk/depinject"
)

func expectContainerErrorContains(t *testing.T, option depinject.Config, contains string) {
	t.Helper()
	err := depinject.Inject(option)
	assert.ErrorContains(t, err, contains)
}

func TestCompose(t *testing.T) {
	opt := appconfig.LoadJSON([]byte(`{"modules":[{}]}`))
	expectContainerErrorContains(t, opt, "module is missing name")

	opt = appconfig.LoadJSON([]byte(`{"modules":[{"name": "a"}]}`))
	expectContainerErrorContains(t, opt, `module "a" is missing a config object`)

	opt = appconfig.LoadYAML([]byte(`
modules:
- name: a
  config:
    "@type": testpb.ModuleFoo
`))
	expectContainerErrorContains(t, opt, `unable to resolve`)

	opt = appconfig.LoadYAML([]byte(`
modules:
- name: a
  config:
    "@type": cosmos.app.v1alpha1.Config # this is not actually a module config type!
`))
	expectContainerErrorContains(t, opt, "does not have the option cosmos.app.v1alpha1.module")
	expectContainerErrorContains(t, opt, "registered modules are")
	expectContainerErrorContains(t, opt, "testpb.TestModuleA")

	opt = appconfig.LoadYAML([]byte(`
modules:
- name: a
  config:
    "@type": testpb.TestUnregisteredModule
`))
	expectContainerErrorContains(t, opt, "did you forget to import cosmossdk.io/core/internal/testpb")
	expectContainerErrorContains(t, opt, "registered modules are")
	expectContainerErrorContains(t, opt, "testpb.TestModuleA")

	var app testpb.App
	opt = appconfig.LoadYAML([]byte(`
modules:
- name: runtime
  config:
   "@type": testpb.TestRuntimeModule
- name: a
  config:
   "@type": testpb.TestModuleA
- name: b
  config:
   "@type": testpb.TestModuleB
`))
	assert.NilError(t, depinject.Inject(opt, &app))
	buf := &bytes.Buffer{}
	app(buf)
	const expected = `got store key a
got store key b
running module handler a
result: hello
running module handler b
result: goodbye
`
	assert.Equal(t, expected, buf.String())

	opt = appconfig.LoadYAML([]byte(`
golang_bindings:
  - interfaceType: interfaceType/package.name 
    implementation: implementationType/package.name
  - interfaceType: interfaceType/package.nameTwo 
    implementation: implementationType/package.nameTwo
modules:
  - name: a
    config:
      "@type": testpb.TestModuleA
    golang_bindings:
      - interfaceType: interfaceType/package.name 
        implementation: implementationType/package.name
      - interfaceType: interfaceType/package.nameTwo 
        implementation: implementationType/package.nameTwo
`))
	assert.NilError(t, depinject.Inject(opt))

	// module registration failures:
	appmodule.Register(&testpb.TestNoModuleOptionModule{})
	opt = appconfig.LoadYAML([]byte(`
modules:
- name: a
  config:
   "@type": testpb.TestNoGoImportModule
`))
	expectContainerErrorContains(t, opt, "module should have the option cosmos.app.v1alpha1.module")

	internal.ModuleRegistry = map[reflect.Type]*internal.ModuleInitializer{} // reset module registry
	appmodule.Register(&testpb.TestNoGoImportModule{})
	opt = appconfig.LoadYAML([]byte(`
modules:
- name: a
  config:
   "@type": testpb.TestNoGoImportModule
`))
	expectContainerErrorContains(t, opt, "module should have ModuleDescriptor.go_import specified")

}
