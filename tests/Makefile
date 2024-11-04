test-integration:
	go test ./integration/... -timeout 30m

test-integration-v1:
	go test -tags='app_v1' ./integration/... -timeout 30m

test-integration-cov:
	go test ./integration/... -timeout 30m -coverpkg=../... -coverprofile=integration-profile.out -covermode=atomic
