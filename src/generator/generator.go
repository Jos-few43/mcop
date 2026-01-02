package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ServerTemplate represents the template for generating an MCP server
type ServerTemplate struct {
	Name        string
	Description string
	Tools       []ToolDefinition
	APIEndpoint string
	AuthType    string
}

// ToolDefinition represents a tool that the server will implement
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// Generator handles MCP server generation
type Generator struct {
	OutputDir string
}

// NewGenerator creates a new MCP server generator
func NewGenerator(outputDir string) *Generator {
	return &Generator{
		OutputDir: outputDir,
	}
}

// GenerateServer generates a new MCP server based on the template
func (g *Generator) GenerateServer(templateConfig ServerTemplate) error {
	// Ensure output directory exists
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create the server directory
	serverDir := filepath.Join(g.OutputDir, strings.ToLower(templateConfig.Name))
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %w", err)
	}

	// Generate the main server file
	if err := g.generateMainServerFile(serverDir, templateConfig); err != nil {
		return fmt.Errorf("failed to generate main server file: %w", err)
	}

	// Generate the configuration file
	if err := g.generateConfigFile(serverDir, templateConfig); err != nil {
		return fmt.Errorf("failed to generate config file: %w", err)
	}

	// Generate the README
	if err := g.generateReadme(serverDir, templateConfig); err != nil {
		return fmt.Errorf("failed to generate README: %w", err)
	}

	fmt.Printf("Successfully generated MCP server: %s\n", templateConfig.Name)
	fmt.Printf("Server location: %s\n", serverDir)
	
	return nil
}

// generateMainServerFile generates the main server implementation
func (g *Generator) generateMainServerFile(serverDir string, config ServerTemplate) error {
	// Define the template for the main server file
	serverTemplate := `package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// {{.Name}}MCPHandler handles MCP requests for {{.Name}}
type {{.Name}}MCPHandler struct {
	apiKey string
	baseURL string
}

// New{{.Name}}MCPHandler creates a new {{.Name}} MCP handler
func New{{.Name}}MCPHandler() *{{.Name}}MCPHandler {
	// Load configuration from environment variables
	envName := strings.ToUpper("{{.Name}}")
	apiKey := os.Getenv(envName + "_API_KEY")
	if apiKey == "" {
		log.Fatal(envName + "_API_KEY environment variable is required")
	}

	baseURL := os.Getenv(envName + "_BASE_URL")
	if baseURL == "" {
		baseURL = "{{.APIEndpoint}}" // Default API endpoint
	}

	return &{{.Name}}MCPHandler{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// HandleRequest handles an MCP request
func (h *{{.Name}}MCPHandler) HandleRequest(request []byte) ([]byte, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %%w", err)
	}

	// Get the method from the request
	method, ok := req["method"].(string)
	if !ok {
		return h.createErrorResponse("1", "Invalid request: method is required"), nil
	}

	// Extract the ID for the response
	id, ok := req["id"].(string)
	if !ok {
		return h.createErrorResponse("", "Invalid request: id is required"), nil
	}

	switch method {
	case "call_tool":
		// Handle tool calling
		params, hasParams := req["params"].(map[string]interface{})
		if !hasParams {
			return h.createErrorResponse(id, "Invalid request: params is required"), nil
		}

		return h.handleCallTool(id, params)
	case "list_tools":
		// Return available tools
		return h.handleListTools(id)
	case "get_server_info":
		// Return server information
		return h.handleGetServerInfo(id)
	default:
		return h.createErrorResponse(id, fmt.Sprintf("Unknown method: %%s", method)), nil
	}
}

// handleCallTool handles tool calling requests
func (h *{{.Name}}MCPHandler) handleCallTool(id string, params map[string]interface{}) ([]byte, error) {
	// Extract the tool name and arguments
	toolName, ok := params["name"].(string)
	if !ok {
		return h.createErrorResponse(id, "tool name is required"), nil
	}

	arguments, hasArgs := params["arguments"].(map[string]interface{})
	if !hasArgs {
		arguments = make(map[string]interface{})
	}

	switch toolName {{range .Tools}}
	case "{{.Name}}":
		return h.handle{{title .Name}}(id, arguments){{end}}
	default:
		return h.createErrorResponse(id, fmt.Sprintf("unknown tool: %%s", toolName)), nil
	}
}

// handleListTools returns the list of available tools
func (h *{{.Name}}MCPHandler) handleListTools(id string) ([]byte, error) {
	tools := []map[string]interface{}{ {{range .Tools}}
		{
			"name":        "{{.Name}}",
			"description": "{{.Description}}",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": {{printf "%v" .Parameters}},
				"required": []string{},
			},
		},{{end}}
	}

	return h.createSuccessResponse(id, tools), nil
}

// handleGetServerInfo returns server information
func (h *{{.Name}}MCPHandler) handleGetServerInfo(id string) ([]byte, error) {
	info := map[string]interface{}{
		"name":        "{{.Name}} MCP Server",
		"version":     "1.0.0",
		"description": "{{.Description}}",
		"tools": []string{ {{range .Tools}}"{{.Name}}", {{end}} },
	}

	return h.createSuccessResponse(id, info), nil
}{{range .Tools}}

// handle{{title .Name}} handles {{.Name}} requests
func (h *{{.Name}}MCPHandler) handle{{title .Name}}(id string, args map[string]interface{}) ([]byte, error) {
	// Implement the logic for {{.Name}} tool
	// This is where you would make actual API calls to {{.Name}}

	return h.createSuccessResponse(id, map[string]interface{}{
		"result": fmt.Sprintf("Mock response for {{.Name}} with arguments: %%v", args),
	}), nil
}{{end}}

// createSuccessResponse creates a success response
func (h *{{.Name}}MCPHandler) createSuccessResponse(id string, result interface{}) []byte {
	response := map[string]interface{}{
		"id":     id,
		"result": result,
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

// createErrorResponse creates an error response
func (h *{{.Name}}MCPHandler) createErrorResponse(id string, message string) []byte {
	response := map[string]interface{}{
		"id": id,
		"error": map[string]interface{}{
			"code":    -32000,
			"message": message,
		},
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

// Run starts the {{.Name}} MCP server in stdio mode
func (h *{{.Name}}MCPHandler) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle the request
		response, err := h.HandleRequest([]byte(line))
		if err != nil {
			errorResponse := h.createErrorResponse("unknown", err.Error())
			fmt.Println(string(errorResponse))
			continue
		}

		// Send the response
		fmt.Println(string(response))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading stdin: %%v", err)
	}
}

func main() {
	handler := New{{.Name}}MCPHandler()
	handler.Run()
}
`

	// Create the template with functions
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"ToUpper": strings.ToUpper,
	}

	tmpl, err := template.New("server").Funcs(funcMap).Parse(serverTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse server template: %w", err)
	}

	// Create the output file
	outputFile := filepath.Join(serverDir, "main.go")
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create server file: %w", err)
	}
	defer file.Close()

	// Execute the template
	if err := tmpl.Execute(file, config); err != nil {
		return fmt.Errorf("failed to execute server template: %w", err)
	}

	return nil
}

