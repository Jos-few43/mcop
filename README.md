# MCOP - MCP Operations Monitor

MCOP (Model Context Protocol Operations) is a Terminal User Interface application designed to monitor, manage, and visualize MCP (Model Context Protocol) servers and connections in real-time. Inspired by top and htop, MCOP provides a system-monitor-like interface for MCP operations.

## Features

- **Real-time Monitoring**: Monitor active MCP server connections and resource usage
- **Process-like View**: View MCP servers similar to how top/htop shows system processes
- **Interactive Management**: Start, stop, and configure MCP servers from the TUI
- **Connection Tracking**: Track active connections and their status
- **Performance Metrics**: Monitor response times, throughput, and health of MCP servers
- **Configuration Management**: Edit and manage MCP server configurations
- **Server Discovery**: Browse and connect to available MCP servers

## Installation

```bash
# Clone the repository
git clone https://github.com/your-username/mcop.git
cd mcop

# Build the application
go build -o mcop ./cmd/mcop

# Run the application
./mcop
```

## Development Setup

```bash
# Install dependencies
go mod tidy

# Build for development
go build -o mcop ./cmd/mcop

# Run directly
go run ./cmd/mcop
```

## Usage

```bash
# Launch MCOP TUI
./mcop

# Connect to specific server
./mcop --connect <server-url>

# Connect with config file
./mcop --config /path/to/config.json
```

## Key Controls

- `q` or `Ctrl+C`: Quit the application
- `↑/↓`: Navigate between MCP processes/connections
- `Enter`: View detailed information about selected connection
- `s`: Start/stop selected MCP server
- `r`: Refresh the MCP connection list
- `c`: Open configuration editor
- `d`: Disconnect selected connection
- `?`: Show help

## Motivation

MCOP provides a system-process-style monitoring interface for Model Context Protocol operations, giving developers and operators a familiar interface for managing MCP infrastructure similar to how system administrators use top/htop to monitor system resources.