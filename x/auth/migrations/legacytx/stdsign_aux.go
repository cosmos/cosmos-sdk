package legacytx

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StdSignDocAux struct {
	AccountNumber uint64            `json:"account_number" yaml:"account_number"`
	Sequence      uint64            `json:"sequence" yaml:"sequence"`
	TimeoutHeight uint64            `json:"timeout_height,omitempty" yaml:"timeout_height"`
	ChainID       string            `json:"chain_id" yaml:"chain_id"`
	Memo          string            `json:"memo" yaml:"memo"`
	Msgs          []json.RawMessage `json:"msgs" yaml:"msgs"`
	Tip           sdk.Coins         `json:"tip" yaml:"tip"`
}

// StdSignBytes returns the bytes to sign for a transaction.
func StdSignAuxBytes(chainID string, accnum, sequence, timeout uint64, tip sdk.Coins, msgs []sdk.Msg, memo string) []byte {
	msgsBytes := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		legacyMsg, ok := msg.(LegacyMsg)
		if !ok {
			panic(fmt.Errorf("expected %T when using AMINO_AUX", (*LegacyMsg)(nil)))
		}

		msgsBytes = append(msgsBytes, json.RawMessage(legacyMsg.GetSignBytes()))
	}

	bz, err := legacy.Cdc.MarshalJSON(StdSignDocAux{
		AccountNumber: accnum,
		ChainID:       chainID,
		Memo:          memo,
		Msgs:          msgsBytes,
		Sequence:      sequence,
		TimeoutHeight: timeout,
		Tip:           tip,
	})
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bz)
}
