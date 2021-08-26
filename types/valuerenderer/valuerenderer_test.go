package valuerenderer_test

import (
	"regexp"
	"strings"
	"testing"
	"strconv"
	"errors"
	"unicode"


	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
)


func TestFormatInt(t *testing.T) {
	
	billionStr := "1000000000"
	x, ok := types.NewIntFromString(billionStr)
	require.True(t, ok)

	d := valuerenderer.DefaultValueRenderer{}
	s, err := d.Format(x)
	require.NoError(t, err)
	require.Equal(t, s, "1,000,000,000")
}

/*
func TestFormatInt(t *testing.T) {
	v := uint64(1000000000)
	x := types.NewIntFromUint64(v)
	x64 := x.Int64()
	p := message.NewPrinter(language.English)
	s := p.Sprintf("%d", x64)
	require.Equal(t, s, "1,000,000,000")
}
*/


func TestParseString(t *testing.T) {
   re := regexp.MustCompile(`\d+[mu]?regen`)

   tt := []struct {
	   str string
	   denomExp bool
	   errExp bool
   }{
	   {"10regen", true, false},
	   {"1,000,000", false, false},
	   {"323,000,000", false, false},
	   {"1mregen", true, false},
	   {"500uregen", true, false},
	   {"500,000,000regen", true, false},
	   {"500,000,000regen", false, true},
	   {"1,500,000,000regen", true, false},
	   {"394,382,328uregen", true, false},
   }

   for _, tc := range tt {
	   t.Run(tc.str, func(t *testing.T) {
			s := strings.ReplaceAll(tc.str, ",", "")
			if tc.denomExp {
				// make sure it matches regexp to make up a Coin
				require.True(t, re.MatchString(s))
				c, err := coinFromString(s)
				require.NoError(t, err)
				require.NotNil(t, c)
			} else {
				// convert to Uint
				i64, err := strconv.ParseUint(s, 10, 64)
				if tc.errExp {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.NotNil(t, types.NewUint(i64))
				}
			} 
	   })
   }
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
	amountStr := s[:index+1]
	// convert to int64 to make up Coin later
	amountInt, ok := types.NewIntFromString(amountStr)
	if !ok {
		return types.Coin{}, errors.New("unable convert amoountStr to int64")
	}

	return types.NewCoin(denom, amountInt), nil
}







