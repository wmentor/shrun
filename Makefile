.PHONY: all

all: build install

build:
	mkdir -p ./bin
	go mod tidy
	go build -o ./bin/shrun cmd/shrun/main.go

install:
	cp -f ./bin/shrun $(GOPATH)/bin/
