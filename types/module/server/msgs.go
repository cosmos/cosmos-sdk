package server

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMsgs takes a slice of sdk.Msg's and turn them into Any's.
// This is similar to what is in the cosmos-sdk tx builder
// and could eventually be merged in.
func SetMsgs(msgs []sdk.Msg) ([]*types.Any, error) {
	anys := make([]*types.Any, len(msgs))
	for i, msg := range msgs {
		var err error
		anys[i], err = types.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
	}
	return anys, nil
}

// GetMsgs takes a slice of Any's and turn them into sdk.Msg's.
// This is similar to what is in the cosmos-sdk sdk.Tx
// and could eventually be merged in.
func GetMsgs(anys []*types.Any) ([]sdk.Msg, error) {
	msgs := make([]sdk.Msg, len(anys))
	for i, any := range anys {
		cached := any.GetCachedValue()
		if cached == nil {
			return nil, fmt.Errorf("any cached value is nil, proposal messages must be correctly packed Any values.")
		}
		msgs[i] = cached.(sdk.Msg)
	}
	return msgs, nil
}

func UnpackInterfaces(unpacker types.AnyUnpacker, anys []*types.Any) error {
	for _, any := range anys {
		var msg sdk.Msg
		err := unpacker.UnpackAny(any, &msg)
		if err != nil {
			return err
		}
	}

	return nil
}
