package keeper_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/KiraCore/cosmos-sdk/simapp"
	store "github.com/KiraCore/cosmos-sdk/store/types"
	"github.com/KiraCore/cosmos-sdk/x/upgrade/keeper"
	"github.com/KiraCore/cosmos-sdk/x/upgrade/types"
)

type KeeperTestSuite struct {
	suite.Suite

	homeDir string
	app     *simapp.SimApp
}

func (s *KeeperTestSuite) SetupTest() {
	app := simapp.Setup(false)

	homeDir, err := ioutil.TempDir(os.TempDir(), "x_upgrade_keeper_test")
	s.Require().NoError(err)

	// recreate keeper in order to use a custom home path
	app.UpgradeKeeper = keeper.NewKeeper(
		make(map[int64]bool), app.GetKey(types.StoreKey), app.AppCodec(), homeDir,
	)

	s.homeDir = homeDir
	s.app = app
}

func (s *KeeperTestSuite) TestReadUpgradeInfoFromDisk() {
	// require no error when the upgrade info file does not exist
	_, err := s.app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	s.Require().NoError(err)

	expected := store.UpgradeInfo{
		Name:   "test_upgrade",
		Height: 100,
	}

	// create an upgrade info file
	s.Require().NoError(s.app.UpgradeKeeper.DumpUpgradeInfoToDisk(expected.Height, expected.Name))

	ui, err := s.app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	s.Require().NoError(err)
	s.Require().Equal(expected, ui)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
