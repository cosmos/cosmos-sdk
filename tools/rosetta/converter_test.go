package rosetta_test

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"cosmossdk.io/tools/rosetta"
	crgerrs "cosmossdk.io/tools/rosetta/lib/errors"

	rosettatypes "github.com/coinbase/rosetta-sdk-go/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type ConverterTestSuite struct {
	suite.Suite

	c               rosetta.Converter
	unsignedTxBytes []byte
	unsignedTx      authsigning.Tx

	ir     codectypes.InterfaceRegistry
	cdc    *codec.ProtoCodec
	txConf client.TxConfig
}

func (s *ConverterTestSuite) SetupTest() {
	// create an unsigned tx
	const unsignedTxHex = "0a8e010a8b010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64126b0a2d636f736d6f733134376b6c68377468356a6b6a793361616a736a3272717668747668396d666465333777713567122d636f736d6f73316d6e7670386c786b616679346c787777617175356561653764787630647a36687767797436331a0b0a057374616b651202313612600a4c0a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a21034c92046950c876f4a5cb6c7797d6eeb9ef80d67ced4d45fb62b1e859240ba9ad12020a0012100a0a0a057374616b651201311090a10f1a00"
	unsignedTxBytes, err := hex.DecodeString(unsignedTxHex)
	s.Require().NoError(err)
	s.unsignedTxBytes = unsignedTxBytes
	// instantiate converter
	cdc, ir := rosetta.MakeCodec()
	txConfig := authtx.NewTxConfig(cdc, authtx.DefaultSignModes)
	s.c = rosetta.NewConverter(cdc, ir, txConfig)
	// add utils
	s.ir = ir
	s.cdc = cdc
	s.txConf = txConfig
	// add authsigning tx
	sdkTx, err := txConfig.TxDecoder()(unsignedTxBytes)
	s.Require().NoError(err)
	builder, err := txConfig.WrapTxBuilder(sdkTx)
	s.Require().NoError(err)

	s.unsignedTx = builder.GetTx()
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

	meta, err := s.c.ToRosetta().Meta(msg)
	s.Require().NoError(err)

	copyMsg := new(bank.MsgSend)
	err = s.c.ToSDK().Msg(meta, copyMsg)
	s.Require().NoError(err)
	s.Require().Equal(msg, copyMsg)
}

func (s *ConverterTestSuite) TestSignedTx() {
	s.Run("success", func() {
		const payloadsJSON = `[{"hex_bytes":"82ccce81a3e4a7272249f0e25c3037a316ee2acce76eb0c25db00ef6634a4d57303b2420edfdb4c9a635ad8851fe5c7a9379b7bc2baadc7d74f7e76ac97459b5","signing_payload":{"address":"cosmos147klh7th5jkjy3aajsj2rqvhtvh9mfde37wq5g","hex_bytes":"ed574d84b095250280de38bf8c254e4a1f8755e5bd300b1f6ca2671688136ecc","account_identifier":{"address":"cosmos147klh7th5jkjy3aajsj2rqvhtvh9mfde37wq5g"},"signature_type":"ecdsa"},"public_key":{"hex_bytes":"034c92046950c876f4a5cb6c7797d6eeb9ef80d67ced4d45fb62b1e859240ba9ad","curve_type":"secp256k1"},"signature_type":"ecdsa"}]`
		const expectedSignedTxHex = "0a8e010a8b010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e64126b0a2d636f736d6f733134376b6c68377468356a6b6a793361616a736a3272717668747668396d666465333777713567122d636f736d6f73316d6e7670386c786b616679346c787777617175356561653764787630647a36687767797436331a0b0a057374616b651202313612620a4e0a460a1f2f636f736d6f732e63727970746f2e736563703235366b312e5075624b657912230a21034c92046950c876f4a5cb6c7797d6eeb9ef80d67ced4d45fb62b1e859240ba9ad12040a02087f12100a0a0a057374616b651201311090a10f1a4082ccce81a3e4a7272249f0e25c3037a316ee2acce76eb0c25db00ef6634a4d57303b2420edfdb4c9a635ad8851fe5c7a9379b7bc2baadc7d74f7e76ac97459b5"

		var payloads []*rosettatypes.Signature
		s.Require().NoError(json.Unmarshal([]byte(payloadsJSON), &payloads))

		signedTx, err := s.c.ToSDK().SignedTx(s.unsignedTxBytes, payloads)
		s.Require().NoError(err)

		signedTxHex := hex.EncodeToString(signedTx)

		s.Require().Equal(signedTxHex, expectedSignedTxHex)
	})

	s.Run("signers data and signing payloads mismatch", func() {
		_, err := s.c.ToSDK().SignedTx(s.unsignedTxBytes, nil)
		s.Require().ErrorIs(err, crgerrs.ErrInvalidTransaction)
	})
}

