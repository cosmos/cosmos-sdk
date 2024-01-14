package types

import (
	"encoding/json"
	"fmt"
	protov2 "google.golang.org/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
)

type (
	// MsgV2 defines the interface a transaction message needed to fulfill.
	MsgV2 = protov2.Message

	ProtoMessage interface {
		Msg | MsgV2
	}

	// TxV2 defines an interface a transaction must fulfill.
	TxV2 interface {
		// GetMsgs gets the transaction's messages as google.golang.org/protobuf/proto.Message's.
		GetMsgs() ([]protov2.Message, error)
	}

	// TxV2WithFee defines the interface to be implemented by Tx to use the FeeDecorators
	TxV2WithFee interface {
		TxV2
		Fee
	}

	// TxV2WithMemo must have GetMemo() method to use ValidateMemoDecorator
	TxV2WithMemo interface {
		TxV2
		GetMemo() string
	}

	// TxV2WithTimeoutHeight extends the Tx interface by allowing a transaction to
	// set a height timeout.
	TxV2WithTimeoutHeight interface {
		TxV2
		GetTimeoutHeight() uint64
	}

	// TxV2Decoder unmarshals transaction bytes
	TxV2Decoder func(txBytes []byte) (TxV2, error)

	// TxV2Encoder marshals transaction to bytes
	TxV2Encoder func(tx TxV2) ([]byte, error)
)

// GetMsgV2FromTypeURL returns a `sdk.MsgV2` message type from a type URL
func GetMsgV2FromTypeURL(cdc codec.Codec, input string) (MsgV2, error) {
	var msg MsgV2
	bz, err := json.Marshal(struct {
		Type string `json:"@type"`
	}{
		Type: input,
	})
	if err != nil {
		return nil, err
	}

	if err := cdc.UnmarshalInterfaceJSON(bz, &msg); err != nil {
		return nil, fmt.Errorf("failed to determine sdk.Msg for %s URL : %w", input, err)
	}

	return msg, nil
}
