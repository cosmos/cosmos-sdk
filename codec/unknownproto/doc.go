/*
unknownproto implements functionality to "type check" protobuf serialized byte sequences
against an expected proto.Message to report:

a) Unknown fields in the stream -- this is indicative of mismatched services, perhaps a malicious actor

b) Mismatched wire types for a field -- this is indicative of mismatched services

Its API signature is similar to proto.Unmarshal([]byte, proto.Message) as

    if err := unknownproto.CheckMismatchedFields(protoBlob, protoMessage); err != nil {
            // Handle the error.
    }

and ideally should be added before invoking proto.Unmarshal, if you'd like to enforce the features mentioned above.
*/
package unknownproto
