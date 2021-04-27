COMMIT_ID=$(shell git rev-parse --short HEAD)
VERSION=$(shell cat VERSION)

NAME=odoh-client

DIRS_NEEDED=dataset
FILENAME="dataset/tranco-1m.zip"

all: clean build

clean:
	@echo "Cleaning and removing the odoh-client ..."
	@rm -f odoh-client

build: clean
	@echo "Building the binary for odoh-client ..."
	@echo "Tag: $(COMMIT_ID)"
	@echo "Version: $(VERSION)"
	@go build -ldflags "-X main.Version=$(VERSION) -X main.CommitId=$(COMMIT_ID)" ./cmd/*

fetch:
	@echo "Create the necessary directories ..."
	chmod +x ./fetch-datasets.sh
	./fetch-datasets.sh

install:
	@go install -ldflags "-X main.Version=$(VERSION) -X main.CommitId=$(COMMIT_ID)" ./cmd/*

.PHONY: all clean build fetch install