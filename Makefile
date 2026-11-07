.PHONY: all build clean run test build-linux

APP_NAME = neopanel
SRC = ./cmd/neopanel

all: build

build:
	go build -ldflags="-s -w" -o bin/$(APP_NAME) $(SRC)

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(APP_NAME)-linux-amd64 $(SRC)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/$(APP_NAME)-linux-arm64 $(SRC)

run:
	go run $(SRC)

clean:
	rm -rf bin/ data/

