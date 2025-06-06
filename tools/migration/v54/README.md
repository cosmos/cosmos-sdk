# v54 Migration

This code provides a simple binary to make all necessary changes to a codebase to update an application to use comet v1.0.1 and Cosmos SDK v54.

# Installation

## Local

```shell
cd tools/migration/v54
go install .
```

# Execution

Running in current directory:
```shell
migration .
```

Running in specified directory

```shell
migration path/to/dir
```