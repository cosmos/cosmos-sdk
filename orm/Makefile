.PHONY: proto

DIRECTORIES_TO_BUILD := "./apis ./internal/testproto ./internal/kvindexer"
proto:
	docker build -t orm-proto:dev .
	docker run -v "$(CURDIR):/genproto" -w /genproto orm-proto:dev ./scripts/genproto.sh "$(DIRECTORIES_TO_BUILD)"
