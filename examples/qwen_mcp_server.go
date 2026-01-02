package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	// In a real implementation, you would import the actual Qwen SDK or HTTP client
	// For this example, we'll simulate the API calls
)

// QwenMCPHandler handles MCP requests for Qwen
type QwenMCPHandler struct {
	apiKey  string
	baseURL string
}

// NewQwenMCPHandler creates a new Qwen MCP handler
func NewQwenMCPHandler() *QwenMCPHandler {
	// In a real implementation, this would load the API key from environment variables
	// For this example, we'll just use a placeholder
	apiKey := os.Getenv("QWEN_API_KEY")
	if apiKey == "" {
		// For demo purposes, we'll proceed without an API key
		log.Println("Warning: QWEN_API_KEY environment variable not set")
	}
	
	baseURL := os.Getenv("QWEN_BASE_URL")
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/api/v1" // Default Qwen endpoint
	}

	return &QwenMCPHandler{
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// HandleRequest handles an MCP request
func (q *QwenMCPHandler) HandleRequest(request []byte) ([]byte, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %w", err)
	}

	// Get the method from the request
	method, ok := req["method"].(string)
	if !ok {
		return q.createErrorResponse("1", "Invalid request: method is required"), nil
	}

	// Extract the ID for the response
	id, ok := req["id"].(string)
	if !ok {
		return q.createErrorResponse("", "Invalid request: id is required"), nil
	}

	switch method {
	case "call_tool":
		// Handle tool calling - in this case, calling Qwen API
		params, hasParams := req["params"].(map[string]interface{})
		if !hasParams {
			return q.createErrorResponse(id, "Invalid request: params is required"), nil
		}

		return q.handleCallTool(id, params)
	case "list_tools":
		// Return available tools
		return q.handleListTools(id)
	case "get_server_info":
		// Return server information
		return q.handleGetServerInfo(id)
	default:
		return q.createErrorResponse(id, fmt.Sprintf("Unknown method: %s", method)), nil
	}
}

// handleCallTool handles tool calling requests
func (q *QwenMCPHandler) handleCallTool(id string, params map[string]interface{}) ([]byte, error) {
	// Extract the tool name and arguments
	toolName, ok := params["name"].(string)
	if !ok {
		return q.createErrorResponse(id, "tool name is required"), nil
	}

	arguments, hasArgs := params["arguments"].(map[string]interface{})
	if !hasArgs {
		arguments = make(map[string]interface{})
	}

	switch toolName {
	case "qwen_chat_complete":
		return q.handleQwenChatComplete(id, arguments)
	case "qwen_text_embedding":
		return q.handleQwenTextEmbedding(id, arguments)
	default:
		return q.createErrorResponse(id, fmt.Sprintf("unknown tool: %s", toolName)), nil
	}
}

// handleQwenChatComplete handles chat completion requests to Qwen
func (q *QwenMCPHandler) handleQwenChatComplete(id string, args map[string]interface{}) ([]byte, error) {
	// Extract parameters
	model, ok := args["model"].(string)
	if !ok {
		model = "qwen-max" // Default model
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

	temperature, hasTemp := args["temperature"].(float64)
	if !hasTemp {
		temperature = 0.7 // Default temperature
	}

	// In a real implementation, this would make an HTTP request to the Qwen API
	// For this example, we'll simulate the response
	responseContent := fmt.Sprintf("This is a simulated response from model %s with temperature %f", model, temperature)
	
	result := map[string]interface{}{
		"content": responseContent,
		"model":   model,
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
	}

	return q.createSuccessResponse(id, result), nil
}

// handleQwenTextEmbedding handles text embedding requests to Qwen
func (q *QwenMCPHandler) handleQwenTextEmbedding(id string, args map[string]interface{}) ([]byte, error) {
	text, ok := args["text"].(string)
	if !ok {
		text = "Default text for embedding"
	}

	model, ok := args["model"].(string)
	if !ok {
		model = "text-embedding-v1" // Default embedding model
	}

	// Simulate embedding response - in a real implementation, this would call the Qwen API
	// Generate a simple mock embedding based on the input text
	embedding := generateMockEmbedding(text)

	result := map[string]interface{}{
		"embedding": embedding,
		"text":      text,
		"model":     model,
	}

	return q.createSuccessResponse(id, result), nil
}

// handleListTools returns the list of available tools
func (q *QwenMCPHandler) handleListTools(id string) ([]byte, error) {
	tools := []map[string]interface{}{
		{
			"name":        "qwen_chat_complete",
			"description": "Send a chat message to Qwen and get a response",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"model": map[string]interface{}{
						"type":        "string",
						"description": "The model to use (e.g., qwen-max, qwen-plus)",
						"default":     "qwen-max",
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
						"default":     0.7,
					},
				},
				"required": []string{"messages"},
			},
		},
		{
			"name":        "qwen_text_embedding",
			"description": "Generate embeddings for text using Qwen",
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
						"default":     "text-embedding-v1",
					},
				},
				"required": []string{"text"},
			},
		},
	}

	return q.createSuccessResponse(id, tools), nil
}

// handleGetServerInfo returns server information
func (q *QwenMCPHandler) handleGetServerInfo(id string) ([]byte, error) {
	info := map[string]interface{}{
		"name":        "Qwen MCP Server",
		"version":     "1.0.0",
		"description": "MCP server for interacting with Qwen AI models",
		"provider":    "Qwen",
		"base_url":    q.baseURL,
		"tools":       []string{"qwen_chat_complete", "qwen_text_embedding"},
	}

	return q.createSuccessResponse(id, info), nil
}

// createSuccessResponse creates a success response
func (q *QwenMCPHandler) createSuccessResponse(id string, result interface{}) []byte {
	response := map[string]interface{}{
		"id":     id,
		"result": result,
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

// createErrorResponse creates an error response
func (q *QwenMCPHandler) createErrorResponse(id string, message string) []byte {
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

// generateMockEmbedding creates a mock embedding based on the input
func generateMockEmbedding(text string) []float64 {
	// In a real implementation, this would call the Qwen embedding API
	// For this example, we'll generate a simple deterministic mock embedding
	embedding := make([]float64, 16) // 16-dimensional embedding
	for i := 0; i < len(embedding); i++ {
		if i < len(text) {
			embedding[i] = float64(text[i%len(text)]) / 255.0
		} else {
			embedding[i] = float64(i) * 0.1
		}
	}
	return embedding
}

// Run starts the Qwen MCP server in stdio mode
func (q *QwenMCPHandler) Run() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Handle the request
		response, err := q.HandleRequest([]byte(line))
		if err != nil {
			errorResponse := q.createErrorResponse("unknown", err.Error())
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
	handler := NewQwenMCPHandler()
	handler.Run()
}