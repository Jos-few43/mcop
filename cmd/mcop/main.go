package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"mcop/src/config"
	"mcop/src/discovery"
	"mcop/src/generator"
	"mcop/src/model"
	"mcop/src/types"
)

var rootCmd = &cobra.Command{
	Use:   "mcop",
	Short: "MCOP - MCP Operations Monitor",
	Long:  `MCOP is a Terminal User Interface application for monitoring Model Context Protocol servers.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start the TUI application
		startTUI()
	},
}

var connectCmd = &cobra.Command{
	Use:   "connect [url]",
	Short: "Connect to an MCP server",
	Long:  `Connect to an MCP server by URL`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			startTUIWithServer(args[0])
		} else {
			startTUI()
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured MCP servers",
	Long:  `List all configured MCP servers from the configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig("")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Configured MCP Servers:")
		for i, server := range cfg.Servers {
			status := "stopped" // Default status
			if server.Status != "" {
				status = server.Status
			}
			fmt.Printf("%d. %s (%s) - %s\n", i+1, server.Name, status, server.URL)
		}
	},
}

var addCmd = &cobra.Command{
	Use:   "add [name] [url]",
	Short: "Add a new MCP server to configuration",
	Long:  `Add a new MCP server to the configuration`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		url := args[1]

		cfg, err := config.LoadConfig("")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Generate a simple ID from the name
		id := generateID(name)

		// Check if server already exists
		for _, server := range cfg.Servers {
			if server.ID == id || server.Name == name {
				fmt.Printf("Server with name '%s' already exists\n", name)
				os.Exit(1)
			}
		}

		// Add the new server
		cfg.AddServer(config.MCPServer{
			ID:                id,
			Name:              name,
			URL:               url,
			Status:            "stopped",
			Description:       "Added via command line",
			ActiveConnections: 0,
			Tools:             []string{},
			StartTime:         nil,
			ResponseTime:      nil,
		})

		// Save the updated config
		err = cfg.SaveConfig("config/default.json")
		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Added server '%s' with ID '%s'\n", name, id)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove [id]",
	Short: "Remove an MCP server from configuration",
	Long:  `Remove an MCP server from the configuration`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

		cfg, err := config.LoadConfig("")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Remove the server
		removed := cfg.RemoveServer(serverID)
		if !removed {
			fmt.Printf("Server with ID '%s' not found\n", serverID)
			os.Exit(1)
		}

		// Save the updated config
		err = cfg.SaveConfig("config/default.json")
		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Removed server with ID '%s'\n", serverID)
	},
}

var runCmd = &cobra.Command{
	Use:   "run [server-id]",
	Short: "Run a specific MCP server without TUI",
	Long:  `Run a specific MCP server directly without the TUI`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverID := args[0]

		cfg, err := config.LoadConfig("")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Find the server
		var targetServer *model.MCPServer
		for i := range cfg.Servers {
			if cfg.Servers[i].ID == serverID {
				// Convert config.MCPServer to model.MCPServer
				convertedServer := model.MCPServer{
					ID:                cfg.Servers[i].ID,
					Name:              cfg.Servers[i].Name,
					URL:               cfg.Servers[i].URL,
					Status:            cfg.Servers[i].Status,
					StartTime:         time.Time{}, // Initialize as zero time
					ResponseTime:      0, // Initialize as 0 duration
					ActiveConnections: cfg.Servers[i].ActiveConnections,
					Description:       cfg.Servers[i].Description,
					Tools:             cfg.Servers[i].Tools,
				}
				targetServer = &convertedServer
				break
			}
		}

		if targetServer == nil {
			fmt.Printf("Server with ID '%s' not found\n", serverID)
			os.Exit(1)
		}

		// Run the server directly
		if targetServer.URL[:7] == "stdio://" {
			command := targetServer.URL[8:] // Remove "stdio://" prefix
			fmt.Printf("Running server command: %s\n", command)
			// For now, we'll just print the command - in a real implementation
			// this would execute the command directly
			fmt.Printf("This would run: %s\n", command)
		} else {
			fmt.Printf("Unsupported protocol for direct execution: %s\n", targetServer.URL[:7])
			os.Exit(1)
		}
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate [name]",
	Short: "Generate a new MCP server implementation",
	Long:  `Generate a new MCP server implementation with boilerplate code`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		// Get additional flags
		description, _ := cmd.Flags().GetString("description")
		apiEndpoint, _ := cmd.Flags().GetString("api-endpoint")
		authType, _ := cmd.Flags().GetString("auth-type")

		// Create a basic tool definition
		tools := []generator.ToolDefinition{
			{
				Name:        "example_tool",
				Description: "An example tool for the server",
				Parameters:  map[string]interface{}{},
			},
		}

		// Create the template configuration
		templateConfig := generator.ServerTemplate{
			Name:        name,
			Description: description,
			Tools:       tools,
			APIEndpoint: apiEndpoint,
			AuthType:    authType,
		}

		// Create the generator
		gen := generator.NewGenerator("./generated-servers")

		// Generate the server
		err := gen.GenerateServer(templateConfig)
		if err != nil {
			fmt.Printf("Error generating server: %v\n", err)
			os.Exit(1)
		}
	},
}

var discoverCmd = &cobra.Command{
	Use:   "discover",
	Short: "Discover available MCP servers",
	Long:  `Discover available MCP servers on the local network and from configuration`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load configuration
		cfg, err := config.LoadConfig("")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		// Create discovery service
		discoveryService := discovery.NewDiscoveryService()

		// Convert config.MCPServer slice to types.MCPServer slice
		convertedServers := make([]types.MCPServer, len(cfg.Servers))
		for i, server := range cfg.Servers {
			convertedServers[i] = types.MCPServer{
				ID:                server.ID,
				Name:              server.Name,
				URL:               server.URL,
				Status:            server.Status,
				StartTime:         time.Time{},
				ResponseTime:      0,
				ActiveConnections: server.ActiveConnections,
				Description:       server.Description,
				Tools:             server.Tools,
			}
		}

		// Discover all servers
		servers, err := discoveryService.DiscoverAll(convertedServers)
		if err != nil {
			fmt.Printf("Error discovering servers: %v\n", err)
			os.Exit(1)
		}

		// Print discovered servers
		discoveryService.PrintDiscoveredServers(servers)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(discoverCmd)

	// Add flags for the generate command
	generateCmd.Flags().String("description", "An MCP server for integration", "Description of the server")
	generateCmd.Flags().String("api-endpoint", "https://api.example.com/v1", "API endpoint for the service")
	generateCmd.Flags().String("auth-type", "api_key", "Authentication type (api_key, oauth, etc.)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// generateID creates a simple ID from a name
func generateID(name string) string {
	// Simple implementation - in a real app, you'd want a more robust ID generation
	id := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			id += string(r)
		} else {
			id += "-"
		}
	}
	return id
}