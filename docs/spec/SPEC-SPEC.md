# Specification of Specifications

This file intends to outline the common structure for specifications within
this directory.

## Tense

For consistency, specs should be written in passive present tense.

## Pseudo-Code

Generally, pseudo-code should be minimized throughout the spec. Often, simple
bulleted-lists which describe a function's operations are sufficient and should
be considered preferable. In certain instances, due to the complex nature of
the functionality being described pseudo-code may the most suitable form of
specification. In these cases use of pseudo-code is permissible, but should be
presented in a concise manner, ideally restricted to only the complex
element as a part of a larger description.

### Notation for key-value mapping

Within `state.md` the following notation `->` should be used to describe key to
value mapping:

```text
key -> value
```

to represent byte concatenation the `|` may be used. In addition, encoding
type may be specified, for example:

```text
0x00 | addressBytes | address2Bytes -> amino(value_object)
```

Additionally, index mappings may be specified by mapping to the `nil` value, for example:

```text
0x01 | address2Bytes | addressBytes -> nil
```