// generateConfigFile generates a configuration file for the server
func (g *Generator) generateConfigFile(serverDir string, config ServerTemplate) error {
	configContent := fmt.Sprintf(`# %s MCP Server Configuration
# Set these environment variables to configure the server

# API Key for %s
%s_API_KEY=your_api_key_here

# Base URL for %s API
%s_BASE_URL=%s

# Additional configuration options
MODEL_NAME=default_model
TEMPERATURE=0.7
MAX_TOKENS=1024
`, config.Name, config.Name, strings.ToUpper(config.Name), 
	config.Name, strings.ToUpper(config.Name), config.APIEndpoint)

	configFile := filepath.Join(serverDir, ".env.example")
	return os.WriteFile(configFile, []byte(configContent), 0644)
}

// generateReadme generates a README file for the server
func (g *Generator) generateReadme(serverDir string, config ServerTemplate) error {
	readmeContent := fmt.Sprintf("# %s MCP Server\n\nThis is an MCP (Model Context Protocol) server for %s.\n\n## Overview\n%s\n\n## Prerequisites\n- Go 1.19 or higher\n- %s API key\n\n## Setup\n\n1. Set up your environment variables:\n   ```bash\n   cp .env.example .env\n   # Edit .env with your API key\n   ```\n\n2. Build the server:\n   ```bash\n   go mod init %s-server\n   go build -o %s-server .\n   ```\n\n## Usage\n\nYou can run the server directly:\n\n```bash\nMODEL_API_KEY=your_key_here go run main.go\n```\n\nOr build and run the binary:\n```bash\ngo build -o %s-server .\nMODEL_API_KEY=your_key_here ./server\n```\n\n## Tools\n\nThis server provides the following tools:\n%s\n\n## Contributing\nPull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.\n\n## License\nMIT\n",
		config.Name, config.Name, config.Description, config.Name,
		strings.ToLower(config.Name), strings.ToLower(config.Name),
		strings.ToLower(config.Name),
		g.generateToolsList(config.Tools))

	readmeFile := filepath.Join(serverDir, "README.md")
	return os.WriteFile(readmeFile, []byte(readmeContent), 0644)
}

// generateToolsList generates a markdown list of tools
func (g *Generator) generateToolsList(tools []ToolDefinition) string {
	var result string
	for _, tool := range tools {
		result += fmt.Sprintf("- `%s`: %s\n", tool.Name, tool.Description)
	}
	return result
}