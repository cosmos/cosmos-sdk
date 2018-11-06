# IBC Spec

## MVP3

`IBCOpenMsg` is added to open the connection between two chains. Also, `IBCUpdateMsg` is added, making it able to prove the header.

### IBC Module


// Implements sdk.Msg
type IBCTransferMsg struct {
    Packet
}

// Implements sdk.Msg
type IBCReceiveMsg struct {
    Packet
}

// Internal API



```golang
// User facing API

type Packet struct {
    Data      Payload
    SrcChain  string
    DestChain string
}

type Payload interface {
    Type() string
    ValidateBasic() sdk.Error
}

type TransferPayload struct {
    SrcAddr  sdk.Address
    DestAddr sdk.Address
    Coins    sdk.Coins
}

// Implements sdk.Msg
type IBCTransferMsg struct {
    Packet
}

// Implements sdk.Msg
type IBCReceiveMsg struct {
    Packet
    Proof           iavl.Proof
    FromChainID     string
    FromChainHeight uint64
}

type RootOfTrust struct {
    // 
}

// Implements sdk.Msg
type IBCOpenMsg struct {
    ROT   RootOfTrust
    Chain string   
}

// Implements sdk.Msg
type IBCUpdateMsg struct {
    Header tm.Header
    Commit tm.Commit
}

// Internal API

type rule struct {
    r string
    f func(sdk.Context, IBCPacket) sdk.Result
}

type Dispatcher struct {
    rules []rule
}

func NewHandler(dispatcher Dispatcher, ibcm IBCMapper) sdk.Handler

type IBCMapper struct {
    ibcKey sdk.StoreKey // IngressKey / EgressKey / HeaderKey => Value
                        // ChannelID              => last income msg's sequence
                        // (ChannelID, Msg index) => length / indexed msg
                        // ChannelID              => last known header
}

type IngressKey struct {
    ChannelID uint64
}

type EgressKey struct {
    ChannelID uint64
    Index     int64
}

type HeaderKey struct {
    ChannelID uint64
}

// Used by other modules
func (ibcm IBCMapper) PushPacket(ctx sdk.Context, dest string, payload Payload)

```


