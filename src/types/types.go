package types

import (
	"time"
)

// MCPServer represents an MCP server instance
type MCPServer struct {
	ID                string
	Name              string
	URL               string
	Status            string // "running", "stopped", "error", "connecting"
	StartTime         time.Time
	ResponseTime      time.Duration
	ActiveConnections int
	Description       string
	Tools             []string
}

// Connection represents an active connection to an MCP server
type Connection struct {
	ID        string
	ServerID  string
	Connected time.Time
	Status    string // "active", "idle", "error"
	LastUsed  time.Time
}