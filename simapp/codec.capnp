@0xa8444e9342b047f5;

using Go = import "/go.capnp";
$Go.package("simapp");
$Go.import("github.com/cosmos/cosmos-sdk/simapp");

using SDK = import "/types/codec.capnp";
using Gov = import "/x/gov/types/codec.capnp";

struct Msg {
    union {
        gov @0 :Gov.Msg(SimappProposal);
    }
}

struct SimappProposal {
    union {
        textProposal @0 :Gov.TextProposal;
    }
}

struct PubKey {
    union {
        secp256k1 @0 :Data;
        ed25519 @1 :Data;
    }
}

using Tx = SDK.StdTx(List(Msg), PubKey);
