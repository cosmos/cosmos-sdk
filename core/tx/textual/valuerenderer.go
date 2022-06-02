package textual

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	cosmos_proto "github.com/cosmos/cosmos-proto"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const thousandSeparator string = "'"

// NewADR050ValueRenderer returns a new ValueRenderer based on the ADR-050
// spec.
func NewADR050ValueRenderer() ValueRenderer {
	return adr050ValueRenderer{}
}

// adr050ValueRenderer is a value renderer based on the ADR-050 spec.
type adr050ValueRenderer struct{}

var _ ValueRenderer = adr050ValueRenderer{}

func (r adr050ValueRenderer) Format(ctx context.Context, fd protoreflect.FieldDescriptor, v protoreflect.Value) ([]string, error) {
	result := []string{}
	switch {
	// Integers
	case fd.Kind() == protoreflect.Uint32Kind ||
		fd.Kind() == protoreflect.Uint64Kind ||
		fd.Kind() == protoreflect.Int32Kind ||
		fd.Kind() == protoreflect.Int64Kind ||
		(fd.Kind() == protoreflect.StringKind && isCosmosScalar(fd, "cosmos.Int")):
		{
			formatted, err := formatInteger(v.String())
			if err != nil {
				return nil, err
			}

			result = append(result, formatted)
		}
	// Decimals
	case fd.Kind() == protoreflect.StringKind && isCosmosScalar(fd, "cosmos.Dec"):
		{
			formatted, err := formatDecimal(v.String())
			if err != nil {
				return nil, err
			}

			result = append(result, formatted)
		}
	// Coins
	case fd.Kind() == protoreflect.MessageKind && (&basev1beta1.Coin{}).ProtoReflect().Descriptor() == fd.Message():
		{
			if fd.Cardinality() != protoreflect.Repeated {
				sdkCoin, err := convertApiCoinToSdkCoin(v.Interface().(protoreflect.Message).Interface().(*basev1beta1.Coin))
				if err != nil {
					return nil, err
				}

				// TODO Insert the correct metadata from state.
				formatted, err := formatCoin(sdkCoin, nil)
				if err != nil {
					return nil, err
				}

				result = append(result, formatted)
			} else {
				sdkCoins, err := convertApiCoinsToSdkCoins(v.Interface().([]*basev1beta1.Coin))
				if err != nil {
					return nil, err
				}

				// TODO Insert the correct metadata from state.
				formatted, err := formatCoins(sdkCoins, nil)
				if err != nil {
					return nil, err
				}

				result = append(result, formatted)
			}

		}

	default:
		return nil, fmt.Errorf("value renderers cannot format value %s of type %s", v, fd.Kind())
	}

	return result, nil
}

func (r adr050ValueRenderer) Parse(context.Context, protoreflect.FieldDescriptor, []string) (proto.Message, error) {
	panic("implement me")
}

// isCosmosScalar returns true if a field has the `cosmos_proto.scalar` field
// option.
func isCosmosScalar(fd protoreflect.FieldDescriptor, scalar string) bool {
	opts := fd.Options().(*descriptorpb.FieldOptions)
	if proto.GetExtension(opts, cosmos_proto.E_Scalar).(string) == scalar {
		return true
	}

	return false
}

// formatInteger formats an integer into a value-rendered string. This function
// operates with string manipulation (instead of manipulating the int or sdk.Int
// object).
func formatInteger(v string) (string, error) {
	if v[0] == '-' {
		v = v[1:]
	}
	if len(v) > 1 {
		v = strings.TrimLeft(v, "0")

	}

	startOffset := 3
	for outputIndex := len(v); outputIndex > startOffset; {
		outputIndex -= 3
		v = v[:outputIndex] + thousandSeparator + v[outputIndex:]
	}
	return v, nil
}

// formatDecimal formats a decimal into a value-rendered string. This function
// operates with string manipulation (instead of manipulating the sdk.Dec
// object).
func formatDecimal(v string) (string, error) {
	parts := strings.Split(v, ".")
	intPart, err := formatInteger(parts[0])
	if err != nil {
		return "", err
	}

	if len(parts) > 2 {
		return "", fmt.Errorf("invalid decimal %s", v)
	}

	if len(parts) == 1 {
		return intPart, nil
	}

	decPart := strings.TrimRight(parts[1], "0")
	if len(decPart) == 0 {
		return intPart, nil
	}

	return intPart + "." + decPart, nil
}

