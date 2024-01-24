package tx

import (
	"fmt"

	authsign "cosmossdk.io/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

var (
	_ client.TxBuilder          = &builder{}
	_ ExtensionOptionsTxBuilder = &builder{}
)

type builder struct {
	msgs          []sdk.Msg
	timeoutHeight uint64
	granter       []byte
	payer         []byte
	unordered     bool
	memo          string
	gasLimit      uint64
	fees          sdk.Coins
	signerInfos   []*tx.SignerInfo
	signatures    [][]byte

	extOptions            []*codectypes.Any
	nonCriticalExtOptions []*codectypes.Any
	extensionOptions      []*codectypes.Any
	nonCriticalExtOption  []*codectypes.Any
}

func (w *builder) GetTx() authsign.Tx {
	tx, err := w.getTx()
	if err != nil {
		panic(err)
	}
	return tx
}

func (w *builder) getTx() (*gogoTxWrapper, error) {

}

func (w *builder) SetMsgs(msgs ...sdk.Msg) error {
	w.msgs = msgs
	return nil
}

// SetTimeoutHeight sets the transaction's height timeout.
func (w *builder) SetTimeoutHeight(height uint64) {
	w.timeoutHeight = height
}

func (w *builder) SetUnordered(v bool) {
	w.unordered = v
}

func (w *builder) SetMemo(memo string) {
	w.memo = memo
}

func (w *builder) SetGasLimit(limit uint64) {
	w.gasLimit = limit
}

func (w *builder) SetFeeAmount(coins sdk.Coins) {
	w.fees = coins
}

func (w *builder) SetFeePayer(feePayer sdk.AccAddress) {
	w.payer = feePayer
}

func (w *builder) SetFeeGranter(feeGranter sdk.AccAddress) {
	w.granter = feeGranter
}

func (w *builder) SetSignatures(signatures ...signing.SignatureV2) error {
	n := len(signatures)
	signerInfos := make([]*tx.SignerInfo, n)
	rawSigs := make([][]byte, n)

	for i, sig := range signatures {
		var (
			modeInfo *tx.ModeInfo
			pubKey   *codectypes.Any
			err      error
		)
		modeInfo, rawSigs[i] = SignatureDataToModeInfoAndSig(sig.Data)
		if sig.PubKey != nil {
			pubKey, err = codectypes.NewAnyWithValue(sig.PubKey)
			if err != nil {
				return err
			}
		}
		signerInfos[i] = &tx.SignerInfo{
			PublicKey: pubKey,
			ModeInfo:  modeInfo,
			Sequence:  sig.Sequence,
		}
	}

	w.setSignerInfos(signerInfos)
	w.setSignatures(rawSigs)

	return nil
}

func (w *builder) setSignerInfos(infos []*tx.SignerInfo) {
	w.signerInfos = infos
}

func (w *builder) setSignatures(sigs [][]byte) {
	w.signatures = sigs
}

func (w *builder) SetExtensionOptions(extOpts ...*codectypes.Any) {
	w.extensionOptions = extOpts
}

func (w *builder) SetNonCriticalExtensionOptions(extOpts ...*codectypes.Any) {
	w.nonCriticalExtOption = extOpts
}

func (w *builder) AddAuxSignerData(data tx.AuxSignerData) error {
	return fmt.Errorf("not supported")
}
