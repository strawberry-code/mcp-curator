.PHONY: build run test clean build-mac install-deps

# Variabili
BINARY_NAME=mcp-curator
BINARY_DIR=bin
CMD_DIR=./cmd/mcp-manager

# Build per la piattaforma corrente
build:
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Esegui l'applicazione
run:
	go run $(CMD_DIR)

# Build per macOS come .app bundle
build-mac:
	fyne package -os darwin -icon assets/icon.png -name "MCP Curator"

# Test
test:
	go test -v ./...

# Test con coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Pulisci build artifacts
clean:
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html
	rm -rf "MCP Curator.app"

# Installa dipendenze
install-deps:
	go mod download
	go mod tidy

# Formatta codice
fmt:
	go fmt ./...

# Lint
lint:
	golangci-lint run

# Verifica compilazione
check:
	go build -o /dev/null $(CMD_DIR)
