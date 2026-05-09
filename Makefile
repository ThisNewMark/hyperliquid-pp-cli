.PHONY: build test lint install clean release release-mcp release-mcpb release-all

# Builds the CLI as `hyperliquid` (short name shipped to users). The Go
# module name is still hyperliquid-pp-cli internally — only the binary
# basename is rebranded.
build:
	go build -o bin/hyperliquid ./cmd/hyperliquid-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

# `make install` puts $GOPATH/bin/hyperliquid on PATH instead of the
# default $GOPATH/bin/hyperliquid-pp-cli. Kept the underscored package name
# so any references in generated source compile unchanged.
install:
	go build -o $$(go env GOPATH)/bin/hyperliquid ./cmd/hyperliquid-pp-cli

clean:
	rm -rf bin/ dist/

build-mcp:
	go build -o bin/hyperliquid-mcp ./cmd/hyperliquid-pp-mcp

install-mcp:
	go build -o $$(go env GOPATH)/bin/hyperliquid-mcp ./cmd/hyperliquid-pp-mcp

build-all: build build-mcp

# release: cross-compile the CLI for darwin-arm64, darwin-amd64, linux-amd64,
# linux-arm64, windows-amd64. Outputs to dist/ as named binaries (and .exe).
release:
	mkdir -p dist
	GOOS=darwin  GOARCH=arm64 go build -o dist/hyperliquid-darwin-arm64       ./cmd/hyperliquid-pp-cli
	GOOS=darwin  GOARCH=amd64 go build -o dist/hyperliquid-darwin-amd64       ./cmd/hyperliquid-pp-cli
	GOOS=linux   GOARCH=amd64 go build -o dist/hyperliquid-linux-amd64        ./cmd/hyperliquid-pp-cli
	GOOS=linux   GOARCH=arm64 go build -o dist/hyperliquid-linux-arm64        ./cmd/hyperliquid-pp-cli
	GOOS=windows GOARCH=amd64 go build -o dist/hyperliquid-windows-amd64.exe  ./cmd/hyperliquid-pp-cli

# release-mcp: same matrix for the MCP server (raw binaries).
release-mcp:
	mkdir -p dist
	GOOS=darwin  GOARCH=arm64 go build -o dist/hyperliquid-mcp-darwin-arm64       ./cmd/hyperliquid-pp-mcp
	GOOS=darwin  GOARCH=amd64 go build -o dist/hyperliquid-mcp-darwin-amd64       ./cmd/hyperliquid-pp-mcp
	GOOS=linux   GOARCH=amd64 go build -o dist/hyperliquid-mcp-linux-amd64        ./cmd/hyperliquid-pp-mcp
	GOOS=linux   GOARCH=arm64 go build -o dist/hyperliquid-mcp-linux-arm64        ./cmd/hyperliquid-pp-mcp
	GOOS=windows GOARCH=amd64 go build -o dist/hyperliquid-mcp-windows-amd64.exe  ./cmd/hyperliquid-pp-mcp

# release-mcpb: build .mcpb bundles (Claude Desktop drag-drop installers) for
# every platform via the printing-press bundle command. Each bundle is a ZIP
# containing the MCP binary and the companion CLI binary.
release-mcpb:
	mkdir -p dist
	printing-press bundle . --platform darwin/arm64  --output dist/hyperliquid-darwin-arm64.mcpb
	printing-press bundle . --platform darwin/amd64  --output dist/hyperliquid-darwin-amd64.mcpb
	printing-press bundle . --platform linux/amd64   --output dist/hyperliquid-linux-amd64.mcpb
	printing-press bundle . --platform linux/arm64   --output dist/hyperliquid-linux-arm64.mcpb
	printing-press bundle . --platform windows/amd64 --output dist/hyperliquid-windows-amd64.mcpb

# release-all: build everything ready for `gh release create`.
release-all: release release-mcp release-mcpb
	@echo
	@echo "Built artifacts in dist/:"
	@ls -lh dist/
