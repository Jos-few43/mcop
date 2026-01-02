package servers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// GenericLLMHandler handles MCP requests for generic LLM interactions
type GenericLLMHandler struct {
	modelProvider string
	apiKey        string
	baseURL       string
}

// NewGenericLLMHandler creates a new generic LLM MCP handler
func NewGenericLLMHandler() *GenericLLMHandler {
	modelProvider := os.Getenv("MODEL_PROVIDER")
	if modelProvider == "" {
		modelProvider = "openai" // Default to a common provider
	}
	
	apiKey := os.Getenv("MODEL_API_KEY")
	if apiKey == "" {
		log.Fatal("MODEL_API_KEY environment variable is required")
	}
	
	baseURL := os.Getenv("MODEL_BASE_URL")
	if baseURL == "" {
		// Default to common OpenAI-style endpoint
		baseURL = "https://api.openai.com/v1"
	}

	return &GenericLLMHandler{
		modelProvider: modelProvider,
		apiKey:        apiKey,
		baseURL:       baseURL,
	}
}

// HandleRequest handles an MCP request
func (g *GenericLLMHandler) HandleRequest(request []byte) ([]byte, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Get the method from the request
	method, ok := req["method"].(string)
	if !ok {
		return g.createErrorResponse("1", "Invalid request: method is required"), nil
	}

	// Extract the ID for the response
	id, ok := req["id"].(string)
	if !ok {
		return g.createErrorResponse("", "Invalid request: id is required"), nil
	}

	switch method {
	case "call_tool":
		// Handle tool calling
		params, hasParams := req["params"].(map[string]interface{})
		if !hasParams {
			return g.createErrorResponse(id, "Invalid request: params is required"), nil
		}

		return g.handleCallTool(id, params)
	case "list_tools":
		// Return available tools
		return g.handleListTools(id)
	case "get_server_info":
		// Return server information
		return g.handleGetServerInfo(id)
	default:
		return g.createErrorResponse(id, fmt.Sprintf("Unknown method: %s", method)), nil
	}
}

// handleCallTool handles tool calling requests
func (g *GenericLLMHandler) handleCallTool(id string, params map[string]interface{}) ([]byte, error) {
	// Extract the tool name and arguments
	toolName, ok := params["name"].(string)
	if !ok {
		return g.createErrorResponse(id, "tool name is required"), nil
	}

	arguments, hasArgs := params["arguments"].(map[string]interface{})
	if !hasArgs {
		arguments = make(map[string]interface{})
	}

	switch toolName {
	case "chat_complete":
		return g.handleChatComplete(id, arguments)
	case "text_embedding":
		return g.handleTextEmbedding(id, arguments)
	case "list_models":
		return g.handleListModels(id)
	default:
		return g.createErrorResponse(id, fmt.Sprintf("unknown tool: %s", toolName)), nil
	}
}

// handleChatComplete handles chat completion requests
func (g *GenericLLMHandler) handleChatComplete(id string, args map[string]interface{}) ([]byte, error) {
	// Prepare request to LLM API
	model, ok := args["model"].(string)
	if !ok {
		model = "gpt-3.5-turbo" // Default model
	}

	messages, ok := args["messages"].([]interface{})
	if !ok {
		messages = []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": "Hello",
			},
		}
	}

	// Add other parameters if they exist
	extraParams := make(map[string]interface{})
	if temperature, exists := args["temperature"]; exists {
		extraParams["temperature"] = temperature
	}
	if maxTokens, exists := args["max_tokens"]; exists {
		extraParams["max_tokens"] = maxTokens
	}

	// Mock response - in a real implementation, this would call the actual LLM API
	// and return the result
	return g.createSuccessResponse(id, map[string]interface{}{
		"result": fmt.Sprintf("Mock response for chat completion with model: %s", model),
		"model":  model,
		"provider": g.modelProvider,
	}), nil
}

// handleTextEmbedding handles text embedding requests
func (g *GenericLLMHandler) handleTextEmbedding(id string, args map[string]interface{}) ([]byte, error) {
	text, ok := args["text"].(string)
	if !ok {
		text = "Default text for embedding"
	}

	model, ok := args["model"].(string)
	if !ok {
		model = "text-embedding-ada-002" // Default embedding model
	}

	// Mock embedding response - in a real implementation, this would call the API
	return g.createSuccessResponse(id, map[string]interface{}{
		"embedding": []float64{0.1, 0.2, 0.3, 0.4, 0.5}, // Mock embedding
		"text":      text,
		"model":     model,
		"provider":  g.modelProvider,
	}), nil
}

// handleListModels returns the list of available models
func (g *GenericLLMHandler) handleListModels(id string) ([]byte, error) {
	// Mock response - in a real implementation, this would call the API to get models
	models := []string{
		"gpt-4",
		"gpt-3.5-turbo", 
		"text-embedding-ada-002",
		"gpt-4-turbo",
	}
	
	if g.modelProvider == "qwen" {
		models = []string{
			"qwen-max",
			"qwen-plus",
			"qwen-turbo",
			"text-embedding-v1",
		}
	}

	return g.createSuccessResponse(id, models), nil
}

// handleListTools returns the list of available tools
func (g *GenericLLMHandler) handleListTools(id string) ([]byte, error) {
	tools := []map[string]interface{}{
		{
			"name":        "chat_complete",
			"description": "Send a chat message to the LLM and get a response",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model": map[string]interface{}{
						"type":        "string",
						"description": "The model to use for completion",
						"default":     "gpt-3.5-turbo",
					},
					"messages": map[string]interface{}{
						"type":        "array",
						"description": "Array of messages in the conversation",
					},
					"temperature": map[string]interface{}{
						"type":        "number",
						"description": "Controls randomness in the response",
						"minimum":     0,
						"maximum":     1,
					},
					"max_tokens": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of tokens to generate",
					},
				},
				"required": []string{"messages"},
			},
		},
		{
			"name":        "text_embedding",
			"description": "Generate embeddings for text",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type":        "string",
						"description": "The text to embed",
					},
					"model": map[string]interface{}{
						"type":        "string",
						"description": "The embedding model to use",
						"default":     "text-embedding-ada-002",
					},
				},
				"required": []string{"text"},
			},
		},
		{
			"name":        "list_models",
			"description": "Get the list of available models for the provider",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	return g.createSuccessResponse(id, tools), nil
}

// handleGetServerInfo returns server information
func (g *GenericLLMHandler) handleGetServerInfo(id string) ([]byte, error) {
	info := map[string]interface{}{
		"name":        "Generic LLM MCP Server",
		"version":     "1.0.0",
		"description": "MCP server for interacting with various LLM providers",
		"provider":    g.modelProvider,
		"tools":       []string{"chat_complete", "text_embedding", "list_models"},
	}

	return g.createSuccessResponse(id, info), nil
}

// createSuccessResponse creates a success response
func (g *GenericLLMHandler) createSuccessResponse(id string, result interface{}) []byte {
	response := map[string]interface{}{
		"id":     id,
		"result": result,
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

// createErrorResponse creates an error response
func (g *GenericLLMHandler) createErrorResponse(id string, message string) []byte {
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

// Run starts the generic LLM MCP server in stdio mode
func (g *GenericLLMHandler) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle the request
		response, err := g.HandleRequest([]byte(line))
		if err != nil {
			errorResponse := g.createErrorResponse("unknown", err.Error())
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