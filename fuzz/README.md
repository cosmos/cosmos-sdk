# Fuzzing

## Installation

```
go get -u github.com/dvyukov/go-fuzz/go-fuzz github.com/dvyukov/go-fuzz/go-fuzz-build
```

## Preparing

Create own go package under `fuzz` directory, which contains the function that you want to fuzz. See `fuzz/types/parsecoin`
for example, which has example for how to generating corpus and directory structure.

## Running

```
$ cd types/parsecoin
$ go-fuzz-build
$ go-fuzz
```

## oss-fuzz build status

https://oss-fuzz-build-logs.storage.googleapis.com/index.html#cosmos-sdk
