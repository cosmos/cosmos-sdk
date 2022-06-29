package sims

import (
	"fmt"
	"math/rand"
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsign "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// GenSignedMockTx generates a signed mock transaction.
func GenSignedMockTx(txConfig client.TxConfig, msgs []sdk.Msg, feeAmt sdk.Coins, gas uint64, chainID string, accNums, accSeqs []uint64, priv ...cryptotypes.PrivKey) (sdk.Tx, error) {
	sigs := make([]signing.SignatureV2, len(priv))

	// create a random length memo
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	memo := simulation.RandStringOfLength(r, simulation.RandIntBetween(r, 0, 100))

	signMode := txConfig.SignModeHandler().DefaultMode()

	// 1st round: set SignatureV2 with empty signatures, to set correct
	// signer infos.
	for i, p := range priv {
		sigs[i] = signing.SignatureV2{
			PubKey: p.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
			Sequence: accSeqs[i],
		}
	}

	tx := txConfig.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}
	err = tx.SetSignatures(sigs...)
	if err != nil {
		return nil, err
	}
	tx.SetMemo(memo)
	tx.SetFeeAmount(feeAmt)
	tx.SetGasLimit(gas)

	// 2nd round: once all signer infos are set, every signer can sign.
	for i, p := range priv {
		signerData := authsign.SignerData{
			Address:       sdk.AccAddress(p.PubKey().Address()).String(),
			ChainID:       chainID,
			AccountNumber: accNums[i],
			Sequence:      accSeqs[i],
			PubKey:        p.PubKey(),
		}
		signBytes, err := txConfig.SignModeHandler().GetSignBytes(signMode, signerData, tx.GetTx())
		if err != nil {
			panic(err)
		}
		sig, err := p.Sign(signBytes)
		if err != nil {
			panic(err)
		}
		sigs[i].Data.(*signing.SingleSignatureData).Signature = sig
		err = tx.SetSignatures(sigs...)
		if err != nil {
			panic(err)
		}
	}

	return tx.GetTx(), nil
}

// SignCheckDeliver checks a generated signed transaction and simulates a
// block commitment with the given transaction. A test assertion is made using
// the parameter 'expPass' against the result. A corresponding result is
// returned.
func SignCheckDeliver(txConfig client.TxConfig, ba *baseapp.BaseApp, header tmproto.Header, msgs []sdk.Msg,
	chainID string, accNums, accSeqs []uint64, expSimPass, expPass bool, priv ...cryptotypes.PrivKey,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := GenSignedMockTx(
		txConfig,
		msgs,
		sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
		DefaultGenTxGas,
		chainID,
		accNums,
		accSeqs,
		priv...,
	)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	txBytes, err := txConfig.TxEncoder()(tx)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	// Must simulate now as CheckTx doesn't run Msgs anymore
	_, res, err := ba.Simulate(txBytes)
	if expSimPass {
		if err != nil {
			return sdk.GasInfo{}, nil, err
		}

		if res == nil {
			return sdk.GasInfo{}, nil, fmt.Errorf("Simulate() returned no result")
		}
	} else {
		if err == nil {
			return sdk.GasInfo{}, nil, fmt.Errorf("Simulate() passed but should have failed")
		}

		if res != nil {
			return sdk.GasInfo{}, nil, fmt.Errorf("Simulate() returned a result")
		}
	}

	// Simulate a sending a transaction and committing a block
	ba.BeginBlock(abci.RequestBeginBlock{Header: header})
	gInfo, res, err := ba.SimDeliver(txConfig.TxEncoder(), tx)
	if expPass {
		if err != nil {
			return gInfo, nil, err
		}

		if res == nil {
			return gInfo, nil, fmt.Errorf("SimDeliver() returned no result")
		}
	} else {
		if err == nil {
			return gInfo, nil, fmt.Errorf("SimDeliver() passed but should have failed")
		}

		if res != nil {
			return gInfo, nil, fmt.Errorf("SimDeliver() returned a result")
		}
	}

	ba.EndBlock(abci.RequestEndBlock{})
	ba.Commit()

	return gInfo, res, err
}
