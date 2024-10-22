BINARY_NAME=p-torrent

build:
	go build -o bin/$(BINARY_NAME) cmd/main.go

run:
	./bin/$(BINARY_NAME)

test:
	go test -v ./...

all : build run

clean:
	rm -rf bin/$(BINARY_NAME)
