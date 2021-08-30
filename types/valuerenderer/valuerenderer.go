package valuerenderer

import (
	"errors"
    "strings"
	"math"
	"regexp"
	"unicode"
	"strconv"


	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type ValueRenderer interface {
	Format(interface{}) (string, error)
	Parse(string) (interface{}, error)
}

// create default value rreenderer in CLI and then get context from CLI 
type DefaultValueRenderer struct {
	// /string is denom that user sents
	denomQuerier func(string) banktypes.Metadata// define in test only //convert DenomUnits to Display units 
	//queryClient banktypes.QueryClient
}

func NewDefaultValueRenderer() DefaultValueRenderer {
	return DefaultValueRenderer{}
}

var _ ValueRenderer = &DefaultValueRenderer{}

// Format converts an empty interface into a string depending on interface type.
func (d DefaultValueRenderer) Format(x interface{}) (string, error) { 
	if x == nil {
		return "", errors.New("x is nil")
	}

	p := message.NewPrinter(language.English)
	var sb strings.Builder

	switch x.(type) {
		case types.Int: 
			i, ok := x.(types.Int)
			if !ok {
				return "", errors.New("unable to cast interface{} to Int")
			}

			s := i.String()
			if len(s) == 0 { 
				return "", errors.New("empty string")
			}

			// TODO  check for negative values in tests
			// find out whether it is Dec or Int type
			strs := strings.Split(s, ".")
			
			if len(strs) == 2 {
				// there is a decimal place
				// format the first part 
				sb.WriteString(p.Sprint(strs[0]))
				sb.WriteString(strs[1])
				return sb.String(), nil

			} else if len(strs) > 2 {
				// invalid input
				return "", types.ErrInvalidDecimalStr	
			}
		
			// there is no decimal place
			sb.WriteString(p.Sprintf("%d",i.Int64()))
	
		case types.Coin:
			/*
			   - name = regen, exponent = 0
    		   - name = uregen, exponent = 6
    		   - name = mregen, exponent = 3
			ex1: Coin(denom, amount)
			    Coin               Display
			    "1000000uregen"(exp 6) => regen (exp 0)
				0 - 6 = -6
			ex2 23000 uregen  ->  mregen ()

				case Coin:  //convert Coin.Denom to Display.Denom
			"1000000uregen" => "1regen"
			"1 * 10^-6 regen
			query denom.metadata from state
		 	we concatanate fields Denom(choose Display.Denom) and Amount
			for Amount use the same algo then in case Int

			*/
			coin, ok := x.(types.Coin)
			if !ok {
				return "", errors.New("unable to cast empty interface to Coin")
			}
		
			metadata := d.denomQuerier(coin.Denom)
		
			var srcExp, dstExp uint32
			// find exponent that matches coin.Denom  {
			for _, denomUnit := range metadata.DenomUnits { 
				//  test  23000000 mregen 3  =>  "regen" exp 0
				if denomUnit.Denom == coin.Denom {
					srcExp = denomUnit.Exponent
				}

				if denomUnit.Denom == metadata.Display {
					dstExp = denomUnit.Exponent
				}
			}
            // wrap this code block into function
			exp := dstExp - srcExp
			multiplier := types.NewInt(int64(math.Pow(float64(10), float64(exp))))
			amount := coin.Amount.Mul(multiplier)

			sb.WriteString(p.Sprintf("%d",amount.Int64()))
			sb.WriteString(coin.Denom)
		/* ?
		default:
			panic("type is invalid")
		}
		*/

		return sb.String(), nil	
}

func (d DefaultValueRenderer) DenomQuerier(s string) banktypes.Metadata {
	

}
			
// Parse parses string and takes a decision whether to convert it into Coin or Uint
func (d DefaultValueRenderer) Parse(s string) (interface{}, error) { 
	if s == ""{
		return nil, errors.New("unable to parse empty string")
	}

	str := strings.ReplaceAll(s, ",", "")
	re := regexp.MustCompile(`\d+[mu]?regen`)
	// case 1: "1000000regen" => Coin
	if re.MatchString(str) {
		coin, err := coinFromString(str)
		if err != nil {
			return nil, err
		}
		return coin,nil
	}

	// case2: convert it to Uint
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return nil, err
	}

	return types.NewUint(i), nil	
}

func coinFromString(s string) (types.Coin, error) {
	index := len(s) -1
	for i := len(s)-1; i >= 0; i--{
		if unicode.IsLetter(rune(s[i])) {
			continue
		}

		index = i
		break
	}

	if index == len(s)-1 {
		return types.Coin{}, errors.New("no denom has been found")
	}
    
	denom := s[index+1:]
	amount := s[:index+1]
	// convert to int64 to make up Coin later
	amountInt, ok := types.NewIntFromString(amount)
	if !ok {
		return types.Coin{}, errors.New("unable convert amountStr to int64")
	}

	return types.NewCoin(denom, amountInt), nil
}

 