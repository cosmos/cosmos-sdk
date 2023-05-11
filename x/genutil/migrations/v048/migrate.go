package v048

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// Migrate migrates exported state from v0.47 to a v0.48 genesis state.
func Migrate(appState types.AppMap, clientCtx client.Context) (types.AppMap, error) {
	// v0.47 group genesis
	// "group": {
	// 	"group_seq": "0",
	// 	"groups": [],
	// 	"group_members": [],
	// 	"group_policy_seq": "0",
	// 	"group_policies": [],
	// 	"proposal_seq": "0",
	// 	"proposals": [],
	// 	"votes": []
	//   },

	// v0.48 group genesis
	// "group": {
	// 	"cosmos.group.v1.GroupInfo": [],
	// 	"cosmos.group.v1.GroupMember": [],
	// 	"cosmos.group.v1.GroupPolicyInfo": [],
	// 	"cosmos.group.v1.Proposal": [],
	// 	"cosmos.group.v1.Vote": []
	//   },

	return appState, nil
}
