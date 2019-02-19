# Specification of Specifications

This file intends to outline the common structure for specifications within
this directory. 

## Tense

For consistency, specs should be written in passive present tense.

## Pseudo-Code

Pseudo-code should be avoided throughout the spec, instead bulleted-lists should 
be used to describe with words all operations which occur. 

## Common Layout

The following generalized structure should be used to breakdown specifications
for modules. With the exception of README.md, `XX` at the beginning of the file
name should be replaced with a number to indicate document flow (ex. read
`01_state.md` before `02_state_transitions.md`). Note that not all files may be
required depending on the modules function.

 - `README.md` - overview of the module
 - `XX_concepts.md` - describe specialized concepts and definitions used throughout the spec
 - `XX_state.md` - specify and describe structures expected to marshalled into the store, and their keys
 - `XX_state_transitions.md` - standard state transition operations triggered by hooks, messages, etc. 
 - `XX_messages.md` - specify message structure(s) and expected state machine behaviour(s)
 - `XX_begin_block.md` - specify any begin-block operations
 - `XX_end_block.md` - specify any end-block operations
 - `XX_hooks.md` - describe available hooks to be called by/from this module 
 - `XX_tags.md` - list and describe event tags used
 - `XX_future_improvements.md` - describe future improvements of this module
 - `XX_appendix.md` - supplementary details referenced elsewhere within the spec

### Notation for key-value mapping

Within `state.md` the following notation `->` should be used to describe key to
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
