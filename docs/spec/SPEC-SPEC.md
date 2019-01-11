# Specification of Specifications

This file intends to outline the common structure for specifications within
this directory. 


## Common Layout

The following generalized structure should be used to breakdown specifications
for modules. Note that not all files may be required depending on the modules 
function

 - `overview.md` - describe module
 - `state.md` - specify and describe structures expected to marshalled into the store, and their keys
 - `hooks.md` - describe available hooks to be called by/from this module 
 - `end_block.md` - specify any end-block operations
 - `begin_block.md` - specify any begin-block operations
 - `tags.md` - list and describe event tags used
 - `transactions.md` - specify transaction structure and expected state machine behaviour 

### Notation for key-value mapping

Within `state.md` the following notation `->` should be used to describe key to
value mapping:

```
key -> value 
```

to represent the byte concatonation the `|` may be used. In addition, encoding
type may be specified, for example:

```
0x00 | addressBytes | address2Bytes -> amino(value_object) 
```

Additionally, Index mappings may be specified by mapping to the `nil` value, for example:

```
0x01 | address2Bytes | addressBytes -> nil
```
