<!--
order: 2
-->

# State

The paths for the values stored in state is defined [here](https://github.com/cosmos/ics/blob/master/spec/ics-024-host-requirements/README.md#path-space).
Additionally, the SDK adds a prefix to the path to be able to aggregate the values for querying purposes.
The client type is not stored since it can be obtained through the client state. 

| Prefix | Path                                                                        | Value type     |
|--------|-----------------------------------------------------------------------------|----------------|
| "0/"   | "clients/{identifier}/clientState"                                          | ClientState    |
| "0/"   | "clients/{identifier}/consensusStates/{height}"                             | ConsensusState |
| "0/"   | "clients/{identifier}/connections"                                          | []string       |
| "0/"   | "connections/{identifier}"                                                  | ConnectionEnd  |
| "0/"   | "ports/{identifier}"                                                        | CapabilityKey  |
| "0/"   | "channelEnds/ports/{identifier}/channels/{identifier}"                      | ChannelEnd     |
| "0/"   | "capabilities/ports/{identifier}/channels/{identifier}/key"                 | CapabilityKey  |
| "0/"   | "seqSends/ports/{identifier}/channels/{identifier}/nextSequenceSend"        | uint64         |
| "0/"   | "seqRecvs/ports/{identifier}/channels/{identifier}/nextSequenceRecv"        | uint64         |
| "0/"   | "seqAcks/ports/{identifier}/channels/{identifier}/nextSequenceAck"          | uint64         |
| "0/"   | "commitments/ports/{identifier}/channels/{identifier}/packets/{sequence}"   | bytes          |
| "0/"   | "receipts/ports/{identifier}/channels/{identifier}/receipts/{sequence}"     | bytes          |
| "0/"   | "acks/ports/{identifier}/channels/{identifier}/acknowledgements/{sequence}" | bytes          |
