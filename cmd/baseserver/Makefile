all: install linux darwin windows

LINKER_FLAGS="-X main.CommitHash=`git rev-parse HEAD`"

linux:
	CGO_ENABLED=0 GOOS=linux go build -ldflags $(LINKER_FLAGS) -o ./bin/baseserver_linux

windows:
	CGO_ENABLED=0 GOOS=windows go build -ldflags $(LINKER_FLAGS) -o ./bin/baseserver_windows

darwin:
	CGO_ENABLED=0 GOOS=darwin go build -ldflags $(LINKER_FLAGS) -o ./bin/baseserver_darwin

install:
	go get -ldflags $(LINKER_FLAGS)
