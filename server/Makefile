.PHONY: test style tst vet fmt gen

gen:
	go run ./scripts/gen-constants

fmt:
	@test -z "$$(gofmt -s -l .)"

vet: 
	go vet ./...

tst: 
	go test ./tst/...

style: fmt vet

test: tst