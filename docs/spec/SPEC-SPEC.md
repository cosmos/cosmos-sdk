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

## Common Layout

The specifications should be contained in a single `README.md` file inside the
`spec/` folder of a given module.

The following generalized document structure should be used to breakdown
specifications for modules. Each bullet item corresponds to a new section in
the document, and should begin with a secondary heading (`## {HEADING}` in
Markdown). The `XX` at the beginning of the section name should be replaced
with a number to indicate document flow (ex. read `01. Concepts` before
`02. State Transitions`). The following list is nonbinding and all sections are
optional.

- `XX. Abstract` - overview of the module
- `XX. Concepts` - describe specialized concepts and definitions used throughout the spec
- `XX. State` - specify and describe structures expected to marshalled into the store, and their keys
- `XX. State Transitions` - standard state transition operations triggered by hooks, messages, etc.
- `XX. Messages` - specify message structure(s) and expected state machine behaviour(s)
- `XX. BeginBlock` - specify any begin-block operations
- `XX. EndBlock` - specify any end-block operations
- `XX. Hooks` - describe available hooks to be called by/from this module
- `XX. Events` - list and describe event tags used
- `XX. Params` - list all module parameters, their types (in JSON) and examples
- `XX. Future Improvements` - describe future improvements of this module
- `XX. Appendix` - supplementary details referenced elsewhere within the spec

### Notation for key-value mapping

Within the `State` section, the following notation `->` should be used to describe key to
value mapping:

```
key -> value
```

to represent byte concatenation the `|` may be used. In addition, encoding
type may be specified, for example:

```
0x00 | addressBytes | address2Bytes -> amino(value_object)
```

Additionally, index mappings may be specified by mapping to the `nil` value, for example:

```
0x01 | address2Bytes | addressBytes -> nil
```
