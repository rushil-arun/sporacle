.PHONY: test style tst vet fmt

fmt: 
	@test -z "$$(gofmt -s -l .)"

vet: 
	go vet ./...

tst: 
	go test ./tst/...

style: fmt vet

test: tst