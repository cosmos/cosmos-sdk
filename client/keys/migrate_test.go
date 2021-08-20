package keys

import (
	"context"
	"fmt"
	"strings"

	design99keyring "github.com/99designs/keyring"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/codec"
)

type setter interface {
	SetItem(item design99keyring.Item) error
}

type MigrateTestSuite struct {
	suite.Suite

	dir  string
	n1   string
	cdc codec.Codec
	kb   keyring.Keyring
	priv cryptotypes.PrivKey
	pub  cryptotypes.PubKey
	setter setter
}

func (s *MigrateTestSuite) SetupSuite() {
	s.dir = s.T().TempDir()
	mockIn := strings.NewReader("")
	s.cdc = simapp.MakeTestEncodingConfig().Codec
	s.n1 = "cosmos"

	kb, err := keyring.New(s.n1, keyring.BackendTest, s.dir, mockIn, s.cdc)
	s.Require().NoError(err)
	s.kb = kb

	setter, ok := kb.(setter)
	s.Require().True(ok)
	s.setter = setter

	s.priv = cryptotypes.PrivKey(secp256k1.GenPrivKey())
	s.pub = s.priv.PubKey()
}

func (s *MigrateTestSuite) Test_runMigrateCmdLegacyInfo() {

	multi := multisig.NewLegacyAminoPubKey(
		1, []cryptotypes.PubKey{
			s.pub,
		},
	)
	legacyMultiInfo, err := keyring.NewLegacyMultiInfo(s.n1, multi)
	s.Require().NoError(err)
	serializedLegacyMultiInfo := keyring.MarshalInfo(legacyMultiInfo)

	// adding LegacyInfo item into keyring
	item := design99keyring.Item{
		Key:         s.n1,
		Data:        serializedLegacyMultiInfo,
		Description: "SDK kerying version",
	}

	s.Require().NoError(s.setter.SetItem(item))

	clientCtx := client.Context{}.WithKeyringDir(s.dir).WithKeyring(s.kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	// run MigrateCommand, it should return no error
	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn2 := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn2, mockOut := testutil.ApplyMockIO(cmd)

	mockIn2.Reset("\n12345678\n\n\n\n\n")
	s.T().Log(mockOut.String())
	s.Assert().NoError(cmd.ExecuteContext(ctx))
}

func (s *MigrateTestSuite) Test_runMigrateCmdRecord() {
	k, err := keyring.NewLocalRecord("test record", s.priv, s.pub)
	s.Require().NoError(err)
	serializedRecord, err := s.cdc.Marshal(k)
	s.Require().NoError(err)

	item := design99keyring.Item{
		Key:         s.n1,
		Data:        serializedRecord,
		Description: "SDK kerying version",
	}

	s.Require().NoError(s.setter.SetItem(item))

	clientCtx := client.Context{}.WithKeyringDir(s.dir).WithKeyring(s.kb)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	// run MigrateCommand, it should return no error
	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	mockIn2 := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn2, mockOut := testutil.ApplyMockIO(cmd)

	mockIn2.Reset("\n12345678\n\n\n\n\n")
	s.T().Log(mockOut.String())
	s.Assert().NoError(cmd.ExecuteContext(ctx))
}

func (s *MigrateTestSuite) Test_runMigrateCmdNoKeys() {
	clientCtx := client.Context{}.WithKeyringDir(s.dir)
	ctx := context.WithValue(context.Background(), client.ClientContextKey, &clientCtx)

	cmd := MigrateCommand()
	cmd.Flags().AddFlagSet(Commands("home").PersistentFlags())
	cmd.SetArgs([]string{fmt.Sprintf("--%s=%s", flags.FlagKeyringBackend, keyring.BackendTest)})

	mockIn := testutil.ApplyMockIODiscardOutErr(cmd)
	mockIn, mockOut := testutil.ApplyMockIO(cmd)

	mockIn.Reset("\n12345678\n\n\n\n\n")
	s.T().Log(mockOut.String())
	s.Assert().NoError(cmd.ExecuteContext(ctx))
}
