package types

import (
	"fmt"
	"strings"
	"strconv"
)

type ValueRenderer interface {
	Format(interface{}) (string, error)
	Parse(s string) (interface{}, error)
}

type defaultValueRenderer struct {
	denomQuerier func(string) banktypes.Metadata //convert DenomUnits to Display units
}

var _ ValueRenderer = &defaultValueRenderer{}

// TODO more test example needed
func (d defaultValueRenderer) Format(i interface{}) (string, error) {
	  // Format(x interface{}) {
		switch i.(type) {
		case Int:
			// 1000000000000 => "1,000,000,000,000"
			// prefer number operations divide
			// https://stackoverflow.com/questions/13020308/how-to-fmt-printf-an-integer-with-thousands-comma
			// 1, there is no methods to convert Int to string
			// 2. use string builder and go backwards and take every 3 characters
			//  000,000,000,000,1  => reverse the string
			// 3.reverse the string at the end
		case Coin:  //convert Coin.Denom to Display.Denom
			// "1000000uregen" => "1regen"
		    // query denom.metadata from state
			//  we concatanate fields Denom(choose Display.Denom) and Amount
			// for Amount use the same algo then in case Int

		case Dec:
			// 

		default:
			return "",fmt.Errorf("incorrect type")
		}

	}
  

}

// TODO only 2 cases possible?
// "1,000,000regen" -> Coin
// "1,000,000" -> uint
func (d defaultValueRenderer) Parse(s string) (interface{}, error) {
	// "1,000,000regen" -> Coin we have to sepearate denom and amount
	                     "regen"

	if !strings.HasSuffix(s, denom) {
		// TODO handle this case  "1,000,000" -> Uint 10000000
		var sb strings.Builder
		for _, s := range strings.Split(s, ",") {
			// check if s does consist only from digits
			sb.WriteString(s)
		}// or use strings.Join

		// check if result does consist only from digits
		// make int from result and return it
		return NewUintFromString(sb.String()), nil // test if panics
	}

	// TODO handle this case "1,000,000regen" -> Coin
	index := strings.Index(s, denom) {
	// "1,000,000", "regen"
	amountStr, denomStr := strings.Join(s[:index], ""), strings.Join(s[index:], "")
	
	// remove all commas from "1,000,000" in amount Str => "1000000"
	// TODO consider to use standalone func for that cause this code block is repeated
	var sb strings.Builder
	for _, s := range strings.Split(amountStr, ",") {
		sb.WriteString(s)
	}

	i, ok := NewIntFromString(sb.String())
	if !ok {
		return nil, fmt.Errorf("unable to construct Int from str")
	}

	return NewInt64Coin(denomStr, i), nil
	
} 
