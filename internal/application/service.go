package application

import (
	"fmt"

	"github.com/strawberry-code/mcp-curator/internal/domain"
	"github.com/strawberry-code/mcp-curator/internal/infrastructure"
)

// MCPService gestisce i casi d'uso per la configurazione MCP
type MCPService struct {
	claudeRepo  *infrastructure.ClaudeConfigRepository
	projectRepo *infrastructure.ProjectConfigRepository
	config      *domain.Configuration
}

// NewMCPService crea un nuovo servizio MCP
func NewMCPService() (*MCPService, error) {
	claudeRepo, err := infrastructure.NewClaudeConfigRepository()
	if err != nil {
		return nil, err
	}

	return &MCPService{
		claudeRepo:  claudeRepo,
		projectRepo: infrastructure.NewProjectConfigRepository(),
	}, nil
}

// Load carica la configurazione
func (s *MCPService) Load() error {
	config, err := s.claudeRepo.Load()
	if err != nil {
		return err
	}
	s.config = config
	return nil
}

// GetConfiguration restituisce la configurazione corrente
func (s *MCPService) GetConfiguration() *domain.Configuration {
	return s.config
}

// GetConfigPath restituisce il path del file di configurazione
func (s *MCPService) GetConfigPath() string {
	return s.claudeRepo.GetConfigPath()
}

// AddGlobalServer aggiunge un server globale
func (s *MCPService) AddGlobalServer(name string, server domain.MCPServer) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	if _, exists := s.config.GlobalServers[name]; exists {
		return fmt.Errorf("server '%s' già esistente", name)
	}

	s.config.AddGlobalServer(name, server)
	return s.claudeRepo.Save(s.config)
}

// AddProjectServer aggiunge un server a un progetto
func (s *MCPService) AddProjectServer(projectPath, name string, server domain.MCPServer) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	project := s.config.GetOrCreateProject(projectPath)

	if _, exists := project.MCPServers[name]; exists {
		return fmt.Errorf("server '%s' già esistente nel progetto", name)
	}

	project.AddServer(name, server)
	return s.claudeRepo.Save(s.config)
}

// RemoveGlobalServer rimuove un server globale
func (s *MCPService) RemoveGlobalServer(name string) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	if !s.config.RemoveGlobalServer(name) {
		return fmt.Errorf("server '%s' non trovato", name)
	}

	return s.claudeRepo.Save(s.config)
}

// RemoveProjectServer rimuove un server da un progetto
func (s *MCPService) RemoveProjectServer(projectPath, name string) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	project, exists := s.config.GetProject(projectPath)
	if !exists {
		return fmt.Errorf("progetto '%s' non trovato", projectPath)
	}

	if !project.RemoveServer(name) {
		return fmt.Errorf("server '%s' non trovato nel progetto", name)
	}

	return s.claudeRepo.Save(s.config)
}

// UpdateGlobalServer aggiorna un server globale
func (s *MCPService) UpdateGlobalServer(name string, server domain.MCPServer) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	if _, exists := s.config.GlobalServers[name]; !exists {
		return fmt.Errorf("server '%s' non trovato", name)
	}

	s.config.GlobalServers[name] = server
	return s.claudeRepo.Save(s.config)
}

// UpdateProjectServer aggiorna un server di progetto
func (s *MCPService) UpdateProjectServer(projectPath, name string, server domain.MCPServer) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	project, exists := s.config.GetProject(projectPath)
	if !exists {
		return fmt.Errorf("progetto '%s' non trovato", projectPath)
	}

	if _, serverExists := project.MCPServers[name]; !serverExists {
		return fmt.Errorf("server '%s' non trovato nel progetto", name)
	}

	project.MCPServers[name] = server
	return s.claudeRepo.Save(s.config)
}

// MoveServerToGlobal sposta un server da un progetto a globale
func (s *MCPService) MoveServerToGlobal(projectPath, name string) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	project, exists := s.config.GetProject(projectPath)
	if !exists {
		return fmt.Errorf("progetto '%s' non trovato", projectPath)
	}

	server, exists := project.GetServer(name)
	if !exists {
		return fmt.Errorf("server '%s' non trovato nel progetto", name)
	}

	// Aggiungi a globali e rimuovi dal progetto
	s.config.AddGlobalServer(name, server)
	project.RemoveServer(name)

	return s.claudeRepo.Save(s.config)
}

// MoveServerToProject sposta un server globale a un progetto
func (s *MCPService) MoveServerToProject(name, projectPath string) error {
	if s.config == nil {
		return fmt.Errorf("configurazione non caricata")
	}

	server, exists := s.config.GetGlobalServer(name)
	if !exists {
		return fmt.Errorf("server globale '%s' non trovato", name)
	}

	project := s.config.GetOrCreateProject(projectPath)

	// Aggiungi al progetto e rimuovi dai globali
	project.AddServer(name, server)
	s.config.RemoveGlobalServer(name)

	return s.claudeRepo.Save(s.config)
}

// GetEffectiveServers restituisce i server effettivi per un progetto
func (s *MCPService) GetEffectiveServers(projectPath string) (map[string]domain.MCPServer, error) {
	if s.config == nil {
		return nil, fmt.Errorf("configurazione non caricata")
	}

	return s.projectRepo.GetEffectiveServers(s.config, projectPath)
}
