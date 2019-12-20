 @0xf6e87acf2c3fc2e7;

 using Go = import "/go.capnp";
 $Go.package("types");
 $Go.import("github.com/cosmos/cosmos-sdk/x/gov/types");

 using SDK = import "/types/codec.capnp";

struct Msg(Content) {
  msgDeposit @0 :MsgDeposit;
  msgSubmitProposal @1 :MsgSubmitProposal;
  msgVote @2 :MsgVote;
}

struct MsgSubmitProposal(Content) {
  content @0 :Content;
  initialDeposit @1 :SDK.Coins;
  proposer @2 :SDK.AccAddress;
}

struct MsgDeposit {
  proposalId @0 :UInt64;
  depositor @1 :SDK.AccAddress;
  amount @2 :SDK.Coins;
}

enum VoteOption {
  empty @0;
  yes @1;
  abstain @2;
  no @3;
  noWithVeto @4;
}

struct MsgVote {
  proposalId @0 :UInt64;
  voter @1 :SDK.AccAddress;
  option @2 :VoteOption;
}

struct TextProposal {
  title @0 :Text;
  description @1 :Text;
}

