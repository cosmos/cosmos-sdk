# Git hooks

Installation:

```
$ git config core.hooksPath contrib/githooks
```

## pre-commit

The hook automatically runs `gofmt`, `goimports`, and `misspell`
to correctly format the `.go` files included in the commit.
