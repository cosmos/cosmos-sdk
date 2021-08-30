package valuerenderer_test

import (

	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
)


func TestFormat(t *testing.T) {

	d := valuerenderer.NewDefaultValueRenderer()

	decimal, _ := types.NewDecFromStr("349383323.894")
	i, _ := types.NewIntFromString("2323293999402003")
	
	
	// TODO consider add panic case and lens(strs) > 2 
	tt := []struct{
		name string
		input interface{}
		expRes string
		isIntType bool
		expErr bool
	}{
		{"nil", nil, "", false, true},
		{"convert a million, no error", types.NewInt(int64(1000000)),"1,000,000", true, false},
		{"empty string error", types.Int{}, "", true, true},
		{"Decimal, no error", decimal, "349,383,323.894", true, false},
		{"Int, no error", i, "232,329,399,940,200,3", true, false},
		{"Coin, no error", i, "232,329,399,940,200,3", true, false},

		//{"invalid string input panic", "qwerty", "", true, true},


	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := d.Format(tc.input)
			if tc.expErr {
				require.Error(t, err)
				require.Nil(t, res)
				return 
			} 
			
			if tc.isIntType {
				require.Equal(t, tc.expRes, res)
			} else {
				

			}

		
            


		})
	}
}

func TestParseString(t *testing.T) {
   re := regexp.MustCompile(`\d+[mu]?regen`)
   d := valuerenderer.NewDefaultValueRenderer()

   tt := []struct {
	   str string
	   denomExp bool
	   expErr bool
   }{
	   {"", false, true},
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
		    x, err := d.Parse(tc.str)
			// TODO reconsider logic - put expErr at first
			if tc.denomExp {
				require.NoError(t, err)
				coin, ok := x.(types.Coin)
				require.True(t, ok)
				require.NotNil(t, coin)
				require.True(t, re.MatchString(tc.str))
			} else {
				if tc.expErr {
					require.Error(t, err)
					require.Nil(t, x)
				} else {
				require.NoError(t, err)
				u, ok := x.(types.Uint)
				require.True(t, ok)
				require.NotNil(t, u)
				}
			} 
	   })
   }
}








