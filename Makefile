.PHONY: build run test clean

build:
	go build -o spider ./cmd/spider

run:
	go run ./cmd/spider

test:
	go test ./...

clean:
	rm -f spider spider.exe
