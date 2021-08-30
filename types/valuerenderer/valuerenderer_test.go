package valuerenderer_test

import (

	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
)

func TestFormatCoin(t *testing.T) {
    dvr := valuerenderer.NewDefaultValueRenderer()

	tt := []struct{
		name string
		coin types.Coin
		expRes string
		expErr bool
	}{
		{
		"convert 1000000uregen to 1regen",
	    types.NewCoin("uregen", types.NewInt(int64(1000000))),
		"1regen",
		false, 
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res, err := dvr.Format(tc.coin)
			require.NoError(t, err)
			require.Equal(t, tc.expRes, res)
		})
	}
}


func TestFormatInt(t *testing.T) {

	d := valuerenderer.NewDefaultValueRenderer()
	decimal, _ := types.NewDecFromStr("349383323.894")
	i, _ := types.NewIntFromString("2323293999402003")
	
	
	// TODO consider add panic case and lens(strs) > 2 
	tt := []struct{
		name string
		input interface{}
		expRes string
		expErr bool
	}{
		{"nil", nil, "",  true},
		{"convert a million, no error", types.NewInt(int64(1000000)),"1,000,000", false},
		{"empty string error", types.Int{}, "", true},
		{"Decimal, no error", decimal, "349,383,323.894", false},
		{"Int, no error", i, "232,329,399,940,200,3", false},

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

			require.Equal(t, tc.expRes, res)
		})
	}
}



func TestParseString(t *testing.T) {
   re := regexp.MustCompile(`\d+[mu]?regen`)
   dvr := valuerenderer.NewDefaultValueRenderer()

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
		    x, err := dvr.Parse(tc.str)
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








