
# The binaries to build (just the basenames).
BIN := bitcoinAddressGeneratorServer

# This version-strategy uses git tags to set the version string
#VERSION ?= $(shell git describe --tags --always --dirty)
#
# This version-strategy uses a manual value to set the version string
VERSION ?= 1.0.0

# Used internally.  Users should pass GOOS and/or GOARCH.
OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
TAG := $(VERSION)_$(OS)_$(ARCH)

SRC_DIRS := cmd
PKG_DIRS := cipher
OUTPUT_DIR := bin

build: # @HELP build binary
	go build -o $(OUTPUT_DIR)/$(BIN)-$(TAG) $(SRC_DIRS)/server.go $(SRC_DIRS)/struct.go

clean: # @HELP removes built binaries and temporary files
	rm -r $(OUTPUT_DIR)

tests: # @HELP run tests
	go test ./$(PKG_DIRS)/... -v
	go test ./$(SRC_DIRS)/... -v


help: # @HELP prints this message
help:
	@echo "VARIABLES:"
	@echo "  BIN = $(BIN)"
	@echo "  OS = $(OS)"
	@echo "  ARCH = $(ARCH)"
	@echo "  VERSION = $(VERSION)"
	@echo
	@echo "TARGETS:"
	@grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)    \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '
