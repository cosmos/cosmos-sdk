package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// The Electionator abstraction covers the concept space for
// a wide variety of election kinds.
type Electionator interface {

	// is the election object accepting votes.
	Active() bool

	// functionality to execute for when a vote is cast in this election, here
	// the vote field is anticipated to be marshalled into a vote type used
	// by an election.
	//
	// NOTE There are no explicit ids here. Just votes which pertain specifically
	// to one electionator. Anyone can create and send a vote to the electionator item
	// which will presumably attempt to marshal those bytes into a particular struct
	// and apply the vote information in some arbitrary way. There can be multiple
	// Electionators within the Cosmos-Hub for multiple specialization groups, votes
	// would need to be routed to the Electionator upstream of here.
	Vote([]byte) error

	// here lies all functionality to authenticate and execute changes for
	// when a member accepts being elected
	AcceptElection(sdk.AccAddress)

	// Register a revoker object
	RegisterRevoker(Revoker)

	// No more revokers may be registered after this function is called
	SealRevokers()

	// register hooks to call when an election actions occur
	RegisterHooks(ElectionatorHooks)

	// query for the current winner(s) of this election based on arbitrary
	// election ruleset
	QueryWinners() []sdk.AccAddress

	// query metadata for an address in the election this
	// could include for example position that an address
	// is being elected for within a group
	//
	// this metadata may be directly related to
	// voting information and/or privileges enabled
	// to members within a group.
	QueryMetadata(sdk.AccAddress) []byte
}

// ElectionatorHooks, once registered with an Electionator,
// trigger execution of relevant interface functions when
// Electionator events occur.
type ElectionatorHooks interface {
	AfterVoteCast(addr sdk.AccAddress, vote []byte)
	AfterMemberAccepted(addr sdk.AccAddress)
	AfterMemberRevoked(addr sdk.AccAddress, cause []byte)
}

// Revoker defines the function required for an membership revocation rule-set
// used by a specialization group. This could be used to create self revoking,
// and evidence based revoking, etc. Revokers types may be created and
// reused for different election types.
//
// When revoking the "cause" bytes may be arbitrarily marshalled into evidence,
// memos, etc.
type Revoker interface {
	RevokeName() string // identifier for this revoker type
	RevokeMember(addr sdk.AccAddress, cause []byte) (successful bool)
}

// The specialization group abstraction firstly extends the Electionator
// but also further defines traits of the group.
type SpecializationGroup interface {
	Electionator
	GetName() string
	GetDescription() string

	// general soft contract the group is expected
	// to fulfill with the greater community
	GetContract() string

	// messages which can be executed by the members of the group
	Handler(ctx sdk.Context, msg sdk.Msg) sdk.Result

	// logic to be executed at endblock, this may for instance
	// include payment of a stipend to the group members
	// for participation in the security group.
	EndBlocker(ctx sdk.Context)
}
