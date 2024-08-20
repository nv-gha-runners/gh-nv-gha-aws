BIN_NAME = gh-nv-gha-aws
VERSION = $(shell git describe --tags --dirty --always)
BUILD_FLAGS = -tags osusergo,netgo \
		-ldflags "-s -extldflags=-static -X main.version=$(VERSION)"

build:
	go build -o $(BIN_NAME) $(BUILD_FLAGS)
