 @0xf6e87acf2c3fc2e7;

 using Go = import "/go.capnp";
 $Go.package("types");
 $Go.import("github.com/cosmos/cosmos-sdk/x/gov");


using Int = Data;
using AccAddress = Data;
using Coins = List(Coin);

struct Coin {
  denom @0 :Text;
  amount @1 :Int;
}

struct MsgSubmitProposal(Content) {
  content @0 :Content;
  initialDeposit @1 :Coins;
  proposer @2 :AccAddress;
}

struct MsgDeposit {
  proposalId @0 :UInt64;
  depositor @1 :AccAddress;
  amount @2 :Coins;
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
  voter @1 :AccAddress;
  option @2 :VoteOption;
}

struct TextProposal {
  title @0 :Text;
  description @1 :Text;
}

struct Tx(Content) {
  msgDeposit @0 :MsgDeposit;
  msgSubmitProposal @1 :MsgSubmitProposal;
  msgVote @2 :MsgVote;
}
