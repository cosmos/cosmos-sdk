test-integration:
	go test ./integration/... -timeout 30m

test-integration-cov:
	go test ./integration/... -timeout 30m -coverpkg=../... -coverprofile=integration-profile.out -covermode=atomic

test-e2e: test-e2e-server
	go test ./e2e/... -mod=readonly -timeout 30m -tags='e2e'

test-e2e-cov: test-e2e-server
	go test ./e2e/... -mod=readonly -timeout 30m -race -tags='e2e' -coverpkg=../... -coverprofile=e2e-profile.out -covermode=atomic

test-e2e-server:
	go test ./e2e/server -mod=readonly -timeout 30m -tags='e2e' -coverpkg=../... -coverprofile=e2e-server-profile.out -covermode=atomic