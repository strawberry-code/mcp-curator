package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/strawberry-code/mcp-curator/internal/domain"
)

// ClaudeConfigRepository gestisce la lettura/scrittura di ~/.claude.json
type ClaudeConfigRepository struct {
	configPath string
	rawConfig  map[string]interface{}
}

// NewClaudeConfigRepository crea un nuovo repository per ~/.claude.json
func NewClaudeConfigRepository() (*ClaudeConfigRepository, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("impossibile determinare home directory: %w", err)
	}

	return &ClaudeConfigRepository{
		configPath: filepath.Join(home, ".claude.json"),
	}, nil
}

// NewClaudeConfigRepositoryWithPath crea un repository con path personalizzato (per test)
func NewClaudeConfigRepositoryWithPath(path string) *ClaudeConfigRepository {
	return &ClaudeConfigRepository{
		configPath: path,
	}
}

// GetConfigPath restituisce il path del file di configurazione
func (r *ClaudeConfigRepository) GetConfigPath() string {
	return r.configPath
}

// Load carica la configurazione da disco
func (r *ClaudeConfigRepository) Load() (*domain.Configuration, error) {
	config := domain.NewConfiguration(r.configPath)

	data, err := os.ReadFile(r.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("impossibile leggere %s: %w", r.configPath, err)
	}

	r.rawConfig = make(map[string]interface{})
	if err := json.Unmarshal(data, &r.rawConfig); err != nil {
		return nil, fmt.Errorf("JSON non valido in %s: %w", r.configPath, err)
	}

	// Estrai mcpServers globali
	if mcpServers, ok := r.rawConfig["mcpServers"].(map[string]interface{}); ok {
		for name, serverData := range mcpServers {
			server, err := parseServer(serverData)
			if err != nil {
				continue
			}
			config.AddGlobalServer(name, server)
		}
	}

	// Estrai projects con i loro mcpServers
	if projects, ok := r.rawConfig["projects"].(map[string]interface{}); ok {
		for path, projectData := range projects {
			projectMap, ok := projectData.(map[string]interface{})
			if !ok {
				continue
			}

			project := domain.NewProject(path)

			if mcpServers, ok := projectMap["mcpServers"].(map[string]interface{}); ok {
				for name, serverData := range mcpServers {
					server, err := parseServer(serverData)
					if err != nil {
						continue
					}
					project.AddServer(name, server)
				}
			}

			// Verifica esistenza file .mcp.json e .mcp.local.json
			project.HasMCPJson = fileExists(filepath.Join(path, ".mcp.json"))
			project.HasMCPLocal = fileExists(filepath.Join(path, ".mcp.local.json"))

			config.AddProject(project)
		}
	}

	return config, nil
}

// Save salva la configurazione su disco
func (r *ClaudeConfigRepository) Save(config *domain.Configuration) error {
	if err := r.backup(); err != nil {
		return fmt.Errorf("impossibile creare backup: %w", err)
	}

	// Se non abbiamo un rawConfig, creane uno nuovo
	if r.rawConfig == nil {
		r.rawConfig = make(map[string]interface{})
	}

	// Aggiorna mcpServers globali
	mcpServers := make(map[string]interface{})
	for name, server := range config.GlobalServers {
		mcpServers[name] = serverToMap(server)
	}
	r.rawConfig["mcpServers"] = mcpServers

	// Aggiorna projects
	projects := make(map[string]interface{})
	if existingProjects, ok := r.rawConfig["projects"].(map[string]interface{}); ok {
		// Mantieni i dati esistenti dei progetti
		for path, data := range existingProjects {
			projects[path] = data
		}
	}

	for path, project := range config.Projects {
		projectData := make(map[string]interface{})

		// Mantieni dati esistenti del progetto
		if existing, ok := projects[path].(map[string]interface{}); ok {
			for k, v := range existing {
				projectData[k] = v
			}
		}

		// Aggiorna mcpServers del progetto
		projectServers := make(map[string]interface{})
		for name, server := range project.MCPServers {
			projectServers[name] = serverToMap(server)
		}
		projectData["mcpServers"] = projectServers

		projects[path] = projectData
	}
	r.rawConfig["projects"] = projects

	// Serializza e scrivi
	data, err := json.MarshalIndent(r.rawConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("impossibile serializzare configurazione: %w", err)
	}

	if err := os.WriteFile(r.configPath, data, 0644); err != nil {
		return fmt.Errorf("impossibile scrivere %s: %w", r.configPath, err)
	}

	return nil
}

