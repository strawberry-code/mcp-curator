package domain

import "path/filepath"

// Project rappresenta un progetto con configurazione MCP
type Project struct {
	Path        string               `json:"-"`
	Name        string               `json:"-"`
	MCPServers  map[string]MCPServer `json:"mcpServers,omitempty"`
	HasMCPJson  bool                 `json:"-"`
	HasMCPLocal bool                 `json:"-"`
}

// NewProject crea un nuovo progetto dal path
func NewProject(path string) *Project {
	return &Project{
		Path:       path,
		Name:       filepath.Base(path),
		MCPServers: make(map[string]MCPServer),
	}
}

// GetServer restituisce un server per nome
func (p *Project) GetServer(name string) (MCPServer, bool) {
	server, ok := p.MCPServers[name]
	if ok {
		server.Name = name
	}
	return server, ok
}

// AddServer aggiunge un server al progetto
func (p *Project) AddServer(name string, server MCPServer) {
	if p.MCPServers == nil {
		p.MCPServers = make(map[string]MCPServer)
	}
	server.Name = name
	p.MCPServers[name] = server
}

// RemoveServer rimuove un server dal progetto
func (p *Project) RemoveServer(name string) bool {
	if _, ok := p.MCPServers[name]; ok {
		delete(p.MCPServers, name)
		return true
	}
	return false
}

// ServerCount restituisce il numero di server nel progetto
func (p *Project) ServerCount() int {
	return len(p.MCPServers)
}
