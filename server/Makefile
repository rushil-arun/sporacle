.PHONY: test style tst vet fmt gen db-migrate

gen:
	go run ./scripts/gen-constants

db-migrate:
	go run ./scripts/migrate-trivia

fmt:
	@test -z "$$(gofmt -s -l .)"

vet: 
	go vet ./...

tst: 
	go test ./tst/...

style: fmt vet

test: tst