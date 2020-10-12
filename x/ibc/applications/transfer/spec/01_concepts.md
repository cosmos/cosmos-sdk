<!--
order: 1
-->

# Concepts

## Acknowledgements

ICS20 uses the recommended acknowledgement format as specified by [ICS 04](https://github.com/cosmos/ics/tree/master/spec/ics-004-channel-and-packet-semantics#acknowledgement-envelope).

A successful receive of a transfer packet will result in a Result Acknowledgement being written
with the value `[]byte(byte(1))` in the `Response` field. 

An unsuccessful receive of a transfer packet will result in an Error Acknowledgement being written
with the error message in the `Response` field. 