// formatDecimal formats a sdk.Coin into a value-rendered string, using the
// given metadata about the denom. It returns the formatted coin string, the
// display denom, and an optional error.
func formatCoin(coin sdk.Coin, metadata *bankv1beta1.Metadata) (string, error) {
	coinDenom := coin.Denom

	// Return early if no display denom or display denom is the current coin denom.
	if metadata == nil || metadata.Display == "" || coinDenom == metadata.Display {
		vr, err := formatDecimal(coin.Amount.String())
		return vr + " " + coin.Denom, err
	}

	dispDenom := metadata.Display

	// Find exponents of both denoms.
	var coinExp, dispExp uint32
	foundCoinExp, foundDispExp := false, false
	for _, unit := range metadata.DenomUnits {
		if coinDenom == unit.Denom {
			coinExp = unit.Exponent
			foundCoinExp = true
		}
		if dispDenom == unit.Denom {
			dispExp = unit.Exponent
			foundDispExp = true
		}
	}

	// If we didn't find either exponent, then we return early.
	if !foundCoinExp || !foundDispExp {
		vr, err := formatInteger(coin.Amount.String())
		return vr + " " + coin.Denom, err
	}

	exponentDiff := int64(coinExp) - int64(dispExp)
	var (
		dispAmount sdk.Dec
		err        error
	)
	if exponentDiff > 0 {
		dispAmount, err = sdk.NewDecFromStr(coin.Amount.String())
		if err != nil {
			return "", err
		}
		dispAmount = dispAmount.Mul(sdk.NewDec(10).Power(uint64(exponentDiff)))
	} else {
		dispAmount, err = sdk.NewDecFromStr(coin.Amount.String())
		if err != nil {
			return "", err
		}
		dispAmount = dispAmount.Quo(sdk.NewDec(10).Power(uint64(-exponentDiff)))
	}

	vr, err := formatDecimal(dispAmount.String())
	return vr + " " + dispDenom, err
}

// formatDecimal formats a sdk.Coins into a value-rendered string, which uses
// `formatCoin` separated by ", " (a comma and a space). It expects
func formatCoins(coins sdk.Coins, metadata []*bankv1beta1.Metadata) (string, error) {
	if len(coins) != len(metadata) {
		return "", fmt.Errorf("formatCoins expect one metadata for each coin; expected %d, got %d", len(coins), len(metadata))
	}

	formatted := make([]string, len(coins))
	for i, coin := range coins {
		var err error
		formatted[i], err = formatCoin(coin, metadata[i])
		if err != nil {
			return "", err
		}
	}

	// Sort the formatted coins by display denom.
	sort.SliceStable(formatted, func(i, j int) bool {
		denomI := strings.Split(formatted[i], " ")[1]
		denomJ := strings.Split(formatted[j], " ")[1]

		return denomI < denomJ
	})

	return strings.Join(formatted, ", "), nil
}

// convertApiCoinToSdkCoin converts *basev1beta1.Coin to sdk.Coin.
func convertApiCoinToSdkCoin(coin *basev1beta1.Coin) (sdk.Coin, error) {
	amt, ok := sdk.NewIntFromString(coin.Amount)
	if !ok {
		return sdk.Coin{}, fmt.Errorf("cannot convert %s to sdk.Int", coin.Amount)
	}

	return sdk.NewCoin(coin.Denom, amt), nil
}

// convertApiCoinsToSdkCoins converts []*basev1beta1.Coin to sdk.Coins.
func convertApiCoinsToSdkCoins(coins []*basev1beta1.Coin) (sdk.Coins, error) {
	sdkCoins := make([]sdk.Coin, len(coins))
	for i, c := range coins {
		sdkCoin, err := convertApiCoinToSdkCoin(c)
		if err != nil {
			return nil, err
		}

		sdkCoins[i] = sdkCoin
	}

	return sdk.NewCoins(sdkCoins...), nil
}
