# Changelog

Tutte le modifiche rilevanti a questo progetto saranno documentate in questo file.

Il formato è basato su [Keep a Changelog](https://keepachangelog.com/it/1.1.0/),
e questo progetto aderisce al [Semantic Versioning](https://semver.org/lang/it/).

## [Unreleased]

### Aggiunto

- Icona applicazione (worker SVG/PNG)
- Target Makefile `install` per installare l'app in /Applications
- Target Makefile `uninstall` per rimuovere l'app
- Splash screen animata minimalista con effetto riempimento cerchio (2.5 secondi)
- Tema personalizzato con palette bianco e antracite

### Modificato

- Bottoni azione (Modifica, Sposta, Elimina) centrati orizzontalmente nel pannello dettagli
- Tree view migliorato: click su item per expand/collapse, conteggio elementi nei branch
- Icone contestuali: cartella piena/vuota per progetti, icona server per MCP server

## [0.0.1] - 2025-12-30

### Aggiunto

- Visualizzazione server MCP globali e per-progetto in tree view ordinata alfabeticamente
- Pannello dettagli con informazioni complete del server (tipo, comando, args, URL, env, timeout)
- Aggiunta nuovi server MCP (scope globale o progetto)
- Modifica server esistenti
- Eliminazione server con conferma
- Spostamento server tra scope (globale ↔ progetto)
- Backup automatico di ~/.claude.json prima di ogni modifica
- Preservazione campi JSON non gestiti durante lettura/scrittura
- Architettura Clean Architecture con separazione domain/application/infrastructure/ui
- Supporto per file di configurazione: ~/.claude.json, .mcp.json, .mcp.local.json
- Versione visualizzata nel titolo della finestra

### Tecnologie

- Go 1.24+
- Fyne v2.7.1 (UI toolkit nativo cross-platform)

[Unreleased]: https://github.com/strawberry-code/mcp-curator/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/strawberry-code/mcp-curator/releases/tag/v0.0.1
