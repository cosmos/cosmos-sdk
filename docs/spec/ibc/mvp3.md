# IBC Spec

## MVP3

`IBCOpenMsg` is added to open the connection between two chains. Also, `IBCUpdateMsg` is added, making it able to prove the header.

### IBC Module

```golang
// User facing API

type IBCTransferData struct {
    SrcAddr  sdk.Address
    DestAddr sdk.Address
    Coins    sdk.Coins
}

// Implements ibc.PacketData
type IBCPacket struct {
    IBCData
}

// Implements ibc.PacketData
type IBCReceivePacket struct {
    IBCData    
}

type Packet struct {
    Data      PacketData
    SrcChain  string    
    DestChain string
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
func (ibcm IBCMapper) PushPacket(ctx sdk.Context, dest string, data PacketData)

```


