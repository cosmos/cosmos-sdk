# ADR 051: Protobuf Versioning

## Changelog

* {date}: {changelog}

## Status

DRAFT

## Abstract

> "If you can't explain it simply, you don't understand it well enough." Provide a simplified and layman-accessible explanation of the ADR.
> A short (~200 word) description of the issue being addressed.

## Context

The official protocol buffers documentation describes [guidelines for updating existing protobuf messages](https://developers.google.com/protocol-buffers/docs/proto3#updating)
in a way that doesn't break existing code. [Buf](https://docs.buf.build/breaking/overview) takes this a step futher with
tooling support for preventing breaking changes with several levels of compatibility guarantees.

These existing guidelines, however, fall short of ensuring full client-server forward
and backward compatibility.

Recall that backward compatibility means that a reader with a newer schema 
may accept data written by a writer with an older schema. Whereas forward compatibility
means that a reader with an older schema may accept data written by a writer
with a newer schema.

Lets consider some concrete examples with blockchain clients and servers. Say
we have a message `MsgFoo` where a new field `baz` has just been added:
```protobuf
message MsgFoo {
  string sender = 1;
  uint32 bar = 2;
  int32 baz = 3;
}
```

We have three scenarios:
- **same client/server version**: no problem
- **newer server**: the server understands `baz`, but the client will never populate it
- **newer client**: the client may populate `baz`, but the server by default will ignore it

Of these three the **newer client** scenario is potentially problematic both for
the client and the server and none of the extant protobuf literature addresses
this.

In the server's case what if `baz` contains some important information
that should change the behavior of `MsgFoo`? To address this,
[ADR 020 Protobuf Transaction Encoding](./adr-020-protobuf-transaction-encoding.md)
proactively included a feature known as unknown field filtering or rejection
which will cause the server to reject a message with a field number that it does
not recognize (in this case `3` for the `baz` field).

But what about the client? Without any additional information, the client will
display a user interface that might cause the `baz` field to be populated
which will result in an error when the user tries to submit the transaction.
This is generally bad UX because the UI should proactively try to prevent the
user from trying to do things which simply can't be done.

In https://github.com/cosmos/cosmos-sdk/discussions/10406 four solutions were
considered to address this:

1. **Reflection:** rely on gRPC reflection services which can inform the client at runtime whether a field is supported or not
2. **Annotations:** introduce annotations into proto files which indicate when a given query/field was added
3. **Bump Package Version for Minor Changes:** bump the proto version when whenever new fields are added, but still support old proto versions as much as possible
4. **New Fields Go in New Package Versions:** similar to the (3) but just add new things to new proto versions without copying over everything. So v1 would have one layer of functionality, v2 would add some new or changed methods, etc.

## Comparison

### 1. Reflection

This solution breaks down pretty quickly as developers don't write code at
runtime. Without some way of knowing whether a field is part of some newer
revision of the protobuf files, they would not know whether to do some runtime
feature check on that field.

### 2. Annotations

### 3. Bump Package Version for Minor Changes

### 4. New Fields Go in New Package Versions

# Decision

---


In a discussion regarding minor updates to  @webmaster128 on some of these issues:

> In https://github.com/confio/cosmjs-types/pull/8/files you see the diff I get when upgrading v1beta1 types from 0.42 to 0.44. Now assume I override those files with the latest version and want to be compatible to 0.42-0.44 backends. I have the following questions:
>
> - How do I know if Accounts  query is available? Reflection I guess, ok, will try.
> - Do I get values in Metadata.name and Metadata.symbol?
> - Does the TotalSupply query support pagination?
> - Can I use PageRequest.reverse?
> - Does VersionInfo.cosmosSdkVersion contain a value?
> - Is MsgVoteWeighted available? Reflection I guess, ok, will try.
> - Can I use GetTxsEventRequest.orderBy?
> - Can I use SimulateRequest.txBytes?
>
> Those questions are the reason we have no 0.44 support yet, because I hesitate to override types.

Here are the solutions I can think of:

1. **Reflection:** rely on gRPC reflection and cosmos.base.reflection - this should provide all of the information at runtime and should be enabled on stargate+ plus chains. I’m not sure how the DevX for this will be? Maybe it’s a bit complicated?
2. **Annotations:** introduce annotations into proto files which indicate when a given query/field was added
3. **Bump Proto Version for Minor Changes:** bump the proto version when we are adding new things but still support old proto versions as much as possible - this way a client can develop against a single major version of an API and just check if that’s supported. with this you’d end up with scenarios where a chain might support v1 partially (because it’s being deprecated), and say v2 and v3 fully - and a reflection service could tell which API versions are supported. maybe this is preferable to gRPC reflection?
4. **New Functionality Goes in New Proto Versions:** similar to the (3) but just add new things to new proto versions without copying over everything. So v1 would have one layer of functionality, v2 would add some new or changed methods, etc.

## Decision

> This section describes our response to these forces. It is stated in full sentences, with active voice. "We will ..."
> {decision body}

## Consequences

> This section describes the resulting context, after applying the decision. All consequences should be listed here, not just the "positive" ones. A particular decision may have positive, negative, and neutral consequences, but all of them affect the team and project in the future.

### Backwards Compatibility

> All ADRs that introduce backwards incompatibilities must include a section describing these incompatibilities and their severity. The ADR must explain how the author proposes to deal with these incompatibilities. ADR submissions without a sufficient backwards compatibility treatise may be rejected outright.

### Positive

{positive consequences}

### Negative

{negative consequences}

### Neutral

{neutral consequences}

## Further Discussions

While an ADR is in the DRAFT or PROPOSED stage, this section should contain a summary of issues to be solved in future iterations (usually referencing comments from a pull-request discussion).
Later, this section can optionally list ideas or improvements the author or reviewers found during the analysis of this ADR.

## Test Cases [optional]

Test cases for an implementation are mandatory for ADRs that are affecting consensus changes. Other ADRs can choose to include links to test cases if applicable.

## References

* {reference link}
