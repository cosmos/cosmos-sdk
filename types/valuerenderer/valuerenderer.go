package valuerenderer

import (
	"errors"
//	"strings"
//	"strconv"
//	"unicode"

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
func NewDefaultValueRenderer() {}

var _ ValueRenderer = &DefaultValueRenderer{}


func (d DefaultValueRenderer) Format(x interface{}) (string, error) { 
	switch x.(type) {
		case types.Int:
			i, ok := x.(types.Int)
			if !ok {
				return "", errors.New("unable to cast interface{} to Int")
			}
			
			p := message.NewPrinter(language.English)
			return p.Sprintf("%d",i.Int64()),nil
		case types.Coin:
			/*

				case Coin:  //convert Coin.Denom to Display.Denom
			"1000000uregen" => "1regen"
			    
			query denom.metadata from state
		 	we concatanate fields Denom(choose Display.Denom) and Amount
			for Amount use the same algo then in case Int

			*/
			coin, ok := x.(types.Coin)
			if !ok {
				return "", errors.New("unable to cast interface{} to Coin")
			}
		
			metadata := d.denomQuerier(coin.Denom)
			for _, denom := range metadata.DenomUnits { // find exponent that matches coin.Denom  {
				// 23000000 mregen   =>  "regen"
				if denom.Denom == metadata.Display {

					//TODO
					// find exponent that matches Display 
					// substract 2 exponents
				}
			}
			

			// som real lifec examplke 23000 uregen   -> regen in tests


			return "", nil
		// TODO case Dec	
		default: 
			return "good", nil
	}
}

// TODO only 2 cases possible?
// "1,000,000regen" -> Coin
// "1,000,000" -> Uint

func (d DefaultValueRenderer) Parse(s string) (interface{}, error) { return nil,nil }
// See TestParseString