package textual_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"cosmossdk.io/x/tx/textual"
	"github.com/stretchr/testify/require"
)

type ledgerTestCase struct {
	Name     string
	Tx       string
	Expected []string
	Expert   bool
}

func TestLedger(t *testing.T) {
	rawTxs, err := os.ReadFile("./internal/testdata/tx.json")
	require.NoError(t, err)
	var txTestcases []txJsonTest
	err = json.Unmarshal(rawTxs, &txTestcases)
	require.NoError(t, err)

	raw, err := os.ReadFile("./internal/testdata/ledger.json")
	require.NoError(t, err)
	var testcases []ledgerTestCase
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			// Find the matching proto Tx from tx.json
			var txTestCase txJsonTest
			for _, t := range txTestcases {
				if strings.HasPrefix(tc.Name, t.Name) {
					txTestCase = t
				}
			}
			_, bodyBz, _, authInfoBz, signerData := createTextualData(t, txTestCase.Proto, txTestCase.SignerData)

			tr := textual.NewTextual(mockCoinMetadataQuerier)
			ctx := addMetadataToContext(context.Background(), txTestCase.Metadata)
			signDoc, err := tr.GetSignBytes(ctx, bodyBz, authInfoBz, signerData)
			require.NoError(t, err)

			// Make sure CBOR matches.
			require.Equal(t, tc.Tx, hex.EncodeToString(signDoc))

			// Make sure screens match.
			require.Equal(t, tc.Expected, displayLedger(txTestCase.Screens, tc.Expert))
		})
	}
}

const maxNumberOfChars = 74

// displayLedger is an attempt to match the implementation of how the ledger
// device processes the given screens. It should match the implementation in
// https://github.com/cosmos/ledger-cosmos/tree/dev_sign_mode_textual
func displayLedger(screens []textual.Screen, showExpert bool) []string {
	var res []string
	for _, s := range screens {
		if s.Expert && !showExpert {
			continue
		}

		// If text ends with @, add another @
		k, v := toAscii(s.Title), toAscii(s.Content)

		innerResLen := int(math.Ceil(float64(len(v)) / float64(maxNumberOfChars)))
		innerRes := make([]string, innerResLen)
		for i := range innerRes {
			// Indent
			if s.Indent > 0 {
				innerRes[i] += strings.Repeat(">", s.Indent) + " "
			}

			// Key (aka title)
			innerRes[i] += k

			// Pagination
			if innerResLen > 1 {
				innerRes[i] += fmt.Sprintf(" [%d/%d]", i+1, innerResLen)
			}

			if k != "" {
				innerRes[i] += ": "
			}

			// Value
			end := int(math.Min(float64(i+1)*maxNumberOfChars, float64(len(v))))
			innerRes[i] += v[i*maxNumberOfChars : end]

		}

		res = append(res, innerRes...)
	}

	// Add number in front of ledger screens
	for i := range res {
		res[i] = fmt.Sprintf("%d | %s", i, res[i])
	}

	return res

}

func toAscii(s string) string {
	// Add an additional @ if string ends with @ or " "
	if strings.HasSuffix(s, "@") || strings.HasSuffix(s, " ") {
		s += "@"
	}

	// Unicode to ascii
	s = strconv.Quote(s)
	s = strings.Trim(s, `"`)

	return s
}
