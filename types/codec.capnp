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

