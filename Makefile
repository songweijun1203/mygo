.PHONY: build run test fmt clean

build:
	go build -o bin/mygo ./cmd/mygo

run:
	go run ./cmd/mygo

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	go clean
	rm -rf bin
