package appconfig_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/container"

	appv1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/app/v1alpha1"
	modulev1 "github.com/cosmos/cosmos-sdk/api/cosmos/auth/module/v1"

	"cosmossdk.io/core/appconfig"
)

func expectContainerErrorContains(t *testing.T, option container.Option, contains string) {
	t.Helper()
	err := container.Run(func() {}, option)
	assert.ErrorContains(t, err, contains)
}

func TestComposeErrors(t *testing.T) {
	opt := appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{},
		},
	})
	expectContainerErrorContains(t, opt, "module is missing name")

	opt = appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: "auth",
			},
		},
	})
	expectContainerErrorContains(t, opt, "missing a config object")

	opt = appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name:   "auth",
				Config: appconfig.MustWrapAny(&appv1alpha1.ModuleConfig{}),
			},
		},
	})
	expectContainerErrorContains(t, opt, "does not have the option")

	opt = appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name:   "auth",
				Config: appconfig.MustWrapAny(&modulev1.Module{}),
			},
		},
	})
	expectContainerErrorContains(t, opt, "did you forget to import")
}
