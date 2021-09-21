package valuerenderer

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// ValueRenderer defines an interface to produce formated output for Int,Dec,Coin types as well as parse a string to Coin or Uint.
type ValueRenderer interface {
	Format(context.Context, interface{}) (string, error)
	Parse(context.Context, string) (interface{}, error)
}

// denomQuerierFunc should take context and baseDenom to invoke DenomMetaData method on queryClient. This method will return banktypes.Metadata.
type denomQuerierFunc func(context.Context, string) (banktypes.Metadata, error)

// DefaultValueRenderer defines a struct that implements ValueRenderer interface
type DefaultValueRenderer struct {
	denomQuerier denomQuerierFunc
}

var _ ValueRenderer = &DefaultValueRenderer{}

// NewDefaultValueRenderer  initiates denomToMetadataMap field and returns DefaultValueRenderer struct
func NewDefaultValueRenderer(denomQuerier denomQuerierFunc) DefaultValueRenderer {
	return DefaultValueRenderer{
		denomQuerier: denomQuerier,
	}
}

// Format converts an empty interface into a string depending on interface type.
func (dvr DefaultValueRenderer) Format(c context.Context, x interface{}) (string, error) {
	var sb strings.Builder

	switch v := x.(type) {
	case types.Dec:
		s := v.String()
		if len(s) == 0 {
			return "", errors.New("empty string")
		}

		i := strings.Index(s, ".")
		first, second := s[:i], s[i+1:]
		first64, err := strconv.ParseInt(first, 10, 64)
		if err != nil {
			return "", err
		}

		sb.WriteString(humanize.Comma(first64))
		sb.WriteString(".")
		sb.WriteString(removeTrailingZeroes(second))

	case types.Int:
		s := v.String()
		if len(s) == 0 {
			return "", errors.New("empty string")
		}

		sb.WriteString(humanize.Comma(v.Int64()))

	case types.Coin:
		metadata, err := dvr.denomQuerier(c, convertToBaseDenom(v.Denom))
		if err != nil {
			return "", err
		}

		expSub := computeExponentSubtraction(v.Denom, metadata)

		formatedAmount := dvr.ComputeAmount(v.Amount.Int64(), expSub)

		sb.WriteString(formatedAmount)
		sb.WriteString(metadata.Display)

	default:
		panic("type is invalid")
	}

	return sb.String(), nil
}

// TODO address the cass where denom starts with "u"
func convertToBaseDenom(denom string) string {
	switch {
	// e.g. uregen => uregen
	case strings.HasPrefix(denom, "u"):
		return denom
	// e.g. mregen => uregen
	case strings.HasPrefix(denom, "m"):
		return "u" + denom[1:]
	// has no prefix regen => uregen
	default:
		return "u" + denom
	}
}

// computeExponentSubtraction iterates over metadata.DenomUnits and computes the subtraction of exponents
func computeExponentSubtraction(denom string, metadata banktypes.Metadata) float64 {
	var coinExp, displayExp int64
	for _, denomUnit := range metadata.DenomUnits {
		if denomUnit.Denom == denom {
			coinExp = int64(denomUnit.Exponent)
		}

		if denomUnit.Denom == metadata.Display {
			displayExp = int64(denomUnit.Exponent)
		}
	}

	return float64(coinExp - displayExp)
}

// removeTrailingZeroes removes trailing zeroes from a string
func removeTrailingZeroes(str string) string {
	index := len(str)-1
	for i := len(str) - 1; i > 0; i-- {
		if rune(str[i]) == rune('0') {
			continue
		} 
		
		index = i
		break
	}

	return str[:index+1]
}

// countTrailingZeroes counts the amount of trailing zeroes in a string
func countTrailingZeroes(str string) int {
	counter := 0
	for i := len(str) - 1; i > 0; i-- {
		if rune(str[i]) == rune('0') {
			counter++
		} else {
			break
		}
	}
	return counter
}

// ComputeAmount calculates an amount to produce formated output
func (dvr DefaultValueRenderer) ComputeAmount(amount int64, expSub float64) string {

	switch {
	// negative , convert mregen to regen less zeroes 23 => 0,023, expSub -3
	case math.Signbit(expSub):

		stringValue := strconv.FormatInt(amount, 10)
		count := countTrailingZeroes(stringValue)
		if count >= int(math.Abs(expSub)) {
			// case 1 if number of zeroes >= Abs(expSub)  23000, -3 => 23 (int64)
			x := amount / int64(math.Pow(10, math.Abs(expSub)))
			return humanize.Comma(x)
		} else {
			// case 2 number of trailing zeroes < abs(expSub)  23, -3,=> 0.023(float64)
			x := float64(float64(amount) / math.Pow(10, math.Abs(expSub)))
			return humanize.Ftoa(x)
		}
	// positive, e.g.convert mregen to uregen
	case !math.Signbit(expSub):
		x := amount * int64(math.Pow(10, expSub))
		return humanize.Comma(x)
	// == 0, convert regen to regen, amount does not change
	default:
		return humanize.Comma(amount)
	}
}

// Parse parses a string and takes a decision whether to convert it into Coin or Uint
func (dvr DefaultValueRenderer) Parse(ctx context.Context, s string) (interface{}, error) {
	if s == "" {
		return nil, errors.New("unable to parse empty string")
	}
	// remove all commas
	str := strings.ReplaceAll(s, ",", "")
	re := regexp.MustCompile(`(\d+)(\w+)`)
	// case 1: "1000000regen" => Coin
	if re.MatchString(str) {
		var amountStr, denomStr string
		s1 := re.FindAllStringSubmatch(str, -1) // [[1000000regen 1000000 regen]]
		amountStr, denomStr = s1[0][1], s1[0][2]

		amount, err := strconv.ParseInt(amountStr, 10, 64)
		if err != nil {
			return nil, err
		}

		return types.NewInt64Coin(denomStr, amount), nil
	}

	// case2: convert it to Uint
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return nil, err
	}

	return types.NewUint(i), nil
}
