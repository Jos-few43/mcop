package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// CLIMCPHandler handles MCP requests for CLI tools
type CLIMCPHandler struct {
	allowedCommands []string
}

// NewCLIMCPHandler creates a new CLI MCP handler
func NewCLIMCPHandler() *CLIMCPHandler {
	// Define allowed commands for security
	allowedCommands := []string{
		"ls", "cat", "echo", "date", "pwd", "whoami",
		"grep", "find", "head", "tail", "wc",
	}
	
	return &CLIMCPHandler{
		allowedCommands: allowedCommands,
	}
}

// HandleRequest handles an MCP request
func (c *CLIMCPHandler) HandleRequest(request []byte) ([]byte, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Get the method from the request
	method, ok := req["method"].(string)
	if !ok {
		return c.createErrorResponse("1", "Invalid request: method is required"), nil
	}

	// Extract the ID for the response
	id, ok := req["id"].(string)
	if !ok {
		return c.createErrorResponse("", "Invalid request: id is required"), nil
	}

	switch method {
	case "call_tool":
		// Handle tool calling - executing CLI commands
		params, hasParams := req["params"].(map[string]interface{})
		if !hasParams {
			return c.createErrorResponse(id, "Invalid request: params is required"), nil
		}

		return c.handleCallTool(id, params)
	case "list_tools":
		// Return available tools
		return c.handleListTools(id)
	case "get_server_info":
		// Return server information
		return c.handleGetServerInfo(id)
	default:
		return c.createErrorResponse(id, fmt.Sprintf("Unknown method: %s", method)), nil
	}
}

// handleCallTool handles tool calling requests
func (c *CLIMCPHandler) handleCallTool(id string, params map[string]interface{}) ([]byte, error) {
	// Extract the tool name and arguments
	toolName, ok := params["name"].(string)
	if !ok {
		return c.createErrorResponse(id, "tool name is required"), nil
	}

	arguments, hasArgs := params["arguments"].(map[string]interface{})
	if !hasArgs {
		arguments = make(map[string]interface{})
	}

	switch toolName {
	case "execute_command":
		return c.handleExecuteCommand(id, arguments)
	case "read_file":
		return c.handleReadFile(id, arguments)
	case "write_file":
		return c.handleWriteFile(id, arguments)
	default:
		return c.createErrorResponse(id, fmt.Sprintf("unknown tool: %s", toolName)), nil
	}
}

// handleExecuteCommand handles command execution requests
func (c *CLIMCPHandler) handleExecuteCommand(id string, args map[string]interface{}) ([]byte, error) {
	command, ok := args["command"].(string)
	if !ok {
		return c.createErrorResponse(id, "command is required"), nil
	}

	// Validate that the command is in the allowed list
	cmdParts := strings.Fields(command)
	if len(cmdParts) == 0 {
		return c.createErrorResponse(id, "empty command"), nil
	}
	
	allowed := false
	for _, allowedCmd := range c.allowedCommands {
		if cmdParts[0] == allowedCmd {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return c.createErrorResponse(id, fmt.Sprintf("command '%s' is not allowed", cmdParts[0])), nil
	}

	// Execute the command
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.CombinedOutput()
	
	result := map[string]interface{}{
		"command": command,
		"output":  string(output),
		"success": err == nil,
	}
	
	if err != nil {
		result["error"] = err.Error()
	}

	return c.createSuccessResponse(id, result), nil
}

// handleReadFile handles file reading requests
func (c *CLIMCPHandler) handleReadFile(id string, args map[string]interface{}) ([]byte, error) {
	path, ok := args["path"].(string)
	if !ok {
		return c.createErrorResponse(id, "path is required"), nil
	}

	// Security check: only allow reading in current directory or subdirectories
	if strings.HasPrefix(path, "../") || strings.HasPrefix(path, "/") {
		return c.createErrorResponse(id, "access denied: only relative paths allowed"), nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return c.createErrorResponse(id, fmt.Sprintf("failed to read file: %v", err)), nil
	}

	result := map[string]interface{}{
		"path":    path,
		"content": string(content),
		"size":    len(content),
	}

	return c.createSuccessResponse(id, result), nil
}

// handleWriteFile handles file writing requests
func (c *CLIMCPHandler) handleWriteFile(id string, args map[string]interface{}) ([]byte, error) {
	path, ok := args["path"].(string)
	if !ok {
		return c.createErrorResponse(id, "path is required"), nil
	}

	content, ok := args["content"].(string)
	if !ok {
		return c.createErrorResponse(id, "content is required"), nil
	}

	// Security check: only allow writing in current directory or subdirectories
	if strings.HasPrefix(path, "../") || strings.HasPrefix(path, "/") {
		return c.createErrorResponse(id, "access denied: only relative paths allowed"), nil
	}

	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return c.createErrorResponse(id, fmt.Sprintf("failed to write file: %v", err)), nil
	}

	result := map[string]interface{}{
		"path": path,
		"size": len([]byte(content)),
	}

	return c.createSuccessResponse(id, result), nil
}

// handleListTools returns the list of available tools
func (c *CLIMCPHandler) handleListTools(id string) ([]byte, error) {
	tools := []map[string]interface{}{
		{
			"name":        "execute_command",
			"description": "Execute a shell command safely",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The command to execute",
					},
				},
				"required": []string{"command"},
			},
		},
		{
			"name":        "read_file",
			"description": "Read the contents of a file",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to read (relative to current directory)",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			"name":        "write_file",
			"description": "Write content to a file",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to write (relative to current directory)",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to write to the file",
					},
				},
				"required": []string{"path", "content"},
			},
		},
	}

	return c.createSuccessResponse(id, tools), nil
}

// handleGetServerInfo returns server information
func (c *CLIMCPHandler) handleGetServerInfo(id string) ([]byte, error) {
	info := map[string]interface{}{
		"name":        "CLI MCP Server",
		"version":     "1.0.0",
		"description": "MCP server for executing CLI commands safely",
		"allowed_commands": c.allowedCommands,
		"tools":       []string{"execute_command", "read_file", "write_file"},
	}

	return c.createSuccessResponse(id, info), nil
}

// createSuccessResponse creates a success response
func (c *CLIMCPHandler) createSuccessResponse(id string, result interface{}) []byte {
	response := map[string]interface{}{
		"id":     id,
		"result": result,
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

// createErrorResponse creates an error response
func (c *CLIMCPHandler) createErrorResponse(id string, message string) []byte {
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

// Run starts the CLI MCP server in stdio mode
func (c *CLIMCPHandler) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle the request
		response, err := c.HandleRequest([]byte(line))
		if err != nil {
			errorResponse := c.createErrorResponse("unknown", err.Error())
			fmt.Println(string(errorResponse))
			continue
		}

		// Send the response
		fmt.Println(string(response))
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading stdin: %v", err)
	}
}

func main() {
	handler := NewCLIMCPHandler()
	handler.Run()
}