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

type defaultValueRenderer struct {}

var _ ValueRenderer = &defaultValueRenderer{}

// TODO more test example needed
func (d defaultValueRenderer) Format(i interface{}) (string, error) {
	  // Format(x interface{}) {
		switch i.(type) {
		case Int:
			// 1000000000000 => 1,000,000,000,000

		case Coins:
			// "1000000000uregen" => "1regen"
			c, ok := i.(Coins)


		case Dec:

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
		result = ""
		for _, s := range strings.Split(s, ",") {
			// check if s does consist only from digits
			result += s
		}// or use strings.Join

		// check if result does consist only from digits
		// make int from result and return it
		return NewUintFromString(result), nil // test if panics
	}

	// TODO handle this case "1,000,000regen" -> Coin
	index := strings.Index(s, denom) {
	// "1,000,000", "regen"
	amountStr, denomStr := strings.Join(s[:index], ""), strings.Join(s[index:], "")
	
	// remove all commas from "1,000,000" in amount Str => "1000000"
	// TODO consider to use standalone func for that cause this code block is repeated
	validAmountStr = ""
	for _, str := range strings.Split(amountStr, ",") {
		validAmountStr += str
	}

	// convert amount "1000000" =< 1000000 int64 
	i, err := strconv.ParseInt(validAmountStr, 10, 64)
	if err != nil {
		return nil, err
	}

	return NewInt64Coin(denomStr, i), nil
	

} 