package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/gogoproto/jsonpb"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}

// SignTx signs a transaction managed by the TxBuilder using a `name` key stored in Keybase.
// The new signature is appended to the TxBuilder when overwrite=false or overwritten otherwise.
// Don't perform online validation or lookups if offline is true.
func SignTx(txFactory tx.Factory, clientCtx client.Context, name string, txBuilder client.TxBuilder, offline, overwriteSig bool) error {
	k, err := txFactory.Keybase().Key(name)
	if err != nil {
		return err
	}

	// Ledger and Multisigs only support LEGACY_AMINO_JSON signing.
	if txFactory.SignMode() == signing.SignMode_SIGN_MODE_UNSPECIFIED &&
		(k.GetType() == keyring.TypeLedger || k.GetType() == keyring.TypeMulti) {
		txFactory = txFactory.WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	pubKey, err := k.GetPubKey()
	if err != nil {
		return err
	}
	addr := sdk.AccAddress(pubKey.Address())
	if !isTxSigner(addr, txBuilder.GetTx().GetSigners()) {
		return fmt.Errorf("%s: %s", sdkerrors.ErrorInvalidSigner, name)
	}
	if !offline {
		txFactory, err = populateAccountFromState(txFactory, clientCtx, addr)
		if err != nil {
			return err
		}
	}

	return tx.Sign(txFactory, name, txBuilder, overwriteSig)
}

// SignTxWithSignerAddress attaches a signature to a transaction.
// Don't perform online validation or lookups if offline is true, else
// populate account and sequence numbers from a foreign account.
// This function should only be used when signing with a multisig. For
// normal keys, please use SignTx directly.
func SignTxWithSignerAddress(txFactory tx.Factory, clientCtx client.Context, addr sdk.AccAddress,
	name string, txBuilder client.TxBuilder, offline, overwrite bool,
) (err error) {
	// Multisigs only support LEGACY_AMINO_JSON signing.
	if txFactory.SignMode() == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		txFactory = txFactory.WithSignMode(signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)
	}

	// check whether the address is a signer
	if !isTxSigner(addr, txBuilder.GetTx().GetSigners()) {
		return fmt.Errorf("%s: %s", sdkerrors.ErrorInvalidSigner, name)
	}

	if !offline {
		txFactory, err = populateAccountFromState(txFactory, clientCtx, addr)
		if err != nil {
			return err
		}
	}

	return tx.Sign(txFactory, name, txBuilder, overwrite)
}

// Read and decode a StdTx from the given filename. Can pass "-" to read from stdin.
func ReadTxFromFile(ctx client.Context, filename string) (tx sdk.Tx, err error) {
	var bytes []byte

	if filename == "-" {
		bytes, err = io.ReadAll(os.Stdin)
	} else {
		bytes, err = os.ReadFile(filename)
	}

	if err != nil {
		return
	}

	return ctx.TxConfig.TxJSONDecoder()(bytes)
}

// ReadTxsFromInput reads multiples txs from the given filename(s). Can pass "-" to read from stdin.
// Unlike ReadTxFromFile, this function does not decode the txs.
func ReadTxsFromInput(txCfg client.TxConfig, filenames ...string) (scanner *BatchScanner, err error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("no file name provided")
	}

	var infile io.Reader = os.Stdin
	if filenames[0] != "-" {
		buf := new(bytes.Buffer)
		for _, f := range filenames {
			bytes, err := os.ReadFile(filepath.Clean(f))
			if err != nil {
				return nil, fmt.Errorf("couldn't read %s: %w", f, err)
			}

			if _, err := buf.WriteString(string(bytes)); err != nil {
				return nil, fmt.Errorf("couldn't write to merged file: %w", err)
			}
		}

		infile = buf
	}

	return NewBatchScanner(txCfg, infile), nil
}

// NewBatchScanner returns a new BatchScanner to read newline-delimited StdTx transactions from r.
func NewBatchScanner(cfg client.TxConfig, r io.Reader) *BatchScanner {
	return &BatchScanner{Scanner: bufio.NewScanner(r), cfg: cfg}
}

// BatchScanner provides a convenient interface for reading batch data such as a file
// of newline-delimited JSON encoded StdTx.
type BatchScanner struct {
	*bufio.Scanner
	theTx        sdk.Tx
	cfg          client.TxConfig
	unmarshalErr error
}

// Tx returns the most recent Tx unmarshalled by a call to Scan.
func (bs BatchScanner) Tx() sdk.Tx { return bs.theTx }

// UnmarshalErr returns the first unmarshalling error that was encountered by the scanner.
func (bs BatchScanner) UnmarshalErr() error { return bs.unmarshalErr }

// Scan advances the Scanner to the next line.
func (bs *BatchScanner) Scan() bool {
	if !bs.Scanner.Scan() {
		return false
	}

	tx, err := bs.cfg.TxJSONDecoder()(bs.Bytes())
	bs.theTx = tx
	if err != nil && bs.unmarshalErr == nil {
		bs.unmarshalErr = err
		return false
	}

	return true
}

func populateAccountFromState(
	txBldr tx.Factory, clientCtx client.Context, addr sdk.AccAddress,
) (tx.Factory, error) {
	num, seq, err := clientCtx.AccountRetriever.GetAccountNumberSequence(clientCtx, addr)
	if err != nil {
		return txBldr, err
	}

	return txBldr.WithAccountNumber(num).WithSequence(seq), nil
}

// GetTxEncoder return tx encoder from global sdk configuration if ones is defined.
// Otherwise returns encoder with default logic.
func GetTxEncoder(cdc *codec.LegacyAmino) (encoder sdk.TxEncoder) {
	encoder = sdk.GetConfig().GetTxEncoder()
	if encoder == nil {
		encoder = legacytx.DefaultTxEncoder(cdc)
	}

	return encoder
}

func ParseQueryResponse(bz []byte) (sdk.SimulationResponse, error) {
	var simRes sdk.SimulationResponse
	if err := jsonpb.Unmarshal(strings.NewReader(string(bz)), &simRes); err != nil {
		return sdk.SimulationResponse{}, err
	}

	return simRes, nil
}

func isTxSigner(user sdk.AccAddress, signers []sdk.AccAddress) bool {
	for _, s := range signers {
		if bytes.Equal(user.Bytes(), s.Bytes()) {
			return true
		}
	}

	return false
}
