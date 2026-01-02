package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"mcop/src/types"
)

// MCPClient handles communication with MCP servers
type MCPClient struct {
	Server   types.MCPServer
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	stdout   io.ReadCloser
	ctx      context.Context
	cancel   context.CancelFunc
	connected bool
}

// MCPRequest represents an MCP request
type MCPRequest struct {
	Method  string      `json:"method"`
	ID      string      `json:"id"`
	Version string      `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(server types.MCPServer) *MCPClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &MCPClient{
		Server: server,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Connect establishes a connection to the MCP server
func (c *MCPClient) Connect() error {
	// Parse the server URL to determine connection method
	if c.Server.URL == "" {
		return fmt.Errorf("server URL is empty")
	}

	if c.Server.URL[:7] == "stdio://" {
		// Handle stdio-based connection
		command := c.Server.URL[8:] // Remove "stdio://" prefix
		parts := parseCommand(command)
		if len(parts) == 0 {
			return fmt.Errorf("invalid command: %s", command)
		}

		c.cmd = exec.CommandContext(c.ctx, parts[0], parts[1:]...)
		var err error
		c.stdin, err = c.cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdin pipe: %w", err)
		}

		c.stdout, err = c.cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdout pipe: %w", err)
		}

		if err := c.cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command: %w", err)
		}

		c.connected = true
		go c.readLoop()
		return nil
	}

	// For now, only stdio is supported
	return fmt.Errorf("unsupported protocol: %s", c.Server.URL[:7])
}

// Disconnect closes the connection to the MCP server
func (c *MCPClient) Disconnect() error {
	c.connected = false
	if c.cancel != nil {
		c.cancel()
	}
	if c.cmd != nil {
		return c.cmd.Process.Kill()
	}
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.stdout != nil {
		c.stdout.Close()
	}
	return nil
}

// readLoop handles reading responses from the MCP server
func (c *MCPClient) readLoop() {
	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		if !c.connected {
			break
		}
		line := scanner.Text()
		// Process the response - for now just log it
		fmt.Printf("Received: %s\n", line)
	}
}

// Call makes an RPC call to the MCP server
func (c *MCPClient) Call(method string, params interface{}) (*MCPResponse, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected to server")
	}

	request := MCPRequest{
		Method:  method,
		ID:      generateID(),
		Version: "1.0",
		Params:  params,
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Add newline as MCP typically uses newline-delimited JSON
	requestBytes = append(requestBytes, '\n')

	_, err = c.stdin.Write(requestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// For now, we're not handling the response properly, just returning a placeholder
	// In a real implementation, we would need to properly handle async responses
	return &MCPResponse{
		ID: request.ID,
	}, nil
}

// parseCommand splits a command string into parts
func parseCommand(command string) []string {
	// Simple parsing - in real implementation may need more sophisticated parsing
	var parts []string
	current := ""
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(command); i++ {
		char := command[i]

		if !inQuotes && (char == '\'' || char == '"') {
			inQuotes = true
			quoteChar = char
		} else if inQuotes && char == quoteChar {
			inQuotes = false
		} else if !inQuotes && char == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// generateID creates a unique ID for requests
func generateID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// IsConnected returns whether the client is connected
func (c *MCPClient) IsConnected() bool {
	return c.connected
}