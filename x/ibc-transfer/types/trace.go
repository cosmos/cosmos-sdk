package types

import (
	"bytes"
	fmt "fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// Hash returns the hex bytes of the SHA256 hash of the DenomTrace fields using the following formula:
//
// 	hash = sha256(trace + "/" + baseDenom)
func (dt DenomTrace) Hash() tmbytes.HexBytes {
	return tmhash.Sum([]byte(dt.GetPrefix() + dt.BaseDenom))
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

func (dt DenomTrace) validateTrace() error {
	// empty trace is accepted when token lives on the original chain

	switch {
	case dt.Trace == "" && dt.BaseDenom != "":
		return nil
	case strings.TrimSpace(dt.Trace == ""):
		return fmt.Errorf("cannot have an empty trace and empty base denomination")
	}

	traceSplit := strings.Split(dt.Trace, "/")

	switch {
	case traceSplit[0] == dt.Trace:
		return fmt.Errorf("trace %s must contain '/' separators", dt.Trace)
	case len(traceSplit)%2 != 0:
		return fmt.Errorf("trace info %s must come in pairs of port and channel identifiers '{portID}/{channelID}'", dt.Trace)
	}

	// validate correctness of port and channel identifiers
	for i := 0; i < len(traceSplit); i += 2 {
		if err := host.PortIdentifierValidator(traceSplit[i]); err != nil {
			return sdkerrors.Wrapf(err, "invalid port ID at position %d", i)
		}
		if err := host.ChannelIdentifierValidator(traceSplit[i+1]); err != nil {
			return sdkerrors.Wrapf(err, "invalid channel ID at position %d", i)
		}
	}
	return nil
}

// Validate performs a basic validation of the DenomTrace fields.
func (dt DenomTrace) Validate() error {
	if err := sdk.ValidateDenom(dt.BaseDenom); err != nil {
		return err
	}

	return dt.validateTrace()
}

// Traces defines a wrapper type for a slice of IdentifiedDenomTraces.
type Traces []IdentifiedDenomTrace

// Validate performs a basic validation of each denomination trace info.
func (t Traces) Validate() error {
	seenTraces := make(map[string]bool)
	for i, trace := range t {
		if seenTraces[trace.Hash] {
			return fmt.Errorf("duplicated denomination trace with hash  %s", trace.Hash)
		}
		hash := tmhash.Sum([]byte(trace.Trace + "/" + trace.BaseDenom))
		if !bytes.Equal(tmbytes.HexBytes(trace.Hash), hash) {
			return fmt.Errorf("trace hash mismatch, expected %s got %s", trace.Hash, hash)
		}

		denomTrace := DenomTrace{
			Trace:     trace.Trace,
			BaseDenom: trace.BaseDenom,
		}

		if err := denomTrace.Validate(); err != nil {
			sdkerrors.Wrapf(err, "failed denom trace %d validation", i)
		}
		seenTraces[trace.Hash] = true
	}
	return nil
}

// ValidateIBCDenom checks that the denomination for an IBC fungible token is valid.
func ValidateIBCDenom(denom string) (tmbytes.HexBytes, error) {
	denomSplit := strings.SplitN(denom, "/", 3)

	var err error
	switch {
	case len(denomSplit) == 0:
		err = sdkerrors.Wrap(ErrInvalidDenomForTransfer, denom)
	case denomSplit[0] == denom:
		err = sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "denomination should be prefixed with the format 'ibc/{hash(trace + \"/\" + %s)}'", denom)
	case denomSplit[0] != "ibc":
		err = sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "denomination %s must start with 'ibc'", denom)
	case len(denomSplit) == 2 && len(denomSplit[1]) != tmhash.Size:
		err = sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "invalid SHA256 hash %s length, expected %d, got %d", denomSplit[1], tmhash.Size, len(denomSplit[1]))
	default:
		err = sdkerrors.Wrap(ErrInvalidDenomForTransfer, denom)
	}

	if err != nil {
		return nil, err
	}

	return tmbytes.HexBytes(denomSplit[1]), nil
}