func (s *ConverterTestSuite) TestOpsAndSigners() {
	s.Run("success", func() {
		addr1 := sdk.AccAddress("address1").String()
		addr2 := sdk.AccAddress("address2").String()

		msg := &bank.MsgSend{
			FromAddress: addr1,
			ToAddress:   addr2,
			Amount:      sdk.NewCoins(sdk.NewInt64Coin("test", 10)),
		}

		builder := s.txConf.NewTxBuilder()
		s.Require().NoError(builder.SetMsgs(msg))

		sdkTx := builder.GetTx()
		txBytes, err := s.txConf.TxEncoder()(sdkTx)
		s.Require().NoError(err)

		ops, signers, err := s.c.ToRosetta().OpsAndSigners(txBytes)
		s.Require().NoError(err)

		s.Require().Equal(len(ops), len(sdkTx.GetMsgs())*len(sdkTx.GetSigners()), "operation number mismatch")

		s.Require().Equal(len(signers), len(sdkTx.GetSigners()), "signers number mismatch")
	})
}

func (s *ConverterTestSuite) TestBeginEndBlockAndHashToTxType() {
	const deliverTxHex = "5229A67AA008B5C5F1A0AEA77D4DEBE146297A30AAEF01777AF10FAD62DD36AB"

	deliverTxBytes, err := hex.DecodeString(deliverTxHex)
	s.Require().NoError(err)

	endBlockTxHex := s.c.ToRosetta().EndBlockTxHash(deliverTxBytes)
	beginBlockTxHex := s.c.ToRosetta().BeginBlockTxHash(deliverTxBytes)

	txType, hash := s.c.ToSDK().HashToTxType(deliverTxBytes)

	s.Require().Equal(rosetta.DeliverTxTx, txType)
	s.Require().Equal(deliverTxBytes, hash, "deliver tx hash should not change")

	endBlockTxBytes, err := hex.DecodeString(endBlockTxHex)
	s.Require().NoError(err)

	txType, hash = s.c.ToSDK().HashToTxType(endBlockTxBytes)

	s.Require().Equal(rosetta.EndBlockTx, txType)
	s.Require().Equal(deliverTxBytes, hash, "end block tx hash should be equal to a block hash")

	beginBlockTxBytes, err := hex.DecodeString(beginBlockTxHex)
	s.Require().NoError(err)

	txType, hash = s.c.ToSDK().HashToTxType(beginBlockTxBytes)

	s.Require().Equal(rosetta.BeginBlockTx, txType)
	s.Require().Equal(deliverTxBytes, hash, "begin block tx hash should be equal to a block hash")

	txType, hash = s.c.ToSDK().HashToTxType([]byte("invalid"))

	s.Require().Equal(rosetta.UnrecognizedTx, txType)
	s.Require().Nil(hash)

	txType, hash = s.c.ToSDK().HashToTxType(append([]byte{0x3}, deliverTxBytes...))
	s.Require().Equal(rosetta.UnrecognizedTx, txType)
	s.Require().Nil(hash)
}

