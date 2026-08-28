PREFIX  ?= /usr/local
VERSION ?= 1.1.1
# Dots look like a file extension to some tools; use dashes in the artifact name.
VERSION_TAG := $(subst .,-,$(VERSION))
ROOT    := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
DIST    := $(ROOT)/dist
BIN     := $(DIST)/playlist-md-v$(VERSION_TAG)
ASSET   := $(ROOT)/go/assets/playlist-md-core
CORE_ARM := $(ROOT)/.build/arm64-apple-macosx/release/playlist-md-core
CORE_X86 := $(ROOT)/.build/x86_64-apple-macosx/release/playlist-md-core

.PHONY: all core launcher test clean install

all: launcher

core:
	swift build -c release --arch arm64 --product playlist-md-core
	swift build -c release --arch x86_64 --product playlist-md-core
	mkdir -p $(dir $(ASSET)) $(DIST)
	lipo -create $(CORE_ARM) $(CORE_X86) -output $(ASSET)
	chmod +x $(ASSET)

launcher: core
	cd $(ROOT)/go && go mod tidy
	cd $(ROOT)/go && GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o $(DIST)/playlist-md-arm64 ./cmd/playlistmd
	cd $(ROOT)/go && GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(DIST)/playlist-md-amd64 ./cmd/playlistmd
	lipo -create $(DIST)/playlist-md-arm64 $(DIST)/playlist-md-amd64 -output $(BIN)
	rm -f $(DIST)/playlist-md-arm64 $(DIST)/playlist-md-amd64
	chmod +x $(BIN)

release-tar: launcher
	rm -rf $(RELEASE_DIR)
	mkdir -p $(RELEASE_DIR)
	install -m 755 $(BIN) $(RELEASE_DIR)/playlist-md
	install -m 644 README.md LICENSE $(RELEASE_DIR)/
	tar -czf $(RELEASE_TAR) -C $(DIST) $(RELEASE_NAME)
	rm -rf $(RELEASE_DIR)

test:
	swift test
	@mkdir -p $(dir $(ASSET))
	@test -f $(ASSET) || touch $(ASSET)
	cd $(ROOT)/go && go test ./...

install: launcher
	install -d $(PREFIX)/bin
	install -m 755 $(BIN) $(PREFIX)/bin/playlist-md

clean:
	rm -rf $(ROOT)/.build $(DIST) $(ASSET)
	cd $(ROOT)/go && go clean
