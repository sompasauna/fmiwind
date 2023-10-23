GO_SRCS := $(shell find . -type f -name '*.go')
TAG_NAME = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
ifeq ($(TAG_NAME),)
TAG_NAME = dev
endif
BUILD_FLAGS = -trimpath -a -tags "netgo,osusergo,static_build" -installsuffix netgo -ldflags "-s -w -X github.com/sompasauna/fmiwind/version.Version=$(TAG_NAME) -extldflags '-static'"
PREFIX = /usr/local

fmiwind: $(GO_SRCS)
	go build $(BUILD_FLAGS) -o fmiwind fmiwind.go

.PHONY: install
install: fmiwind
	install -d $(DESTDIR)$(PREFIX)/bin/
	install -m 755 fmiwind $(DESTDIR)$(PREFIX)/bin/

bin/fmiwind-linux-amd64: $(GO_SRCS)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/fmiwind-linux-amd64 fmiwind.go

bin/fmiwind-linux-arm64: $(GO_SRCS)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/fmiwind-linux-arm64 fmiwind.go

bin/fmiwind-linux-arm: $(GO_SRCS)
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build $(BUILD_FLAGS) -o bin/fmiwind-linux-arm fmiwind.go

bin/fmiwind-darwin-amd64: $(GO_SRCS)
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o bin/k0sctl-darwin-amd64 fmiwind.go

bin/fmiwind-darwin-arm64: $(GO_SRCS)
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o bin/k0sctl-darwin-arm64 fmiwind.go

bins := bin/fmiwind-linux-amd64 bin/fmiwind-linux-arm64 bin/fmiwind-linux-arm bin/fmiwind-darwin-amd64 bin/fmiwind-darwin-arm64

all: $(bins)

clean:
	rm -f fmiwind
	rm -rf bin/
