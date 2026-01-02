# MCOP Example MCP Servers

This directory contains example implementations of MCP (Model Context Protocol) servers that can be used with MCOP.

## Available Examples

### 1. Qwen MCP Server

The Qwen MCP server provides an interface to interact with Qwen AI models through the MCP protocol.

#### Features:
- Chat completion using Qwen models
- Text embeddings using Qwen embedding models
- Configurable model selection

#### Setup:
```bash
# Set your Qwen API key
export QWEN_API_KEY=your_api_key_here

# Run the server
go run examples/qwen_mcp_server.go
```

#### Tools:
- `qwen_chat_complete`: Send chat messages to Qwen and get responses
- `qwen_text_embedding`: Generate text embeddings using Qwen models

### 2. CLI MCP Server

The CLI MCP server allows safe execution of command-line tools through the MCP protocol.

#### Features:
- Execute predefined shell commands
- Read and write files (with security restrictions)
- Configurable list of allowed commands

#### Setup:
```bash
# Run the server
go run examples/cli_mcp_server.go
```

#### Tools:
- `execute_command`: Execute shell commands safely
- `read_file`: Read file contents
- `write_file`: Write content to files

## How to Use with MCOP

Once you have an MCP server running, you can connect MCOP to it in several ways:

### Using the TUI:
```bash
# Start MCOP with TUI
./mcop
```

Then use the connect command within the TUI.

### Using CLI command:
```bash
# Connect to a server directly
./mcop connect stdio://go run examples/qwen_mcp_server.go
```

### Adding to configuration:
You can also add servers to your configuration file (`config/default.json`):

```json
{
  "servers": [
    {
      "id": "qwen-server",
      "name": "Qwen Server",
      "url": "stdio://go run ./examples/qwen_mcp_server.go",
      "description": "Qwen AI MCP Server"
    },
    {
      "id": "cli-server",
      "name": "CLI Server",
      "url": "stdio://go run ./examples/cli_mcp_server.go",
      "description": "CLI Tools MCP Server"
    }
  ]
}
```

## Creating Your Own MCP Server

Use the generator command to create your own MCP server:

```bash
./mcop generate MyCustomServer --description "My custom MCP server" --api-endpoint "https://api.example.com/v1"
```

This will create a boilerplate server in the `generated-servers/` directory that you can customize for your specific needs.

## Security Considerations

When running MCP servers:
- Be cautious about which commands are allowed in CLI servers
- Protect API keys and sensitive information
- Validate all inputs to prevent code injection
- Only connect to trusted MCP servers from AI models