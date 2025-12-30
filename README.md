# MCP Curator

A native desktop application to view and manage MCP (Model Context Protocol) configurations for Claude Code.

## Features

- View global and per-project MCP servers in a tree view
- Add, edit, delete, and move servers between scopes
- Support for `~/.claude.json`, `.mcp.json`, and `.mcp.local.json`
- Automatic backup before modifications
- Native macOS app with anthracite theme

## Installation

### From Release (macOS Apple Silicon)

Download the latest `.dmg` from [Releases](https://github.com/strawberry-code/mcp-curator/releases) and drag to Applications.

**Note:** The app is not signed. After first launch, if macOS shows "app is damaged", run:
```bash
xattr -cr "/Applications/MCP Curator.app"
```

### From Source

**Requirements:** Go 1.24+

```bash
# Clone
git clone https://github.com/strawberry-code/mcp-curator.git
cd mcp-curator

# Build and install to /Applications
make install

# Or just run
make run
```

## Build Commands

```bash
make build      # Build binary
make run        # Run app
make build-mac  # Build .app bundle
make install    # Build and install to /Applications
make uninstall  # Remove from /Applications
make test       # Run tests
make clean      # Clean build artifacts
```

## License

MIT
