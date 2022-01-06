BINARY := compost
PKG := compost
VERSION := $(shell scripts/get-version.sh HEAD $(NO_DIRTY))
LD_FLAGS := -ldflags="-X 'github.com/infracost/compost-go/internal/version.Version=$(VERSION)'"
BUILD_FLAGS := $(LD_FLAGS) -v

.PHONY: deps run build windows linux darwin build_all install release clean test fmt lint

deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download

run:
	go run $(LD_FLAGS) $(PKG) $(ARGS)

build:
	CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY) $(PKG)

windows:
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY).exe $(PKG)

linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-linux-amd64 $(PKG)

darwin:
	env GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-darwin-amd64 $(PKG)
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o build/$(BINARY)-darwin-arm64 $(PKG)

build_all: build windows linux darwin

install:
	CGO_ENABLED=0 go install $(BUILD_FLAGS) $(PKG)

release: build_all
	cd build; tar -czf $(BINARY)-windows-amd64.tar.gz $(BINARY).exe
	cd build; tar -czf $(BINARY)-linux-amd64.tar.gz $(BINARY)-linux-amd64
	cd build; tar -czf $(BINARY)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64
	cd build; tar -czf $(BINARY)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64

clean:
	go clean
	rm -rf build/$(BINARY)*

test:
	go test $(LD_FLAGS) ./... $(or $(ARGS), -v -cover)

fmt:
	go fmt ./...

lint:
	golangci-lint run
