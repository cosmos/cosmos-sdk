package types

import (
	"encoding/hex"
	"errors"
	fmt "fmt"
	"strings"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// ParseDenomTrace parses a string with the ibc prefix (denom trace) and the base denomination
// into a DenomTrace type.
//
// Examples:
//
// 	- "portidone/channelidone/uatom" => DenomTrace{Trace: "portidone/channelidone", BaseDenom: "uatom"}
// 	- "uatom" => DenomTrace{Trace: "", BaseDenom: "uatom"}
func ParseDenomTrace(rawDenom string) DenomTrace {
	denomSplit := strings.Split(rawDenom, "/")

	if denomSplit[0] == rawDenom {
		return DenomTrace{
			Trace:     "",
			BaseDenom: rawDenom,
		}
	}

	return DenomTrace{
		Trace:     strings.Join(denomSplit[:len(denomSplit)-1], "/"),
		BaseDenom: denomSplit[len(denomSplit)-1],
	}
}

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

// IBCDenom a coin denomination for an ICS20 fungible token in the format 'ibc/{hash(trace +
// baseDenom)}'. If the trace is empty, it will return the base denomination.
func (dt DenomTrace) IBCDenom() string {
	if dt.Trace != "" {
		return fmt.Sprintf("ibc/%s", dt.Hash())
	}
	return dt.BaseDenom
}

// RemovePrefix trims the first portID/channelID pair from the trace info. If the trace is already
// empty it will perform a no-op. If the trace is incorrectly constructed or doesn't have separators
// it will return an error.
func (dt *DenomTrace) RemovePrefix() {
	if dt.Trace == "" {
		return
	}

	traceSplit := strings.SplitN(dt.Trace, "/", 3)

	switch {
	// NOTE: other cases are checked during msg validation
	case len(traceSplit) == 2:
		dt.Trace = ""
	case len(traceSplit) == 3:
		dt.Trace = traceSplit[2]
	}
}

func validateTraceIdentifiers(identifiers []string) error {
	if len(identifiers)%2 != 0 {
		return errors.New("trace info must come in pairs of port and channel identifiers '{portID}/{channelID}'")
	}

	// validate correctness of port and channel identifiers
	for i := 0; i < len(identifiers); i += 2 {
		if err := host.PortIdentifierValidator(identifiers[i]); err != nil {
			return sdkerrors.Wrapf(err, "invalid port ID at position %d", i)
		}
		if err := host.ChannelIdentifierValidator(identifiers[i+1]); err != nil {
			return sdkerrors.Wrapf(err, "invalid channel ID at position %d", i)
		}
	}
	return nil
}

// Validate performs a basic validation of the DenomTrace fields.
func (dt DenomTrace) Validate() error {
	// empty trace is accepted when token lives on the original chain
	switch {
	case dt.Trace == "" && dt.BaseDenom != "":
		return nil
	case strings.TrimSpace(dt.Trace) == "" && strings.TrimSpace(dt.BaseDenom) == "":
		return fmt.Errorf("cannot have an empty trace and empty base denomination")
	case dt.Trace != "" && strings.TrimSpace(dt.BaseDenom) == "":
		return sdkerrors.Wrap(ErrInvalidDenomForTransfer, "denomination cannot be blank")
	}

	// NOTE: no base denomination validation

	identifiers := strings.Split(dt.Trace, "/")
	return validateTraceIdentifiers(identifiers)
}

// Traces defines a wrapper type for a slice of IdentifiedDenomTraces.
type Traces []DenomTrace

// Validate performs a basic validation of each denomination trace info.
func (t Traces) Validate() error {
	seenTraces := make(map[string]bool)
	for i, trace := range t {
		hash := trace.Hash().String()
		if seenTraces[hash] {
			return fmt.Errorf("duplicated denomination trace with hash  %s", trace.Hash())
		}

		if err := trace.Validate(); err != nil {
			return sdkerrors.Wrapf(err, "failed denom trace %d validation", i)
		}
		seenTraces[hash] = true
	}
	return nil
}

// ValidatePrefixedDenom checks that the denomination for an IBC fungible token packet denom is correctly prefixed.
// The function will return no error if the given string follows one of the two formats:
//
// - Prefixed denomination: '{portIDN}/{channelIDN}/.../{portID0}/{channelID0}/baseDenom'
//
// - Unprefixed denomination: 'baseDenom'
func ValidatePrefixedDenom(denom string) error {
	denomSplit := strings.Split(denom, "/")
	if denomSplit[0] == denom && strings.TrimSpace(denom) != "" {
		// NOTE: no base denomination validation
		return nil
	}

	if strings.TrimSpace(denomSplit[len(denomSplit)-1]) == "" {
		return sdkerrors.Wrap(ErrInvalidDenomForTransfer, "base denomination cannot be blank")
	}

	identifiers := denomSplit[:len(denomSplit)-1]
	return validateTraceIdentifiers(identifiers)
}

// ValidateIBCDenom validates that the given denomination is either:
//
// - A valid base denomination (eg: 'uatom')
//
// - A valid fungible token representation (i.e 'ibc/{hash}') per ADR 001 https://github.com/cosmos/cosmos-sdk/blob/master/docs/architecture/adr-001-coin-source-tracing.md
func ValidateIBCDenom(denom string) error {
	denomSplit := strings.SplitN(denom, "/", 2)

	switch {
	case denomSplit[0] != "ibc" && denomSplit[0] == denom && strings.TrimSpace(denom) != "":
		// NOTE: coin base denomination already verified
		return nil
	case len(denomSplit) != 2, denomSplit[0] != "ibc", strings.TrimSpace(denomSplit[1]) == "":
		return sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "denomination should be prefixed with the format 'ibc/{hash(trace + \"/\" + %s)}'", denom)
	}

	hash, err := hex.DecodeString(denomSplit[1])
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "invalid denom trace hash %s: %s", denomSplit[1], err)
	}

	hash = tmbytes.HexBytes(hash)
	if err := tmtypes.ValidateHash(hash); err != nil {
		return sdkerrors.Wrapf(ErrInvalidDenomForTransfer, "invalid denom trace hash %s: %s", denomSplit[1], err)
	}

	return nil
}
