.PHONY: build test clean run install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)"
GOFLAGS := -trimpath

build:
	go build $(GOFLAGS) $(LDFLAGS) -o cmdperf

test:
	go test -v ./...

clean:
	rm -f cmdperf

run:
	go run $(GOFLAGS) $(LDFLAGS) .

install:
	go install $(GOFLAGS) $(LDFLAGS)

demo: build
	chmod +x examples/demo.sh
	./examples/demo.sh

.PHONY: release-binaries
release-binaries:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o dist/cmdperf-linux-amd64
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o dist/cmdperf-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) $(LDFLAGS) -o dist/cmdperf-darwin-arm64
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) $(LDFLAGS) -o dist/cmdperf-windows-amd64.exe
