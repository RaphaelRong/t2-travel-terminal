package mcp

// Server describes a registered MCP server.
type Server struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Endpoint    string            `json:"endpoint"`
	Tools       []Tool            `json:"tools,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Tool describes an MCP tool exposed by a server.
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema string `json:"input_schema"`
}

// Registry maintains a list of MCP servers.
type Registry struct {
	servers map[string]Server
}

// NewRegistry creates an empty MCP registry.
func NewRegistry() *Registry {
	return &Registry{
		servers: make(map[string]Server),
	}
}

// Register adds an MCP server to the registry.
func (r *Registry) Register(s Server) {
	r.servers[s.Name] = s
}

// List returns all registered MCP servers.
func (r *Registry) List() []Server {
	list := make([]Server, 0, len(r.servers))
	for _, s := range r.servers {
		list = append(list, s)
	}
	return list
}
