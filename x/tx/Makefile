codegen:
	@(cd internal/testpb; buf mod update; buf generate)
	@(cd signing/aminojson/internal; make codegen)