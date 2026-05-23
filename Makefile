.PHONY: build run test format clean coverage coverage-html

build:
	go build ./cmd/tko

run:
	go run ./cmd/tko

test:
	go test ./...

format:
	go fmt ./...

clean:
	go clean

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
