What's where?
-------------

- `client/cli/tx.go`: client commands create-client, establish exists, no way to update for now
- `client/cli/query.go`: query the latest consensus state of the Tendermint node
- `client/utils/receive.go`: a helper function to relay a latest single packet is defined, merkle proof is disabled so height does not need to be specified
- `keeper/25-interface.go`: the main entry point for the ibc module, the developers can open channel, send and verify packets with the methods
- `keeper/msgs.go`: create client and open connection are defined, receiving parts may be implemented manually, channel opening may be implemented within the custom keeper logic
- `keeper/keeper.go`: NewKeeper is the constructor

How to use this module?
-----------------------

1. Define types that implement ibc.Packet
2. Define msg types that contain that type
3. Handler should call ibckeeper.Receive() to receive those packet msgs
4. Keeper can call ibckeeper.Send() to send the packets
5. Define client command that internally calls utils.GetRelayPacket
6. Call that command each time the packet is sent and relay it to the other chain

Questions?
----------

Ask @mossid (Joon) or @cwgoes (Christopher) in person or on Telegram.
