package params

// Default simulation operation weights for messages and gov proposals
const (
	DefaultWeightMsgSend         int = 100
	DefaultWeightMsgMultiSend    int = 10
	DefaultWeightMsgDeposit      int = 100
	DefaultWeightMsgVote         int = 67
	DefaultWeightMsgVoteWeighted int = 33

	// feegrant
	DefaultWeightGrantAllowance  int = 100
	DefaultWeightRevokeAllowance int = 100
)
