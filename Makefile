.PHONY: all

version := $(shell git describe --tags --abbrev=64)
dirty := ''
ifneq (,$(shell git diff --name-only))
dirty =-dirty
endif
module_root := github.com/wmentor/shrun
ld_flags := -X $(module_root)/internal/common.Version=$(version)$(dirty)

all: build install

build:
	mkdir -p ./bin
	go mod tidy
	go build -ldflags "$(ld_flags)" -o ./bin/shrun cmd/shrun/main.go

install:
	mkdir -p $(GOPATH)/bin/
	cp -f ./bin/shrun $(GOPATH)/bin/
