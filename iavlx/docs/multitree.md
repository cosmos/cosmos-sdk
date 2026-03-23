# Multi-Tree

An app's root state store is actually a "multi-tree" of multiple iavl trees.

Each iavl tree is stored in its own directory within the main iavl data dir under
the `stores/` and with the `.iavl` suffix, ex:

```
└── stores/
│   └── auth.iavl
│   └── bank.iavl
│   └── staking.iavl
```