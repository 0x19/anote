VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
PROJECTNAME := $(shell basename "$(PWD)")

# Go related variables.
GOBASE := $(shell pwd)
GOPATH := $(GOBASE)/vendor:$(GOBASE)
GOBIN := $(GOBASE)/bin
GOCMDROOT := $(GOBASE)/cmd/anote
GOFILES := $(wildcard $(GOCMDROOT)/*.go)

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

##build:  Executes build instructions for anote project 
build:
	go build $(LDFLAGS) -o $(GOBIN)/$(PROJECTNAME) $(GOFILES)

##test: Tests the codebase
test:
	go test -v -cover ./...

.PHONY: help
all: help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo