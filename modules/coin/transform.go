package coin

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
)

type coinTx struct {
	Inputs  []coinInput  `json:"inputs"`
	Outputs []coinOutput `json:"outputs"`
}

type coinInput struct {
	Sender string `json:"sender"`
	Coins  Coins  `json:"coins"`
}

type coinOutput struct {
	Receiver string `json:"receiver"`
	Coins    Coins  `json:"coins"`
}

// ExtractCoinTx makes nice json from raw tx bytes
func ExtractCoinTx(data []byte) (interface{}, error) {
	tx, err := sdk.LoadTx(data)
	if err != nil {
		return nil, err
	}
	txl, ok := tx.Unwrap().(sdk.TxLayer)
	for ok {
		tx = txl.Next()
		txl, ok = tx.Unwrap().(sdk.TxLayer)
	}
	ctx, ok := tx.Unwrap().(SendTx)
	if !ok {
		return nil, errors.ErrUnknownTxType(tx)
	}
	// now reformat this....
	return &coinTx{
		Inputs:  transformInputs(ctx.Inputs),
		Outputs: transformOutputs(ctx.Outputs),
	}, nil
}

func transformInputs(input []TxInput) []coinInput {
	out := make([]coinInput, 0, len(input))
	for _, in := range input {
		out = append(out, coinInput{
			Sender: in.Address.String(),
			Coins:  in.Coins,
		})
	}
	return out
}

func transformOutputs(output []TxOutput) []coinOutput {
	out := make([]coinOutput, 0, len(output))
	for _, val := range output {
		out = append(out, coinOutput{
			Receiver: val.Address.String(),
			Coins:    val.Coins,
		})
	}
	return out
}
