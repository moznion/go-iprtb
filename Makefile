.PHONY: vendor

check: fmt-check lint test
ci-check: fmt-check test

test: vendor
	go test ./... -race -v -coverprofile="coverage.txt" -covermode=atomic

fmt:
	gofmt -w -s *.go && goimports -w *.go

fmt-check:
	goimports -l *.go | grep [^*][.]go$$; \
		EXIT_CODE=$$?; \
		if [ $$EXIT_CODE -eq 0 ]; then exit 1; fi \

lint:
	golangci-lint run ./...

vendor:
	go mod vendor

