package simapp

import (
	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AminoJSONTxDecoder returns an Amino JSON TxDecoder using the provided cfg Marshaler
func AminoJSONTxDecoder(cfg params.EncodingConfig) types.TxDecoder {
	return func(txBytes []byte) (types.Tx, error) {
		var tx authtypes.StdTx
		err := cfg.Marshaler.UnmarshalJSON(txBytes, &tx)
		if err != nil {
			return nil, err
		}
		return tx, nil
	}
}
