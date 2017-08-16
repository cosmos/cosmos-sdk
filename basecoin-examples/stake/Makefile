all: get_vendor_deps test install

test:
	go test .

install:
	go install ./cmd/...

get_vendor_deps:
	go get github.com/Masterminds/glide
	glide install
