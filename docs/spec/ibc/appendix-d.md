## Appendix D: Universal IBC Packets

([Back to table of contents](specification.md#contents))

The structures above can be used to define standard encodings for the basic IBC transactions that must be exposed by a blockchain: _IBCreceive_, _IBCreceipt_,_ IBCtimeout_, and _IBCcleanup_. As mentioned above, these are not complete transactions to be posted as is to a blockchain, but rather the "data" content of a transaction, which must also contain fees, nonce, and signatures. The other IBC transaction types _IBCregisterChain_, _IBCupdateHeader_, and _IBCchangeValidators_ are specific to the consensus engine and use unique encodings. We define the tendermint-specific format in the next section.

```
 // IBCPacket sends a proven key/value pair from an IBCQueue.
 // Depending on the type of message, we require a certain type
 // of key (MessageKey at a given height, or StateKey).
 //
 // Includes src_chain and src_height to look up the proper
 // header to verify the merkle proof.
 message IBCPacket {
  // chain id it is coming from
  string src_chain = 1;
  // height for the header the proof belongs to
  uint64 src_height = 2;
  // the message type, which determines what key/value mean
  enum MsgType {
          RECEIVE = 0;
          RECEIPT = 1;
          TIMEOUT = 2;
          CLEANUP = 3;
  }
  MsgType msgType = 3;
  bytes key = 4;
  bytes value = 5;
  // the proof of the message
  MerkleProof proof = 6;
 }
```

