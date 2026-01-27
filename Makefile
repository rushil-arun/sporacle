.PHONY: ci fmt vet test

fmt: 
	@test -z "$$(gofmt -s -l .)"

vet: 
	go vet ./...

test: 
	go test ./...

ci: fmt vet test