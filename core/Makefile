proto-gen:
	(cd internal/example; buf generate)
	# update-proto-pin
	(cd internal/example; buf build --exclude-imports -o proto_image.bin.gz)

.PHONY: proto-gen
