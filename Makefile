.PHONY: build run test

CONFIG ?= tests/example.ini

build:
	mkdir -p bin
	go build -o bin/visualiser ./src/cmd/visualiser

run: build
	./bin/visualiser -config "$(CONFIG)"

test:
	go test ./...
