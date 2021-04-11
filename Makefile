.PHONY: build

ICON="🔞"

# Project binaries
COMMANDS=tdraw

BINARIES=$(addprefix bin/,$(COMMANDS))

all: binaries

FORCE:
define BUILD_BINARY
@echo "$(ICON) $@"
@go build -o $@ ./$<
endef

build: cmd/tdraw
	@echo "$(ICON) $@"
	@go build -o bin/tdraw ./$<