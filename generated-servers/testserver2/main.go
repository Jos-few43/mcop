package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// TestServer2MCPHandler handles MCP requests for TestServer2
type TestServer2MCPHandler struct {
	apiKey string
	baseURL string
}

// NewTestServer2MCPHandler creates a new TestServer2 MCP handler
func NewTestServer2MCPHandler() *TestServer2MCPHandler {
	// Load configuration from environment variables
	envName := strings.ToUpper("TestServer2")
	apiKey := os.Getenv(envName + "_API_KEY")
	if apiKey == "" {
		log.Fatal(envName + "_API_KEY environment variable is required")
	}

	baseURL := os.Getenv(envName + "_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.example.com/v1" // Default API endpoint
	}

	return &TestServer2MCPHandler{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// HandleRequest handles an MCP request
func (h *TestServer2MCPHandler) HandleRequest(request []byte) ([]byte, error) {
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
func (h *TestServer2MCPHandler) handleCallTool(id string, params map[string]interface{}) ([]byte, error) {
	// Extract the tool name and arguments
	toolName, ok := params["name"].(string)
	if !ok {
		return h.createErrorResponse(id, "tool name is required"), nil
	}

	arguments, hasArgs := params["arguments"].(map[string]interface{})
	if !hasArgs {
		arguments = make(map[string]interface{})
	}

	switch toolName 
	case "example_tool":
		return h.handleExample_tool(id, arguments)
	default:
		return h.createErrorResponse(id, fmt.Sprintf("unknown tool: %%s", toolName)), nil
	}
}

// handleListTools returns the list of available tools
func (h *TestServer2MCPHandler) handleListTools(id string) ([]byte, error) {
	tools := []map[string]interface{}{ 
		{
			"name":        "example_tool",
			"description": "An example tool for the server",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[],
				"required": []string{},
			},
		},
	}

	return h.createSuccessResponse(id, tools), nil
}

// handleGetServerInfo returns server information
func (h *TestServer2MCPHandler) handleGetServerInfo(id string) ([]byte, error) {
	info := map[string]interface{}{
		"name":        "TestServer2 MCP Server",
		"version":     "1.0.0",
		"description": "Test server for Qwen",
		"tools": []string{ "example_tool",  },
	}

	return h.createSuccessResponse(id, info), nil
}

// handleExample_tool handles example_tool requests
func (h *example_toolMCPHandler) handleExample_tool(id string, args map[string]interface{}) ([]byte, error) {
	// Implement the logic for example_tool tool
	// This is where you would make actual API calls to example_tool

	return h.createSuccessResponse(id, map[string]interface{}{
		"result": fmt.Sprintf("Mock response for example_tool with arguments: %%v", args),
	}), nil
}

// createSuccessResponse creates a success response
func (h *TestServer2MCPHandler) createSuccessResponse(id string, result interface{}) []byte {
	response := map[string]interface{}{
		"id":     id,
		"result": result,
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

// createErrorResponse creates an error response
func (h *TestServer2MCPHandler) createErrorResponse(id string, message string) []byte {
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

// Run starts the TestServer2 MCP server in stdio mode
func (h *TestServer2MCPHandler) Run() {
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
	handler := NewTestServer2MCPHandler()
	handler.Run()
}
