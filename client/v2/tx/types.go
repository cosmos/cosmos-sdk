package tx

import (
	"cosmossdk.io/client/v2/offchain"
	"fmt"
	typestx "github.com/cosmos/cosmos-sdk/types/tx"
	protov2 "google.golang.org/protobuf/proto"
)

// GasEstimateResponse defines a response definition for tx gas estimation.
type GasEstimateResponse struct {
	GasEstimate uint64 `json:"gas_estimate" yaml:"gas_estimate"`
}

func (gr GasEstimateResponse) String() string {
	return fmt.Sprintf("gas estimate: %d", gr.GasEstimate)
}

type TxWrapper struct {
	Tx *typestx.Tx
}

func (tx TxWrapper) GetMsgs() ([]protov2.Message, error) {
	//TODO implement me
	panic("implement me")
}

func (tx TxWrapper) GetSignatures() ([]offchain.OffchainSignature, error) {
	//TODO implement me
	panic("implement me")
}
