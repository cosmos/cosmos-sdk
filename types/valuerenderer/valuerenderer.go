package valuerenderer

import (
	"context"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ValueRenderer defines an interface to produce formated output for Int,Dec,Coin types as well as parse a string to Coin or Uint.
type ValueRenderer interface {
	Format(context.Context, interface{}) (string, error)
	Parse(context.Context, string) (interface{}, error)
}

// denomQuerierFunc takes a context and a denom as arguments and returns metadata, error
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
	p := message.NewPrinter(language.English)
	var sb strings.Builder

	switch v := x.(type) {
	case types.Dec:
		s := v.String()
		if len(s) == 0 {
			return "", errors.New("empty string")
		}

		strs := strings.Split(s, ".")

		if len(strs) == 2 {
			// there is a decimal place

			// format the first part
			i64, err := strconv.ParseInt(strs[0], 10, 64)
			if err != nil {
				return "", errors.New("unable to convert string to int64")
			}
			formated := p.Sprintf("%d", i64)

			// concatanate first part, "." and second part
			sb.WriteString(formated)
			sb.WriteString(".")
			sb.WriteString(strs[1])
		}

	case types.Int:
		s := v.String()
		if len(s) == 0 {
			return "", errors.New("empty string")
		}

		sb.WriteString(p.Sprintf("%d", v.Int64()))

	case types.Coin:
		metadata, err := dvr.denomQuerier(c, convertToBaseDenom(v.Denom))
		if err != nil {
			return "", err
		}

		newAmount, newDenom := p.Sprintf("%d", dvr.ComputeAmount(v, metadata)), metadata.Display
		sb.WriteString(newAmount)
		sb.WriteString(newDenom)

	default:
		panic("type is invalid")
	}

	return sb.String(), nil
}

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

// ComputeAmount calculates an amount to produce formated output
func (dvr DefaultValueRenderer) ComputeAmount(coin types.Coin, metadata banktypes.Metadata) int64 {
	var coinExp, displayExp int64
	for _, denomUnit := range metadata.DenomUnits {
		if denomUnit.Denom == coin.Denom {
			coinExp = int64(denomUnit.Exponent)
		}

		if denomUnit.Denom == metadata.Display {
			displayExp = int64(denomUnit.Exponent)
		}
	}

	expSub := float64(coinExp - displayExp)
	var amount int64

	switch {
	// negative , convert mregen to regen less zeroes
	case math.Signbit(expSub):
		amount = types.NewDecFromIntWithPrec(coin.Amount, int64(math.Abs(expSub))).TruncateInt64() // use Dec or just golang built in methods
	case !math.Signbit(expSub):
		amount = coin.Amount.Mul(types.NewInt(int64(math.Pow(10, expSub)))).Int64()
	// == 0, convert regen to regen, amount does not change
	default:
		amount = coin.Amount.Int64()
	}

	return amount
}

// Parse parses a string and takes a decision whether to convert it into Coin or Uint
func (dvr DefaultValueRenderer) Parse(ctx context.Context, s string) (interface{}, error) {
	if s == "" {
		return nil, errors.New("unable to parse empty string")
	}
	// remove all commas
	str := strings.ReplaceAll(s, ",", "")
	re := regexp.MustCompile(`\d+[mu]?regen`)
	// case 1: "1000000regen" => Coin
	if re.MatchString(str) {
		coin, err := coinFromString(str)
		if err != nil {
			return nil, err
		}

		return coin, nil
	}

	// case2: convert it to Uint
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return nil, err
	}

	return types.NewUint(i), nil
}

// should I add tests for coinFromString?
// coinFromString converts a string to coin
func coinFromString(s string) (types.Coin, error) {
	index := len(s) - 1
	for i := len(s) - 1; i >= 0; i-- {
		if unicode.IsLetter(rune(s[i])) {
			continue
		}

		index = i
		break
	}

	if index == len(s)-1 {
		return types.Coin{}, errors.New("unable to find a denonination")
	}

	amount, denom := s[:index+1], s[index+1:]
	// convert to int64 to make up Coin later
	amountInt, ok := types.NewIntFromString(amount)
	if !ok {
		return types.Coin{}, errors.New("unable to convert amountStr into int64")
	}

	return types.NewCoin(denom, amountInt), nil
}
