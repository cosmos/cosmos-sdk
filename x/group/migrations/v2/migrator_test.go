package v2_test

import (
	"testing"

	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
)

func TestMigrate(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(groupmodule.AppModuleBasic{})
	_ = encCfg.Codec

}
