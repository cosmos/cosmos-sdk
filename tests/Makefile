test-integration:
	go test ./integration/... -timeout 30m -coverprofile=integration-profile.out -covermode=atomic

test-e2e:
	go test ./e2e/... -mod=readonly -timeout 30m -race -tags='e2e'
