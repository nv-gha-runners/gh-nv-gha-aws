BIN_NAME = gh-nv-gha-aws
VERSION = $(shell git describe --tags --dirty --always)
BUILD_FLAGS = -tags osusergo,netgo \
              -ldflags "-s -extldflags=-static -X main.version=$(VERSION)"

build:
	go build -o $(BIN_NAME) $(BUILD_FLAGS)

check: gofmt-verify ci-lint

gofmt:
	@gofmt -w -l $$(find . -name '*.go')

gofmt-verify:
	@out=`gofmt -w -l -d $$(find . -name '*.go')`; \
		if [ -n "$$out" ]; then \
		echo "$$out"; \
		exit 1; \
		fi

ci-lint:
	@docker run --pull always --rm -v $(PWD):/app -w /app golangci/golangci-lint:latest golangci-lint run

clean:
	@rm -f $(BIN_NAME)
