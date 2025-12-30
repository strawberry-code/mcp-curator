# MCP Manager - Specifiche e Requisiti

## Overview

Tool desktop nativo per visualizzare e gestire le configurazioni MCP (Model Context Protocol) di Claude Code, con supporto per scope globale e per-progetto.

---

## Problema

Claude Code memorizza le configurazioni MCP in modo frammentato:

1. **Globale**: `~/.claude.json` → `mcpServers`
2. **Per-progetto (in settings)**: `~/.claude.json` → `projects.[path].mcpServers`
3. **Per-progetto (file)**: `[project]/.mcp.json`
4. **Per-progetto locale**: `[project]/.mcp.local.json` (git-ignored)

Non esiste un modo semplice per:
- Vedere tutti i server configurati e dove sono definiti
- Capire quale configurazione si applica a quale progetto
- Spostare server tra scope diversi
- Evitare duplicazioni

---

## Obiettivi

### Must Have (v1.0)

1. **Visualizzazione chiara** di tutte le configurazioni MCP:
   - Lista server globali
   - Lista progetti con i loro server specifici
   - Merge view: cosa vede effettivamente Claude in un dato progetto

2. **Gestione base**:
   - Aggiungere server (globale o per-progetto)
   - Rimuovere server
   - Modificare configurazione server esistente
   - Spostare server tra scope (globale ↔ progetto)

3. **App desktop nativa**:
   - Avvio veloce (< 1s)
   - UI responsive
   - Cross-platform (macOS prioritario, Linux secondario, Windows opzionale)

### Nice to Have (v1.1+)

- Copiare configurazione tra progetti
- Template di server predefiniti
- Validazione configurazioni
- Test connessione server
- Import/export configurazioni
- Backup automatico prima di modifiche

---

## Architettura

### Stack Tecnologico

**Linguaggio**: Go
- Single binary, zero dipendenze runtime
- Ottime performance
- Cross-compilation nativa

