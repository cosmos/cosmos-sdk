---
order: false
parent:
  order: false
---

# Architecture Decision Records (ADR)

This is a location to record all high-level architecture decisions in the Cosmos-SDK.

You can read more about the ADR concept in this [blog post](https://product.reverb.com/documenting-architecture-decisions-the-reverb-way-a3563bb24bd0#.78xhdix6t).

An ADR should provide:

- Context on the relevant goals and the current state
- Proposed changes to achieve the goals
- Summary of pros and cons
- References
- Changelog

Note the distinction between an ADR and a spec. The ADR provides the context, intuition, reasoning, and
justification for a change in architecture, or for the architecture of something
new. The spec is much more compressed and streamlined summary of everything as
it stands today.

If recorded decisions turned out to be lacking, convene a discussion, record the new decisions here, and then modify the code to match.

Note the context/background should be written in the present tense.

Please add a entry below in your Pull Request for an ADR.

## ADR Table of Contents

- [ADR 002: SDK Documentation Structure](./adr-002-docs-structure.md)
- [ADR 003: Dynamic Capability Store](./adr-003-dynamic-capability-store.md)
- [ADR 004: Split Denomination Keys](./adr-004-split-denomination-keys.md)
- [ADR 006: Secret Store Replacement](./adr-006-secret-store-replacement.md)
- [ADR 009: Evidence Module](./adr-009-evidence-module.md)
- [ADR 010: Modular AnteHandler](./adr-010-modular-antehandler.md)
- [ADR 011: Generalize Genesis Accounts](./adr-011-generalize-genesis-accounts.md)
- [ADR 012: State Accessors](./adr-012-state-accessors.md)
- [ADR 015: IBC Packet Receiver](./adr-015-ibc-packet-receiver.md)
- [ADR 016: Validator Consensus Key Rotation](./adr-016-validator-consensus-key-rotation.md)
- [ADR 017: Historical Header Module](./adr-017-historical-header-module.md)
