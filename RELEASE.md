# Release Procedure

## Bump version

1. Aggiorna versione in `internal/version/version.go`
2. Aggiorna `CHANGELOG.md`: sposta [Unreleased] in [X.Y.Z] con data
3. Aggiorna link in fondo al CHANGELOG

## Commit e tag

4. `git add -A && git commit -m "release: vX.Y.Z"`
5. `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
6. `git push && git push --tags`

## Build e release

7. `make build-mac`
8. `hdiutil create -volname "MCP Curator" -srcfolder "MCP Curator.app" -ov -format UDZO "MCP-Curator-vX.Y.Z-apple-silicon.dmg"`
9. `gh release create vX.Y.Z --title "MCP Curator vX.Y.Z" --notes "See CHANGELOG.md" "MCP-Curator-vX.Y.Z-apple-silicon.dmg"`

## Homebrew tap

10. `shasum -a 256 MCP-Curator-vX.Y.Z-apple-silicon.dmg`
11. Aggiorna `version` e `sha256` in `/tmp/homebrew-mcp-curator/Casks/mcp-curator.rb`
12. `cd /tmp/homebrew-mcp-curator && git add -A && git commit -m "bump: vX.Y.Z" && git push`
