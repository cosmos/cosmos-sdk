package rosetta

import (
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

	ops, err := s.c.ToRosetta().MsgToOps("", msg1)
	s.Require().NoError(err)

	ops2, err := s.c.ToRosetta().MsgToOps("", msg2)
	s.Require().NoError(err)

	ops = append(ops, ops2...)

	tx, err := s.c.FromRosetta().OpsToUnsignedTx(ops)
	s.Require().NoError(err)

	msgs := tx.GetMsgs()

	s.Require().Equal(2, len(msgs))

	s.Require().Equal(msgs[0], msg1)
	s.Require().Equal(msgs[1], msg2)

}

func (s *ConverterTestSuite) TestFromRosettaOpsToTxErrors() {
	s.Run("unrecognized op", func() {
		op := &rosettatypes.Operation{
			Type: "non-existent",
		}

		_, err := s.c.FromRosetta().OpsToUnsignedTx([]*rosettatypes.Operation{op})

		s.Require().ErrorIs(err, crgerrs.ErrBadArgument)
	})

	s.Run("codec type but not sdk.Msg", func() {
		op := &rosettatypes.Operation{
			Type: "cosmos.crypto.ed25519.PubKey",
		}

		_, err := s.c.FromRosetta().OpsToUnsignedTx([]*rosettatypes.Operation{op})

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

	meta, err := s.c.ToRosetta().MsgToMeta(msg)
	s.Require().NoError(err)

	copyMsg := new(bank.MsgSend)

	err = s.c.FromRosetta().MetaToMsg(meta, copyMsg)
	s.Require().NoError(err)

	s.Require().Equal(msg, copyMsg)
}

func TestConverterTestSuite(t *testing.T) {
	suite.Run(t, new(ConverterTestSuite))
}
