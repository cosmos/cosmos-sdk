@0xb02b1ca217480de7;

 using Go = import "/go.capnp";
 $Go.package("types");
 $Go.import("github.com/cosmos/cosmos-sdk/types");


using Int = Data;
using AccAddress = Data;
using Coins = List(CoinE);

struct CoinE {
  denom @0 :Text;
  amount @1 :Text;
}

struct StdTx(Msg, PubKey) {
  msg @0 :Msg;
  fee @1 :StdFee;
  signatures @2 :List(StdSignature(PubKey));
  memo @3 :Text;
}

struct StdFee {
  coins @0 :Coins;
  gas @1 :UInt64;
}

struct StdSignature(PubKey) {
  pubkey @0 :PubKey;
  signature @1 :Data;
}