.PHONEY: proto test

proto:
	protoc --gogo_out=. *.proto

test: proto
	go install .
