## Appendix B: IBC Queue Format

([Back to table of contents](specification.md#contents))

The foundational data structure of the IBC protocol are the message queues stored inside each chain. We start with a well-defined binary representation of the keys and values used in these queues. The encodings mirror the semantics defined above:

_key = _(_remote id, [send|receipt], [head|tail|index])_

_V<sub>send</sub> = (maxHeight, maxTime, type, data)_

_V<sub>receipt</sub> = (result, [success|error code])_


```
 message QueueName {
  // chain_id is which chain this queue is
  // associated with
  string chain_id = 1;
  enum Purpose {
          SEND = 0;
          RECEIPT = 1;
  }
  Purpose purpose = 2;
 }
 // StateKey is a key for the head/tail of a given queue
 message StateKey {
  QueueName queue = 1;
  // both encode into one byte with varint encoding
  // never clash with 8 byte message indexes
  enum State {
          HEAD = 0;
          TAIL = 0x7f;
  }
  State state = 2;
 }
 // StateValue is the type stored under a StateKey
 message StateValue {
    fixed64 index = 1;
 }
 // MessageKey is the key for message *index* in a given queue
 message MessageKey {
  QueueName queue = 1;
  fixed64 index = 2;
 }
 // SendValue is stored under a MessageKey in the SEND queue
 message SendValue {
  uint64 maxHeight = 1;
  google.protobuf.Timestamp maxTime = 2;
  // use kind instead of type to avoid keyword conflict
  bytes kind = 3;
  bytes data = 4;
 }
 // ReceiptValue is stored under a MessageKey in the RECEIPT queue
 message ReceiptValue {
  // 0 is success, others are application-defined errors
  int32 errorCode = 1;
  // contains result on success, optional info on error
  bytes data = 2;
 }
```

Keys and values are binary encoded and stored as bytes in the merkle tree in order to generate the root hash stored in the block header, which validates all proofs. They are treated as arrays of bytes by the merkle proofs for deterministically generating the sequence of hashes, and passed as such in all interchain messages. Once the validity of a key value pair has been determined from the merkle proof and header, the bytes can be deserialized and the contents interpreted by the protocol.
