package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/upgrade/types"

	"github.com/cosmos/cosmos-sdk/codec"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type ProposalWrapper struct {
	Prop gov.Content
}

func TestContentAccessors(t *testing.T) {
	cases := map[string]struct {
		p     gov.Content
		title string
		desc  string
		typ   string
		str   string
	}{
		"upgrade": {
			p: types.NewSoftwareUpgradeProposal("Title", "desc", types.Plan{
				Name:   "due_height",
				Info:   "https://foo.bar",
				Height: 99999999999,
			}),
			title: "Title",
			desc:  "desc",
			typ:   "SoftwareUpgrade",
			str:   "title:\"Title\" description:\"desc\" plan:<name:\"due_height\" time:<seconds:-62135596800 > height:99999999999 info:\"https://foo.bar\" > ",
		},
		"cancel": {
			p:     types.NewCancelSoftwareUpgradeProposal("Cancel", "bad idea"),
			title: "Cancel",
			desc:  "bad idea",
			typ:   "CancelSoftwareUpgrade",
			str:   "title:\"Cancel\" description:\"bad idea\" ",
		},
	}

	cdc := codec.NewLegacyAmino()
	gov.RegisterLegacyAminoCodec(cdc)
	types.RegisterLegacyAminoCodec(cdc)

	for name, tc := range cases {
		tc := tc // copy to local variable for scopelint
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.title, tc.p.GetTitle())
			assert.Equal(t, tc.desc, tc.p.GetDescription())
			assert.Equal(t, tc.typ, tc.p.ProposalType())
			assert.Equal(t, "upgrade", tc.p.ProposalRoute())
			assert.Equal(t, tc.str, tc.p.String())

			// try to encode and decode type to ensure codec works
			wrap := ProposalWrapper{tc.p}
			bz, err := cdc.Marshal(&wrap)
			require.NoError(t, err)
			unwrap := ProposalWrapper{}
			err = cdc.Unmarshal(bz, &unwrap)
			require.NoError(t, err)

			// all methods should look the same
			assert.Equal(t, tc.title, unwrap.Prop.GetTitle())
			assert.Equal(t, tc.desc, unwrap.Prop.GetDescription())
			assert.Equal(t, tc.typ, unwrap.Prop.ProposalType())
			assert.Equal(t, "upgrade", unwrap.Prop.ProposalRoute())
			assert.Equal(t, tc.str, unwrap.Prop.String())
		})

	}
}
