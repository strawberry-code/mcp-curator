package domain

// Scope rappresenta dove è definito un server MCP
type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
)

// Configuration è l'aggregate root che gestisce tutte le configurazioni MCP
type Configuration struct {
	GlobalServers  map[string]MCPServer
	Projects       map[string]*Project
	ClaudeJsonPath string
}

// NewConfiguration crea una nuova configurazione vuota
func NewConfiguration(claudeJsonPath string) *Configuration {
	return &Configuration{
		GlobalServers:  make(map[string]MCPServer),
		Projects:       make(map[string]*Project),
		ClaudeJsonPath: claudeJsonPath,
	}
}

// GetGlobalServer restituisce un server globale per nome
func (c *Configuration) GetGlobalServer(name string) (MCPServer, bool) {
	server, ok := c.GlobalServers[name]
	if ok {
		server.Name = name
	}
	return server, ok
}

// AddGlobalServer aggiunge un server globale
func (c *Configuration) AddGlobalServer(name string, server MCPServer) {
	if c.GlobalServers == nil {
		c.GlobalServers = make(map[string]MCPServer)
	}
	server.Name = name
	c.GlobalServers[name] = server
}

// RemoveGlobalServer rimuove un server globale
func (c *Configuration) RemoveGlobalServer(name string) bool {
	if _, ok := c.GlobalServers[name]; ok {
		delete(c.GlobalServers, name)
		return true
	}
	return false
}

// GetProject restituisce un progetto per path
func (c *Configuration) GetProject(path string) (*Project, bool) {
	project, ok := c.Projects[path]
	return project, ok
}

// AddProject aggiunge un progetto
func (c *Configuration) AddProject(project *Project) {
	if c.Projects == nil {
		c.Projects = make(map[string]*Project)
	}
	c.Projects[project.Path] = project
}

// GetOrCreateProject restituisce un progetto esistente o ne crea uno nuovo
func (c *Configuration) GetOrCreateProject(path string) *Project {
	if project, ok := c.Projects[path]; ok {
		return project
	}
	project := NewProject(path)
	c.AddProject(project)
	return project
}

// GetEffectiveServers restituisce i server effettivi per un progetto (merge)
// Ordine di precedenza: globali < project settings < .mcp.json < .mcp.local.json
func (c *Configuration) GetEffectiveServers(projectPath string) map[string]MCPServer {
	result := make(map[string]MCPServer)

	// 1. Copia server globali
	for name, server := range c.GlobalServers {
		s := server.Clone()
		s.Name = name
		result[name] = s
	}

	// 2. Sovrascrivi con server del progetto (se esiste)
	if project, ok := c.Projects[projectPath]; ok {
		for name, server := range project.MCPServers {
			s := server.Clone()
			s.Name = name
			result[name] = s
		}
	}

	return result
}

// GlobalServerNames restituisce i nomi dei server globali in ordine
func (c *Configuration) GlobalServerNames() []string {
	names := make([]string, 0, len(c.GlobalServers))
	for name := range c.GlobalServers {
		names = append(names, name)
	}
	return names
}

// ProjectPaths restituisce i path dei progetti in ordine
func (c *Configuration) ProjectPaths() []string {
	paths := make([]string, 0, len(c.Projects))
	for path := range c.Projects {
		paths = append(paths, path)
	}
	return paths
}
