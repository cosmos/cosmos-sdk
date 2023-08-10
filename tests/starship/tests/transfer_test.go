package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/math"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestChainTokenTransfer() {
	txConfig := s.cdc.TxConfig
	txBuilder := txConfig.NewTxBuilder()

	// create a new address, and send it some tokens from faucet
	priv1, _, addr1 := testdata.KeyTestPubAddr()
	_, _, addr2 := testdata.KeyTestPubAddr()
	err := CreditFromFaucet(s.config, addr1.String())
	s.Require().NoError(err)

	s.T().Run("query balance for addr1", func(t *testing.T) {
		balance, err := banktypes.NewQueryClient(s.grpcConn).Balance(context.Background(), &banktypes.QueryBalanceRequest{
			Address: addr1.String(),
			Denom:   denom,
		})
		s.Require().NoError(err)
		s.Require().Equal(int64(10000000000), balance.Balance.Amount.Int64())
	})

	s.T().Run("send tokens from addr1 to addr2", func(t *testing.T) {
		// get account number and sequence
		accNum, seq, err := GetAccSeqNumber(s.grpcConn, addr1.String())
		s.Require().NoError(err)

		// build tx into the txBuilder
		msg := banktypes.NewMsgSend(addr1, addr2, sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(1230000))))
		s.Require().NoError(err)
		err = txBuilder.SetMsgs(msg)
		s.Require().NoError(err)
		txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(denom, math.NewInt(200000))))
		txBuilder.SetGasLimit(200000)
		txBuilder.SetTimeoutHeight(100000)

		// sign txn
		_, txBytes, err := CreateTestTx(txConfig, txBuilder, []cryptotypes.PrivKey{priv1}, []uint64{accNum}, []uint64{seq}, chainID)
		s.Require().NoError(err)

		// broadcast tx
		txClient := txtypes.NewServiceClient(s.grpcConn)
		res, err := txClient.BroadcastTx(context.Background(), &txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		})
		s.Require().NoError(err)
		s.Require().Equal(uint32(0), res.TxResponse.Code)
		s.WaitForTx(res.TxResponse.TxHash)
	})

	s.T().Run("query balance for addr2", func(t *testing.T) {
		balance, err := banktypes.NewQueryClient(s.grpcConn).Balance(context.Background(), &banktypes.QueryBalanceRequest{
			Address: addr2.String(),
			Denom:   denom,
		})
		s.Require().NoError(err)
		s.Require().Equal(int64(1230000), balance.Balance.Amount.Int64())
	})
}