func (s *ConverterTestSuite) TestSigningComponents() {
	s.Run("invalid metadata coins", func() {
		_, _, err := s.c.ToRosetta().SigningComponents(nil, &rosetta.ConstructionMetadata{GasPrice: "invalid"}, nil)
		s.Require().ErrorIs(err, crgerrs.ErrBadArgument)
	})

	s.Run("length signers data does not match signers", func() {
		_, _, err := s.c.ToRosetta().SigningComponents(s.unsignedTx, &rosetta.ConstructionMetadata{GasPrice: "10stake"}, nil)
		s.Require().ErrorIs(err, crgerrs.ErrBadArgument)
	})

	s.Run("length pub keys does not match signers", func() {
		_, _, err := s.c.ToRosetta().SigningComponents(
			s.unsignedTx,
			&rosetta.ConstructionMetadata{GasPrice: "10stake", SignersData: []*rosetta.SignerData{
				{
					AccountNumber: 0,
					Sequence:      0,
				},
			}},
			nil)
		s.Require().ErrorIs(err, crgerrs.ErrBadArgument)
	})

	s.Run("ros pub key is valid but not the one we expect", func() {
		validButUnexpected, err := hex.DecodeString("030da9096a40eb1d6c25f1e26e9cbf8941fc84b8f4dc509c8df5e62a29ab8f2415")
		s.Require().NoError(err)

		_, _, err = s.c.ToRosetta().SigningComponents(
			s.unsignedTx,
			&rosetta.ConstructionMetadata{GasPrice: "10stake", SignersData: []*rosetta.SignerData{
				{
					AccountNumber: 0,
					Sequence:      0,
				},
			}},
			[]*rosettatypes.PublicKey{
				{
					Bytes:     validButUnexpected,
					CurveType: rosettatypes.Secp256k1,
				},
			})
		s.Require().ErrorIs(err, crgerrs.ErrBadArgument)
	})

	s.Run("success", func() {
		expectedPubKey, err := hex.DecodeString("034c92046950c876f4a5cb6c7797d6eeb9ef80d67ced4d45fb62b1e859240ba9ad")
		s.Require().NoError(err)

		_, _, err = s.c.ToRosetta().SigningComponents(
			s.unsignedTx,
			&rosetta.ConstructionMetadata{GasPrice: "10stake", SignersData: []*rosetta.SignerData{
				{
					AccountNumber: 0,
					Sequence:      0,
				},
			}},
			[]*rosettatypes.PublicKey{
				{
					Bytes:     expectedPubKey,
					CurveType: rosettatypes.Secp256k1,
				},
			})
		s.Require().NoError(err)
	})
}

func (s *ConverterTestSuite) TestBalanceOps() {
	s.Run("not a balance op", func() {
		notBalanceOp := abci.Event{
			Type: "not-a-balance-op",
		}

		ops := s.c.ToRosetta().BalanceOps("", []abci.Event{notBalanceOp})
		s.Len(ops, 0, "expected no balance ops")
	})

	s.Run("multiple balance ops from 2 multicoins event", func() {
		subBalanceOp := bank.NewCoinSpentEvent(
			sdk.AccAddress("test"),
			sdk.NewCoins(sdk.NewInt64Coin("test", 10), sdk.NewInt64Coin("utxo", 10)),
		)

		addBalanceOp := bank.NewCoinReceivedEvent(
			sdk.AccAddress("test"),
			sdk.NewCoins(sdk.NewInt64Coin("test", 10), sdk.NewInt64Coin("utxo", 10)),
		)

		ops := s.c.ToRosetta().BalanceOps("", []abci.Event{(abci.Event)(subBalanceOp), (abci.Event)(addBalanceOp)})
		s.Len(ops, 4)
	})

	s.Run("spec broken", func() {
		s.Require().Panics(func() {
			specBrokenSub := abci.Event{
				Type: bank.EventTypeCoinSpent,
			}
			_ = s.c.ToRosetta().BalanceOps("", []abci.Event{specBrokenSub})
		})

		s.Require().Panics(func() {
			specBrokenSub := abci.Event{
				Type: bank.EventTypeCoinBurn,
			}
			_ = s.c.ToRosetta().BalanceOps("", []abci.Event{specBrokenSub})
		})

		s.Require().Panics(func() {
			specBrokenSub := abci.Event{
				Type: bank.EventTypeCoinReceived,
			}
			_ = s.c.ToRosetta().BalanceOps("", []abci.Event{specBrokenSub})
		})
	})
}

func TestConverterTestSuite(t *testing.T) {
	suite.Run(t, new(ConverterTestSuite))
}
