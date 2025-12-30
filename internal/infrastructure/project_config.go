package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/strawberry-code/mcp-curator/internal/domain"
)

// ProjectConfigRepository gestisce la lettura di .mcp.json e .mcp.local.json
type ProjectConfigRepository struct{}

// NewProjectConfigRepository crea un nuovo repository per i file di progetto
func NewProjectConfigRepository() *ProjectConfigRepository {
	return &ProjectConfigRepository{}
}

// LoadProjectMCP carica i server MCP da .mcp.json di un progetto
func (r *ProjectConfigRepository) LoadProjectMCP(projectPath string) (map[string]domain.MCPServer, error) {
	return r.loadMCPFile(filepath.Join(projectPath, ".mcp.json"))
}

// LoadProjectMCPLocal carica i server MCP da .mcp.local.json di un progetto
func (r *ProjectConfigRepository) LoadProjectMCPLocal(projectPath string) (map[string]domain.MCPServer, error) {
	return r.loadMCPFile(filepath.Join(projectPath, ".mcp.local.json"))
}

// loadMCPFile carica un file .mcp.json o .mcp.local.json
func (r *ProjectConfigRepository) loadMCPFile(path string) (map[string]domain.MCPServer, error) {
	result := make(map[string]domain.MCPServer)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("impossibile leggere %s: %w", path, err)
	}

	var raw struct {
		MCPServers map[string]interface{} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("JSON non valido in %s: %w", path, err)
	}

	for name, serverData := range raw.MCPServers {
		server, err := parseServer(serverData)
		if err != nil {
			continue
		}
		server.Name = name
		result[name] = server
	}

	return result, nil
}

// HasMCPJson verifica se un progetto ha un file .mcp.json
func (r *ProjectConfigRepository) HasMCPJson(projectPath string) bool {
	return fileExists(filepath.Join(projectPath, ".mcp.json"))
}

// HasMCPLocal verifica se un progetto ha un file .mcp.local.json
func (r *ProjectConfigRepository) HasMCPLocal(projectPath string) bool {
	return fileExists(filepath.Join(projectPath, ".mcp.local.json"))
}

// GetEffectiveServers restituisce i server effettivi per un progetto con merge completo
// Ordine: globali < project settings (da ~/.claude.json) < .mcp.json < .mcp.local.json
func (r *ProjectConfigRepository) GetEffectiveServers(
	config *domain.Configuration,
	projectPath string,
) (map[string]domain.MCPServer, error) {
	// Parti dal merge base (globali + project settings)
	result := config.GetEffectiveServers(projectPath)

	// Aggiungi server da .mcp.json
	mcpServers, err := r.LoadProjectMCP(projectPath)
	if err != nil {
		return nil, err
	}
	for name, server := range mcpServers {
		result[name] = server
	}

	// Aggiungi server da .mcp.local.json
	localServers, err := r.LoadProjectMCPLocal(projectPath)
	if err != nil {
		return nil, err
	}
	for name, server := range localServers {
		result[name] = server
	}

	return result, nil
}