// backup crea un backup del file di configurazione
func (r *ClaudeConfigRepository) backup() error {
	if !fileExists(r.configPath) {
		return nil
	}

	backupPath := r.configPath + ".bak"
	timestampedPath := fmt.Sprintf("%s.%s.bak", r.configPath, time.Now().Format("20060102-150405"))

	// Copia nel backup principale
	data, err := os.ReadFile(r.configPath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return err
	}

	// Copia nel backup con timestamp
	if err := os.WriteFile(timestampedPath, data, 0644); err != nil {
		return err
	}

	// Pulisci vecchi backup (mantieni ultimi 5)
	r.cleanOldBackups()

	return nil
}

// cleanOldBackups rimuove i backup più vecchi mantenendo gli ultimi 5
func (r *ClaudeConfigRepository) cleanOldBackups() {
	dir := filepath.Dir(r.configPath)
	base := filepath.Base(r.configPath)
	pattern := base + ".*.bak"

	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil || len(matches) <= 5 {
		return
	}

	// I file sono già ordinati per nome (quindi per timestamp)
	// Rimuovi i più vecchi
	for i := 0; i < len(matches)-5; i++ {
		os.Remove(matches[i])
	}
}

// parseServer converte un map[string]interface{} in MCPServer
func parseServer(data interface{}) (domain.MCPServer, error) {
	serverMap, ok := data.(map[string]interface{})
	if !ok {
		return domain.MCPServer{}, fmt.Errorf("formato server non valido")
	}

	server := domain.MCPServer{}

	if t, ok := serverMap["type"].(string); ok {
		server.Type = domain.ServerType(t)
	}
	if cmd, ok := serverMap["command"].(string); ok {
		server.Command = cmd
	}
	if url, ok := serverMap["url"].(string); ok {
		server.URL = url
	}
	if timeout, ok := serverMap["timeout"].(float64); ok {
		server.Timeout = int(timeout)
	}

	if args, ok := serverMap["args"].([]interface{}); ok {
		server.Args = make([]string, 0, len(args))
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				server.Args = append(server.Args, s)
			}
		}
	}

	if headers, ok := serverMap["headers"].(map[string]interface{}); ok {
		server.Headers = make(map[string]string)
		for k, v := range headers {
			if s, ok := v.(string); ok {
				server.Headers[k] = s
			}
		}
	}

	if env, ok := serverMap["env"].(map[string]interface{}); ok {
		server.Env = make(map[string]string)
		for k, v := range env {
			if s, ok := v.(string); ok {
				server.Env[k] = s
			}
		}
	}

	return server, nil
}

// serverToMap converte MCPServer in map[string]interface{}
func serverToMap(server domain.MCPServer) map[string]interface{} {
	result := make(map[string]interface{})

	if server.Type != "" {
		result["type"] = string(server.Type)
	}
	if server.Command != "" {
		result["command"] = server.Command
	}
	if server.URL != "" {
		result["url"] = server.URL
	}
	if server.Timeout > 0 {
		result["timeout"] = server.Timeout
	}
	if len(server.Args) > 0 {
		result["args"] = server.Args
	}
	if len(server.Headers) > 0 {
		result["headers"] = server.Headers
	}
	if len(server.Env) > 0 {
		result["env"] = server.Env
	}

	return result
}

// fileExists verifica se un file esiste
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
