<!--
order: 5
-->

# Client

## CLI

A user can query and interact with the `crisis` module using the CLI.

### Transactions

The `tx` commands allow users to interact with the `crisis` module.

```bash
simd tx crisis --help
```

#### invariant-broken

The `invariant-broken` command submits proof when an invariant was broken to halt the chain

```bash
simd tx crisis invariant-broken [module-name] [invariant-route] [flags]
```

Example:

```bash
simd tx crisis invariant-broken bank total-supply --from=[keyname or address]
```
