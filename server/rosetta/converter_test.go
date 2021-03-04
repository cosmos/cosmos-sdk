package rosetta

import (
	"encoding/hex"
	"testing"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"
	"github.com/stretchr/testify/suite"
	crgerrs "github.com/tendermint/cosmos-rosetta-gateway/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type ConverterTestSuite struct {
	suite.Suite

	c Converter
}

func (s *ConverterTestSuite) SetupTest() {
	cdc, ir := MakeCodec()
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	s.c = NewConverter(cdc, ir, txConfig)
}

func (s *ConverterTestSuite) TestFromRosettaOpsToTxSuccess() {
	addr1 := sdk.AccAddress("address1").String()
	addr2 := sdk.AccAddress("address2").String()

	msg1 := &bank.MsgSend{
		FromAddress: addr1,
		ToAddress:   addr2,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("test", 10)),
	}

	msg2 := &bank.MsgSend{
		FromAddress: addr2,
		ToAddress:   addr1,
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("utxo", 10)),
	}

	ops, err := s.c.ToRosetta().Ops("", msg1)
	s.Require().NoError(err)

	ops2, err := s.c.ToRosetta().Ops("", msg2)
	s.Require().NoError(err)

	ops = append(ops, ops2...)

	tx, err := s.c.ToSDK().UnsignedTx(ops)
	s.Require().NoError(err)

	getMsgs := tx.GetMsgs()

	s.Require().Equal(2, len(getMsgs))

	s.Require().Equal(getMsgs[0], msg1)
	s.Require().Equal(getMsgs[1], msg2)

}

func (s *ConverterTestSuite) TestFromRosettaOpsToTxErrors() {
	s.Run("unrecognized op", func() {
		op := &rosettatypes.Operation{
			Type: "non-existent",
		}

		_, err := s.c.ToSDK().UnsignedTx([]*rosettatypes.Operation{op})

		s.Require().ErrorIs(err, crgerrs.ErrBadArgument)
	})

	s.Run("codec type but not sdk.Msg", func() {
		op := &rosettatypes.Operation{
			Type: "cosmos.crypto.ed25519.PubKey",
		}

		_, err := s.c.ToSDK().UnsignedTx([]*rosettatypes.Operation{op})

		s.Require().ErrorIs(err, crgerrs.ErrBadArgument)

	})

}

func (s *ConverterTestSuite) TestMsgToMetaMetaToMsg() {
	msg := &bank.MsgSend{
		FromAddress: "addr1",
		ToAddress:   "addr2",
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("test", 10)),
	}

	msg.Route()

	meta, err := s.c.ToRosetta().Meta(msg)
	s.Require().NoError(err)

	copyMsg := new(bank.MsgSend)

	err = s.c.ToSDK().Msg(meta, copyMsg)
	s.Require().NoError(err)

	s.Require().Equal(msg, copyMsg)
}

func TestConverterTestSuite(t *testing.T) {
	suite.Run(t, new(ConverterTestSuite))
}

func (s *ConverterTestSuite) TestX() {
	const txRaw = "0a8e010a8b010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64126b0a2d636f736d6f7331656e377a6574686b6c6c79307761386a7778777878727638396565386a383668656374747337122d636f736d6f73317377383670393076393076753875706d363478327173373068756663796330746d34766e37361a0b0a057374616b651202313812600a4c0a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a21030c65f93f08cc27ee461ce1cd0a11647f0cfe852d044874d4d0905240a9787ec412020a0012100a0a0a057374616b651201311090a10f1a00"

	raw, _ := hex.DecodeString(txRaw)
	cdc, _ := MakeCodec()
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)

	tx, err := txConfig.TxDecoder()(raw)
	s.Require().NoError(err)
	txB, err := txConfig.TxJSONEncoder()(tx)
	s.Require().NoError(err)
	s.T().Logf("%s", txB)
}
