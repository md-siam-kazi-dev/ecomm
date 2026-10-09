.PHONY: run build test vet fmt tidy clean

BINARY := bin/api
MAIN := ./cmd/api

run:
	go run $(MAIN)

build:
	go build -o $(BINARY) $(MAIN)

test:
	go test ./... -race -count=1

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

clean:
	rm -rf bin
