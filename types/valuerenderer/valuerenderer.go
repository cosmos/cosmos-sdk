package valuerenderer

import (
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
	Format(interface{}) (string, error)
	Parse(string) (interface{}, error)
}

// DefaultValueRenderer defines a struct that implements ValueRenderer interface
type DefaultValueRenderer struct {
	// denomToMetadataMap represents a hashmap to lookup banktypes.Metadata by coin denom
	denomToMetadataMap map[string]banktypes.Metadata
}

var _ ValueRenderer = &DefaultValueRenderer{}

// TODO consider to move into valuerenderer_test.go
// NewDefaultValueRenderer  initiates denomToMetadataMap field and returns DefaultValueRenderer struct
func NewDefaultValueRenderer() DefaultValueRenderer {
	return DefaultValueRenderer{denomToMetadataMap: make(map[string]banktypes.Metadata)}
}

// TODO decide what is the most effiecnt way to map key to metadata
// 1. current implementation O(m*d)
// 2. use metadata.Symbol is a key ,but this is an optional field
// 3. use metadata.Base or ,metadata.Display as a key in this case the helper function
// SetDenomToMetadataMap populates a hashmap that maps coin denomination to a metadata.
func (dvr DefaultValueRenderer) SetDenomToMetadataMap(metadatas []banktypes.Metadata) error {
	// TODO should I validate denom?
	if metadatas == nil {
		return errors.New("empty metadatas")
	}

	// O(m*d), wheren m is number of metadatas and d is a number of denomUnits in each metadata
	for _, m := range metadatas {
		for _, denomUnit := range m.DenomUnits {
			dvr.denomToMetadataMap[denomUnit.Denom] = m
		}
	}

	return nil
}

// Format converts an empty interface into a string depending on interface type.
func (dvr DefaultValueRenderer) Format(x interface{}) (string, error) {
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
		metadata, err := dvr.LookupMetadataByDenom(v.Denom)
		if err != nil {
			return "", err
		}

		newAmount, newDenom := p.Sprintf("%d", dvr.ComputeAmount(v, metadata)), metadata.Display
		sb.WriteString(newAmount)
		sb.WriteString(newDenom)

		//	default:
		//	panic("type is invalid")
	}

	return sb.String(), nil
}

// LookupMetadataByDenom lookups metadata by coin denom
func (dvr DefaultValueRenderer) LookupMetadataByDenom(denom string) (banktypes.Metadata, error) {
	// lookup metadata by displayDenom
	metadata, ok := dvr.denomToMetadataMap[denom]
	if !ok {
		return banktypes.Metadata{}, errors.New("unable to lookup displayDenom in denomToMetadataMap")
	}

	return metadata, nil
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
		// TODO or should i use math package?
		amount = types.NewDecFromIntWithPrec(coin.Amount, int64(math.Abs(expSub))).TruncateInt64() // use Dec or just golang built in methods
	// positive, convert mregen to uregen
	case !math.Signbit(expSub):
		amount = coin.Amount.Mul(types.NewInt(int64(math.Pow(10, expSub)))).Int64()
	// == 0, convert regen to regen, amount does not change
	default:
		amount = coin.Amount.Int64()
	}

	return amount
}

// Parse parses a string and takes a decision whether to convert it into Coin or Uint
func (dvr DefaultValueRenderer) Parse(s string) (interface{}, error) {
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
