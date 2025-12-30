.PHONY: build run test clean build-mac install uninstall install-deps

# Aggiungi GOPATH/bin al PATH per tool come fyne
export PATH := $(PATH):$(shell go env GOPATH)/bin

# Variabili
BINARY_NAME=mcp-curator
BINARY_DIR=bin
CMD_DIR=./cmd/mcp-manager
APP_NAME=MCP Curator
APP_BUNDLE=$(APP_NAME).app
INSTALL_DIR=/Applications

# Build per la piattaforma corrente
build:
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(BINARY_NAME) $(CMD_DIR)

# Esegui l'applicazione
run:
	go run $(CMD_DIR)

# Build per macOS come .app bundle
build-mac:
	fyne package --target darwin --icon $(CURDIR)/assets/icon.png --name "$(APP_NAME)" --src $(CMD_DIR)

# Installa l'app in /Applications (richiede privilegi admin se necessario)
install: build-mac
	@echo "Installazione $(APP_BUNDLE) in $(INSTALL_DIR)..."
	@if [ -d "$(INSTALL_DIR)/$(APP_BUNDLE)" ]; then \
		echo "Rimozione versione precedente..."; \
		rm -rf "$(INSTALL_DIR)/$(APP_BUNDLE)"; \
	fi
	@cp -R "$(APP_BUNDLE)" "$(INSTALL_DIR)/"
	@echo "Installazione completata: $(INSTALL_DIR)/$(APP_BUNDLE)"

# Disinstalla l'app da /Applications
uninstall:
	@if [ -d "$(INSTALL_DIR)/$(APP_BUNDLE)" ]; then \
		echo "Rimozione $(APP_BUNDLE) da $(INSTALL_DIR)..."; \
		rm -rf "$(INSTALL_DIR)/$(APP_BUNDLE)"; \
		echo "Disinstallazione completata."; \
	else \
		echo "$(APP_BUNDLE) non trovata in $(INSTALL_DIR)."; \
	fi

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
