package types

import (
	fmt "fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Hash returns the hex bytes of the SHA256 hash of the DenomTrace fields using the following formula:
//
// 	hash = sha256(trace + "/" + baseDenom)
func (dt DenomTrace) Hash() tmbytes.HexBytes {
	return tmhash.Sum(dt.GetDenomPrefix() + dt.BaseDenom)
}

// GetPrefix returns the receiving denomination prefix composed by the trace info and a separator.
func (dt DenomTrace) GetPrefix() string {
	return dt.Trace + "/"
}

// IBCDenom a coin denomination for an ICS20 fungible token in the format 'ibc/{hash(trace + "/" + baseDenom)}'. If the trace is empty, it will return the base denomination.
func (dt DenomTrace) IBCDenom() string {
	if dt.Trace != "" {
		return fmt.Sprintf("ibc/%s", dt.Hash())
	}
	return dt.BaseDenom
}

// RemovePrefix trims the first portID/channelID pair from the trace info. If the trace is already empty it will perform a no-op. If the trace is incorrectly constructed or doesn't have separators it will return an error.
func (dt *DenomTrace) RemovePrefix() error {
	if dt.Trace == "" {
		return nil
	}

	traceSplit := strings.SplitN(dt.Trace, "/", 3)

	var err error
	switch {
	case len(traceSplit) == 0, traceSplit[0] == dt.Trace:
		err = sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "trace info %s must contain '/' separators", dt.Trace)
	case len(traceSplit) == 1:
		err = sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "trace info %s must come in pairs of '{portID}/channelID}'", dt.Trace)
	case len(traceSplit) == 2:
		dt.Trace = ""
	case len(traceSplit) == 3:
		dt.Trace = traceSplit[2]
	}

	if err != nil {
		return err
	}

	return nil
}