**UI Framework**: [Fyne](https://fyne.io/)
- UI nativa Go
- Cross-platform (macOS, Linux, Windows)
- Leggero e veloce
- No Electron, no webview

### Struttura Progetto (Clean Architecture)

```
mcp-manager/
├── cmd/
│   └── mcp-manager/
│       └── main.go              # Entry point
├── internal/
│   ├── domain/                  # Entità e regole business
│   │   ├── server.go            # MCPServer entity
│   │   ├── project.go           # Project entity
│   │   └── config.go            # Configuration aggregate
│   ├── application/             # Use cases
│   │   ├── list_servers.go
│   │   ├── add_server.go
│   │   ├── remove_server.go
│   │   ├── move_server.go
│   │   └── get_merged_config.go
│   ├── infrastructure/          # Implementazioni concrete
│   │   ├── claude_config.go     # Lettura/scrittura ~/.claude.json
│   │   ├── project_config.go    # Lettura/scrittura .mcp.json
│   │   └── file_watcher.go      # Watch per reload automatico
│   └── ui/                      # Interfaccia utente Fyne
│       ├── app.go               # Setup applicazione
│       ├── main_window.go       # Finestra principale
│       ├── server_list.go       # Lista server component
│       ├── server_form.go       # Form aggiunta/modifica
│       └── project_tree.go      # Tree view progetti
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Domain Model

### MCPServer

```go
type MCPServer struct {
    Name    string
    Type    ServerType        // stdio | http | sse
    Command string            // per stdio
    Args    []string          // per stdio
    URL     string            // per http/sse
    Headers map[string]string // per http/sse
    Env     map[string]string
    Timeout int               // ms, opzionale
}

type ServerType string

const (
    ServerTypeStdio ServerType = "stdio"
    ServerTypeHTTP  ServerType = "http"
    ServerTypeSSE   ServerType = "sse"
)
```

### Project

```go
type Project struct {
    Path           string
    Name           string                  // basename del path
    MCPServers     map[string]MCPServer    // da ~/.claude.json projects
    HasMCPJson     bool                    // esiste .mcp.json?
    HasMCPLocal    bool                    // esiste .mcp.local.json?
}
```

### Configuration (Aggregate Root)

```go
type Configuration struct {
    GlobalServers   map[string]MCPServer
    Projects        []Project
    ClaudeJsonPath  string
}

// Restituisce i server effettivi per un progetto (merge)
func (c *Configuration) GetEffectiveServers(projectPath string) map[string]MCPServer
```

---

## Use Cases

### 1. ListServers

**Input**: nessuno
**Output**: Configuration completa
**Logica**:
1. Leggi `~/.claude.json`
2. Estrai `mcpServers` (globali)
3. Estrai `projects` con relativi `mcpServers`
4. Per ogni progetto, verifica esistenza `.mcp.json` e `.mcp.local.json`
5. Costruisci e ritorna Configuration

### 2. AddServer

**Input**: MCPServer, Scope (global | project path)
**Output**: success/error
**Logica**:
1. Valida MCPServer
2. Se scope = global: aggiungi a `mcpServers`
3. Se scope = project: aggiungi a `projects.[path].mcpServers`
4. Salva `~/.claude.json`

### 3. RemoveServer

**Input**: serverName, Scope
**Output**: success/error
**Logica**:
1. Rimuovi da scope appropriato
2. Salva `~/.claude.json`

### 4. MoveServer

**Input**: serverName, fromScope, toScope
**Output**: success/error
**Logica**:
1. Leggi server da fromScope
2. Aggiungi a toScope
3. Rimuovi da fromScope
4. Salva `~/.claude.json`

### 5. GetMergedConfig

**Input**: projectPath
**Output**: map[string]MCPServer effettivi
**Logica**:
1. Parti da server globali
2. Sovrascrivi con server da `projects.[path].mcpServers`
3. Sovrascrivi con server da `.mcp.json` (se esiste)
4. Sovrascrivi con server da `.mcp.local.json` (se esiste)

---

## UI Design

### Layout Principale

```
┌─────────────────────────────────────────────────────────────┐
│  MCP Manager                                          [−][×]│
├─────────────────────────────────────────────────────────────┤
│ ┌─────────────────┐ ┌─────────────────────────────────────┐ │
│ │ 📁 Scope        │ │ Server Details                      │ │
│ │                 │ │                                     │ │
│ │ ▼ 🌐 Global     │ │ Name: serena                        │ │
│ │   ├─ memory     │ │ Type: stdio                         │ │
│ │   ├─ serena     │ │ Command: uvx                        │ │
│ │   ├─ playwright │ │ Args: --from git+https://...        │ │
│ │   └─ shadcn     │ │                                     │ │
│ │                 │ │ Env:                                │ │
│ │ ▼ 📂 Projects   │ │   (none)                            │ │
│ │   ▼ easy-cqs    │ │                                     │ │
│ │     ├─ ragify   │ │ ┌─────────┐ ┌─────────┐ ┌────────┐  │ │
│ │     └─ atlassian│ │ │  Edit   │ │  Move   │ │ Delete │  │ │
│ │   ▶ alpharag    │ │ └─────────┘ └─────────┘ └────────┘  │ │
│ │   ▶ neuroswap   │ │                                     │ │
│ │                 │ │                                     │ │
│ └─────────────────┘ └─────────────────────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│ [+ Add Server]                              [↻ Refresh]     │
└─────────────────────────────────────────────────────────────┘
```

### Componenti

1. **Sidebar (Scope Tree)**
   - Nodo "Global" espandibile con lista server
   - Nodo "Projects" con sotto-nodi per ogni progetto
   - Icone per distinguere sorgente (.mcp.json vs settings)
   - Click seleziona, doppio-click espande/collassa

2. **Detail Panel**
   - Mostra dettagli server selezionato
   - Campi read-only con possibilità di edit
   - Bottoni azione: Edit, Move, Delete

3. **Toolbar**
   - Add Server (apre dialog)
   - Refresh (ricarica configurazione)

### Dialog: Add/Edit Server

```
┌─────────────────────────────────────┐
│ Add MCP Server                      │
├─────────────────────────────────────┤
│ Name:     [___________________]     │
│                                     │
│ Scope:    (•) Global                │
│           ( ) Project: [dropdown▼]  │
│                                     │
│ Type:     [stdio        ▼]          │
│                                     │
│ ── STDIO Config ──                  │
│ Command:  [___________________]     │
│ Args:     [___________________]     │
│                                     │
│ ── Environment ──                   │
│ [KEY        ] [VALUE          ] [+] │
│ [KEY        ] [VALUE          ] [−] │
│                                     │
│        [Cancel]  [Save]             │
└─────────────────────────────────────┘
```

### Dialog: Move Server

```
┌─────────────────────────────────────┐
│ Move Server: serena                 │
├─────────────────────────────────────┤
│                                     │
│ From: Global                        │
│                                     │
│ To:   ( ) Global                    │
│       (•) Project: [easy-cqs    ▼]  │
│                                     │
│        [Cancel]  [Move]             │
└─────────────────────────────────────┘
```

---

## File Handling

### Lettura ~/.claude.json

```go
type ClaudeConfig struct {
    MCPServers map[string]json.RawMessage `json:"mcpServers"`
    Projects   map[string]ProjectConfig   `json:"projects"`
    // ... altri campi ignorati ma preservati
}
```

**Importante**: Preservare tutti i campi non gestiti durante la scrittura (usare `json.RawMessage` o map generico per campi sconosciuti).

### Backup

Prima di ogni modifica:
1. Copia `~/.claude.json` → `~/.claude.json.bak`
2. Mantieni ultimi 5 backup con timestamp

---

## Error Handling

- File non trovato → UI mostra stato vuoto con messaggio
- JSON malformato → Errore con path al problema, no crash
- Permessi insufficienti → Messaggio chiaro, suggerimento fix
- Server duplicato → Warning, chiedi conferma sovrascrittura

---

## Testing Strategy

### Unit Tests
- Domain entities
- Use cases (con mock del repository)
- JSON parsing/serialization

### Integration Tests
- Lettura/scrittura file reali (in temp dir)
- Scenari completi (add → move → delete)

### Manual Testing
- macOS native look & feel
- Resize window
- Keyboard navigation

---

## Build & Distribution

### Makefile

```makefile
.PHONY: build run test clean

build:
	go build -o bin/mcp-manager ./cmd/mcp-manager

build-mac:
	fyne package -os darwin -icon assets/icon.png

run:
	go run ./cmd/mcp-manager

test:
	go test ./...

clean:
	rm -rf bin/
```

### Release

- macOS: `.app` bundle via `fyne package`
- Linux: Binary + `.desktop` file
- Windows: `.exe` (opzionale)

---

## Milestones

### v0.1 - MVP Read-Only
- [ ] Setup progetto Go + Fyne
- [ ] Lettura ~/.claude.json
- [ ] UI con tree view e detail panel
- [ ] Visualizzazione server globali e per-progetto

### v0.2 - CRUD Base
- [ ] Add server (global e project)
- [ ] Remove server
- [ ] Edit server
- [ ] Backup automatico

### v0.3 - Move & Polish
- [ ] Move server tra scope
- [ ] Lettura .mcp.json progetti
- [ ] Merged view per progetto
- [ ] Keyboard shortcuts

### v1.0 - Release
- [ ] Packaging macOS
- [ ] README e documentazione
- [ ] Error handling completo
- [ ] Test coverage > 70%

---

## Note Tecniche

### Dipendenze Go

```go
require (
    fyne.io/fyne/v2 v2.4.0
)
```

### Path Resolution

```go
func getClaudeConfigPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".claude.json")
}
```

### JSON Preservation

Per non perdere campi sconosciuti durante read/write:

```go
// Leggi come map generico
var raw map[string]interface{}
json.Unmarshal(data, &raw)

// Modifica solo i campi necessari
raw["mcpServers"] = newServers

// Riscrivi tutto
json.MarshalIndent(raw, "", "  ")
```

---

## Riferimenti

- [Claude Code MCP Docs](https://docs.anthropic.com/en/docs/claude-code/mcp)
- [Fyne Documentation](https://developer.fyne.io/)
- [MCP Specification](https://modelcontextprotocol.io/)
