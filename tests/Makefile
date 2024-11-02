test-integration:
	go test ./integration/... -timeout 30m

test-integration-cov:
	go test ./integration/... -timeout 30m -coverpkg=../... -coverprofile=integration-profile.out -covermode=atomic
