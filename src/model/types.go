package model

import (
	"time"
	"mcop/src/types"
)

// MCPServer represents an MCP server instance
type MCPServer = types.MCPServer

// Connection represents an active connection to an MCP server
type Connection = types.Connection

// Config holds application configuration
type Config struct {
	Servers      []MCPServer
	AutoRefresh  bool
	RefreshRate  time.Duration
	DefaultTheme string
}