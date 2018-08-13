## Makefile

The [Makefile](https://en.wikipedia.org/wiki/Makefile) compiles the Go program by defining a set of rules with targets and recipes. We'll need to add our application commands to it:

```
// Makefile
build_examples:
ifeq ($(OS),Windows_NT)
	...
	go build $(BUILD_FLAGS) -o build/simplegovd.exe ./examples/simpleGov/cmd/simplegovd
	go build $(BUILD_FLAGS) -o build/simplegovcli.exe ./examples/simpleGov/cmd/simplegovcli
else
	...
	go build $(BUILD_FLAGS) -o build/simplegovd ./examples/simpleGov/cmd/simplegovd
	go build $(BUILD_FLAGS) -o build/simplegovcli ./examples/simpleGov/cmd/simplegovcli
endif
...
install_examples:
    ...
	go install $(BUILD_FLAGS) ./examples/simpleGov/cmd/simplegovd
	go install $(BUILD_FLAGS) ./examples/simpleGov/cmd/simplegovcli
```