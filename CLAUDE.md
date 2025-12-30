# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCP Manager è un tool desktop nativo in Go per visualizzare e gestire le configurazioni MCP (Model Context Protocol) di Claude Code. Supporta scope globale e per-progetto.

## Tech Stack

- **Linguaggio**: Go
- **UI Framework**: Fyne (fyne.io) - UI nativa, no Electron/webview
- **Architettura**: Clean Architecture

## Build Commands

```bash
# Build
go build -o bin/mcp-manager ./cmd/mcp-manager

# Run
go run ./cmd/mcp-manager

# Test
go test ./...

# Package per macOS
fyne package -os darwin -icon assets/icon.png

# Clean
rm -rf bin/
```

## Architecture

```
cmd/mcp-manager/main.go     # Entry point
internal/
├── domain/                  # Entità e regole business (MCPServer, Project, Configuration)
├── application/             # Use cases (list, add, remove, move server)
├── infrastructure/          # Implementazioni I/O (claude_config.go, project_config.go)
└── ui/                      # Componenti Fyne (app, main_window, server_list, forms)
```

### Domain Model

- **MCPServer**: server MCP con tipo (stdio|http|sse), command/args/url, env, timeout
- **Project**: progetto con path, server MCP, flag per .mcp.json/.mcp.local.json
- **Configuration**: aggregate root con server globali e progetti

### Config Files (gestiti dal tool)

1. `~/.claude.json` → `mcpServers` (globali)
2. `~/.claude.json` → `projects.[path].mcpServers` (per-progetto in settings)
3. `[project]/.mcp.json` (per-progetto)
4. `[project]/.mcp.local.json` (git-ignored)

## Key Implementation Notes

- **JSON Preservation**: usare `map[string]interface{}` per non perdere campi sconosciuti durante read/write
- **Backup**: creare `~/.claude.json.bak` prima di ogni modifica
- La risoluzione merge dei server: globali → project settings → .mcp.json → .mcp.local.json
