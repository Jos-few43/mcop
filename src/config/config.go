package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPServer represents an MCP server configuration
type MCPServer struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	URL               string `json:"url"`
	Status            string `json:"status,omitempty"`
	StartTime         interface{} `json:"start_time,omitempty"`  // Using interface{} to avoid import cycle
	ResponseTime      interface{} `json:"response_time,omitempty"` // Using interface{} to avoid import cycle
	ActiveConnections int    `json:"active_connections,omitempty"`
	Description       string `json:"description"`
	Tools             []string `json:"tools,omitempty"`
}

// AppConfig represents the application configuration
type AppConfig struct {
	Servers       []MCPServer `json:"servers"`
	AutoRefresh   bool        `json:"auto_refresh"`
	RefreshRate   int         `json:"refresh_rate"`
	DefaultTheme  string      `json:"default_theme"`
	APIKeys       map[string]string `json:"api_keys,omitempty"`
	ServerConfigs map[string]ServerConfig `json:"server_configs,omitempty"`
}

// ServerConfig represents configuration specific to a server
type ServerConfig struct {
	APIKey      string            `json:"api_key,omitempty"`
	BaseURL     string            `json:"base_url,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Environment map[string]string `json:"environment,omitempty"`
}

// LoadConfig loads the application configuration from a file
func LoadConfig(configPath string) (*AppConfig, error) {
	if configPath == "" {
		// Default to config/default.json
		configPath = "config/default.json"
	}

	// Check if the config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// If the file doesn't exist, return a default config
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and set defaults
	if config.RefreshRate <= 0 {
		config.RefreshRate = 5 // Default to 5 seconds
	}

	// Load any environment-specific configurations
	config.loadEnvironmentVars()

	return &config, nil
}

// SaveConfig saves the application configuration to a file
func (c *AppConfig) SaveConfig(configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddServer adds a new server to the configuration
func (c *AppConfig) AddServer(server MCPServer) {
	// Check if server already exists
	for i, existingServer := range c.Servers {
		if existingServer.ID == server.ID {
			// Update existing server
			c.Servers[i] = server
			return
		}
	}

	// Add new server
	c.Servers = append(c.Servers, server)
}

// RemoveServer removes a server from the configuration
func (c *AppConfig) RemoveServer(serverID string) bool {
	for i, server := range c.Servers {
		if server.ID == serverID {
			// Remove the server
			c.Servers = append(c.Servers[:i], c.Servers[i+1:]...)
			
			// Remove associated config
			delete(c.ServerConfigs, serverID)
			
			return true
		}
	}
	return false
}

// GetServer retrieves a server by ID
func (c *AppConfig) GetServer(serverID string) *MCPServer {
	for _, server := range c.Servers {
		if server.ID == serverID {
			return &server
		}
	}
	return nil
}

// loadEnvironmentVars loads configuration from environment variables
func (c *AppConfig) loadEnvironmentVars() {
	// Load global API keys
	if apiKey := os.Getenv("MODEL_API_KEY"); apiKey != "" {
		if c.APIKeys == nil {
			c.APIKeys = make(map[string]string)
		}
		c.APIKeys["default"] = apiKey
	}
	
	if provider := os.Getenv("MODEL_PROVIDER"); provider != "" {
		if c.APIKeys == nil {
			c.APIKeys = make(map[string]string)
		}
		c.APIKeys["provider"] = provider
	}
	
	// Apply environment variables to server-specific configurations
	for i := range c.Servers {
		if c.Servers[i].URL == "" {
			// Set default URL from environment if available
			if defaultURL := os.Getenv("DEFAULT_MCP_URL"); defaultURL != "" {
				c.Servers[i].URL = defaultURL
			}
		}
	}
}

// DefaultConfig returns a default configuration
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Servers: []MCPServer{
			{
				ID:          "generic-llm-server",
				Name:        "Generic LLM Server",
				URL:         "stdio://go run ./src/mcp/servers/generic_llm.go",
				Status:      "stopped", // Default to stopped until user starts it
				Description: "Generic LLM server compatible with various providers (OpenAI, Qwen, etc.)",
				ActiveConnections: 0,
				Tools:       []string{"chat_complete", "text_embedding", "list_models"},
			},
			{
				ID:          "github-server",
				Name:        "GitHub Integration Server",
				URL:         "stdio://npx @modelcontextprotocol/server-github",
				Status:      "stopped",
				Description: "MCP server for GitHub operations",
				ActiveConnections: 0,
				Tools:       []string{"get_repo_info", "create_issue", "search_issues"},
			},
		},
		AutoRefresh: true,
		RefreshRate: 5,
		APIKeys:     make(map[string]string),
		ServerConfigs: make(map[string]ServerConfig),
	}
}

// Validate validates the configuration
func (c *AppConfig) Validate() error {
	for _, server := range c.Servers {
		if server.ID == "" {
			return fmt.Errorf("server ID cannot be empty")
		}
		if server.URL == "" {
			return fmt.Errorf("server URL cannot be empty for server %s", server.ID)
		}
	}
	
	return nil
}

// GetServerConfig returns the configuration for a specific server
func (c *AppConfig) GetServerConfig(serverID string) ServerConfig {
	config, exists := c.ServerConfigs[serverID]
	if !exists {
		return ServerConfig{
			Parameters:  make(map[string]interface{}),
			Environment: make(map[string]string),
		}
	}
	return config
}

// SetServerConfig sets the configuration for a specific server
func (c *AppConfig) SetServerConfig(serverID string, config ServerConfig) {
	if c.ServerConfigs == nil {
		c.ServerConfigs = make(map[string]ServerConfig)
	}
	c.ServerConfigs[serverID] = config
}

// GetServersAsModelServers converts the config servers to model servers format
// This function is designed to be used when importing in model package to avoid cycle
func (c *AppConfig) GetServersAsModelServers() interface{} {
	// Return as a generic interface that model package will type assert
	return c.Servers
}