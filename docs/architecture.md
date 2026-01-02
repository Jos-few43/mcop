# MCOP - Model Context Protocol Operations Monitor

This is a TUI application for monitoring and managing MCP (Model Context Protocol) servers, inspired by top/htop.

## Project Structure

```
mcop/
├── cmd/
│   └── mcop/
│       └── main.go
├── src/
│   ├── ui/
│   │   ├── app.go
│   │   ├── views/
│   │   │   ├── main_view.go
│   │   │   ├── server_list.go
│   │   │   ├── detail_view.go
│   │   │   └── config_view.go
│   │   └── components/
│   │       ├── table.go
│   │       ├── status_bar.go
│   │       └── help_modal.go
│   ├── model/
│   │   ├── app_state.go
│   │   ├── mcp_server.go
│   │   ├── connection.go
│   │   └── config.go
│   └── mcp/
│       ├── client.go
│       ├── server_discovery.go
│       └── protocol.go
├── config/
│   └── default.json
├── docs/
│   └── architecture.md
├── tests/
│   └── integration_test.go
├── go.mod
└── README.md
```

## Components

### UI Layer
- `app.go`: Main application state and update loop
- `views/`: Different screens of the TUI
- `components/`: Reusable UI elements

### Model Layer
- `app_state.go`: Global application state
- `mcp_server.go`: MCP server representation
- `connection.go`: Connection tracking
- `config.go`: Configuration management

### MCP Layer
- `client.go`: MCP protocol client
- `server_discovery.go`: Discovering available MCP servers
- `protocol.go`: MCP protocol implementation