package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/strawberry-code/mcp-curator/internal/domain"
)

// ParseServer converte un map[string]interface{} in MCPServer
func ParseServer(data interface{}) (domain.MCPServer, error) {
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

// ServerToMap converte MCPServer in map[string]interface{}
func ServerToMap(server domain.MCPServer) map[string]interface{} {
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

// LoadMCPFileServers carica i server da un file .mcp.json o .mcp.local.json
func LoadMCPFileServers(path string) map[string]domain.MCPServer {
	result := make(map[string]domain.MCPServer)

	data, err := os.ReadFile(path)
	if err != nil {
		return result
	}

	var raw struct {
		MCPServers map[string]interface{} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return result
	}

	for name, serverData := range raw.MCPServers {
		server, err := ParseServer(serverData)
		if err != nil {
			continue
		}
		server.Name = name
		result[name] = server
	}

	return result
}
