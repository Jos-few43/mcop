# TestServer2 MCP Server

This is an MCP (Model Context Protocol) server for TestServer2.

## Overview
Test server for Qwen

## Prerequisites
- Go 1.19 or higher
- TestServer2 API key

## Setup

1. Set up your environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your API key
   ```

2. Build the server:
   ```bash
   go mod init testserver2-server
   go build -o testserver2-server .
   ```

## Usage

You can run the server directly:

```bash
MODEL_API_KEY=your_key_here go run main.go
```

Or build and run the binary:
```bash
go build -o testserver2-server .
MODEL_API_KEY=your_key_here ./server
```

## Tools

This server provides the following tools:
- `example_tool`: An example tool for the server


## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
MIT
