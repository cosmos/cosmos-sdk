package v2

const (
	ModuleName = "group"

	// Group Table
	GroupTablePrefix    byte = 0x0
	GroupTableSeqPrefix byte = 0x1

	// Group Member Table
	GroupMemberTablePrefix byte = 0x10

	// Group Policy Table
	GroupPolicyTablePrefix    byte = 0x20
	GroupPolicyTableSeqPrefix byte = 0x21

	// Proposal Table
	ProposalTablePrefix    byte = 0x30
	ProposalTableSeqPrefix byte = 0x31

	// Vote Table
	VoteTablePrefix byte = 0x40
)
