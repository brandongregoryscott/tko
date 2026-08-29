.PHONY: build run test format format-check vet clean coverage coverage-html

build:
	go build ./cmd/tko

run:
	go run ./cmd/tko

test:
	go test ./...

format:
	go fmt ./...

format-check:
	@files=$$(gofmt -l .); \
	if [ -n "$$files" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$files"; \
		exit 1; \
	fi

vet:
	go vet ./...

clean:
	go clean

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out
