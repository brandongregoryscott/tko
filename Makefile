.PHONY: build run test format clean coverage coverage-html check-release tag release

TAG ?=

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

tag:
	@test -n "$(TAG)" || (echo "Usage: make tag TAG=v0.1.0"; exit 1)
	git tag -a $(TAG) -m "$(TAG)"
	git push origin $(TAG)
