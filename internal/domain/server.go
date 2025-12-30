package domain

// ServerType rappresenta il tipo di server MCP
type ServerType string

const (
	ServerTypeStdio ServerType = "stdio"
	ServerTypeHTTP  ServerType = "http"
	ServerTypeSSE   ServerType = "sse"
)

// MCPServer rappresenta un server MCP configurato
type MCPServer struct {
	Name    string            `json:"-"`
	Type    ServerType        `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	Timeout int               `json:"timeout,omitempty"`
}

// IsStdio verifica se il server è di tipo stdio
func (s *MCPServer) IsStdio() bool {
	return s.Type == ServerTypeStdio || s.Command != ""
}

// IsHTTP verifica se il server è di tipo HTTP o SSE
func (s *MCPServer) IsHTTP() bool {
	return s.Type == ServerTypeHTTP || s.Type == ServerTypeSSE || s.URL != ""
}

// Clone crea una copia del server
func (s *MCPServer) Clone() MCPServer {
	clone := MCPServer{
		Name:    s.Name,
		Type:    s.Type,
		Command: s.Command,
		URL:     s.URL,
		Timeout: s.Timeout,
	}

	if s.Args != nil {
		clone.Args = make([]string, len(s.Args))
		copy(clone.Args, s.Args)
	}

	if s.Headers != nil {
		clone.Headers = make(map[string]string, len(s.Headers))
		for k, v := range s.Headers {
			clone.Headers[k] = v
		}
	}

	if s.Env != nil {
		clone.Env = make(map[string]string, len(s.Env))
		for k, v := range s.Env {
			clone.Env[k] = v
		}
	}

	return clone
}
