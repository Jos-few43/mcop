package model

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"mcop/src/mcp"
	"mcop/src/config"
)

// AppState represents the main application state
type AppState struct {
	Servers           []MCPServer
	Connections       []Connection
	MCPConnections    map[string]*mcp.MCPClient // Map of server ID to MCP client
	SelectedIndex     int
	View              string // "list", "detail", "config"
	Error             string
	IsLoading         bool
	RefreshRate       int
	AutoRefresh       bool
	InitialServerURL  string
}

// AppModel is the main Bubble Tea model
type AppModel struct {
	State AppState
	Width int
	Height int
	Config *config.AppConfig
}

func NewAppModel() *AppModel {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		// If there's an error loading config, use default config
		fmt.Printf("Warning: failed to load config, using defaults: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Convert config.MCPServer to model.MCPServer
	servers := make([]MCPServer, len(cfg.Servers))
	for i, cfgServer := range cfg.Servers {
		startTime := time.Time{}
		responseTime := time.Duration(0)

		// Convert interface{} fields if possible
		if cfgServer.StartTime != nil {
			// Try to convert to time.Time if possible
			if ts, ok := cfgServer.StartTime.(time.Time); ok {
				startTime = ts
			}
		}

		if cfgServer.ResponseTime != nil {
			// Try to convert to time.Duration if possible
			if rt, ok := cfgServer.ResponseTime.(time.Duration); ok {
				responseTime = rt
			}
		}

		servers[i] = MCPServer{
			ID:                cfgServer.ID,
			Name:              cfgServer.Name,
			URL:               cfgServer.URL,
			Status:            cfgServer.Status,
			StartTime:         startTime,
			ResponseTime:      responseTime,
			ActiveConnections: cfgServer.ActiveConnections,
			Description:       cfgServer.Description,
			Tools:             cfgServer.Tools,
		}
	}

	return &AppModel{
		State: AppState{
			Servers:        servers,
			Connections:    []Connection{},
			MCPConnections: make(map[string]*mcp.MCPClient),
			SelectedIndex:  0,
			View:           "list",
			RefreshRate:    cfg.RefreshRate,
			AutoRefresh:    cfg.AutoRefresh,
		},
		Width:  80,
		Height: 24,
		Config: cfg,
	}
}

func (m *AppModel) SetInitialServerURL(url string) {
	m.State.InitialServerURL = url
}

// Init is called when the program starts
func (m *AppModel) Init() tea.Cmd {
	// The servers are already loaded from config in NewAppModel
	// Only load mock servers if config has no servers
	if len(m.State.Servers) == 0 {
		m.loadMockServers()
	}
	return nil
}

// Update handles messages and updates the model
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}
	return m, nil
}

// View renders the UI using the styled renderer
func (m *AppModel) View() string {
	// Instead of calling directly, we'll create a simple renderer here that can be overridden
	// For now, we'll implement the styled view directly
	return m.renderStyledView()
}

func (m *AppModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.State.SelectedIndex > 0 {
			m.State.SelectedIndex--
		}
	case "down", "j":
		if m.State.SelectedIndex < len(m.State.Servers)-1 {
			m.State.SelectedIndex++
		}
	case "enter":
		m.State.View = "detail"
	case "r":
		m.loadMockServers() // Refresh
	case "c":
		m.State.View = "config"
	case "s":
		if m.State.View == "list" && m.State.SelectedIndex < len(m.State.Servers) {
			m.ToggleServer(m.State.SelectedIndex)

			// The UI layer will handle logging
		}
	case "d":
		if m.State.View == "detail" && m.State.SelectedIndex < len(m.State.Servers) {
			m.DisconnectServer(m.State.SelectedIndex)
		}
	case "esc":
		m.State.View = "list"
	}
	return m, nil
}

func (m *AppModel) renderStyledView() string {
	// Import locally to avoid cycle
	// We'll create the styled view directly here
	return m.renderServerList()
}

func (m *AppModel) renderServerList() string {
	// For now, we'll return a simple string until we resolve the import cycle
	return m.renderSimpleListView()
}

func (m *AppModel) renderSimpleListView() string {
	// Enhanced list view implementation using lipgloss
	var sb strings.Builder

	// Render title
	title := "MCOP - Model Context Protocol Operations Monitor"
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Render table header
	sb.WriteString(fmt.Sprintf("%-4s %-30s %-12s %-10s %s\n", "ID", "NAME", "STATUS", "CONNS", "URL"))
	sb.WriteString(fmt.Sprintf("%-4s %-30s %-12s %-10s %s\n", "--", "----", "------", "-----", "---"))

	// Render server list
	for i, server := range m.State.Servers {
		prefix := "  "
		if i == m.State.SelectedIndex {
			prefix = " >"
		}

		statusSymbol := "●" // running
		if server.Status == "stopped" {
			statusSymbol = "○" // stopped
		} else if server.Status == "error" {
			statusSymbol = "●" // error (red)
		}

		name := server.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		sb.WriteString(fmt.Sprintf("%s %s %-30s %-12s %3d      %s\n",
			prefix, statusSymbol, name, server.Status, server.ActiveConnections, server.URL))
	}

	// Add key bindings information
	sb.WriteString("\nControls: ↑↓=Navigate | Enter=Details | S=Start/Stop | R=Refresh | C=Config | Q=Quit\n")
	return sb.String()
}

func (m *AppModel) loadMockServers() {
	// Mock data for initial implementation
	m.State.Servers = []MCPServer{
		{
			ID:                "1",
			Name:              "GitHub MCP Server",
			URL:               "stdio://npx @modelcontextprotocol/server-github",
			Status:            "running",
			StartTime:         time.Now().Add(-30 * time.Minute),
			ResponseTime:      120 * time.Millisecond,
			ActiveConnections: 2,
			Description:       "GitHub integration server",
			Tools:             []string{"get_repo_info", "create_issue", "search_issues"},
		},
		{
			ID:                "2",
			Name:              "Calendar MCP Server", 
			URL:               "http://localhost:8000/sse",
			Status:            "running",
			StartTime:         time.Now().Add(-2 * time.Hour),
			ResponseTime:      85 * time.Millisecond,
			ActiveConnections: 1,
			Description:       "Personal calendar integration",
			Tools:             []string{"get_events", "create_event", "update_event"},
		},
		{
			ID:                "3",
			Name:              "File System MCP",
			URL:               "stdio://python filesystem_server.py",
			Status:            "stopped",
			StartTime:         time.Time{},
			ResponseTime:      0,
			ActiveConnections: 0,
			Description:       "File system operations",
			Tools:             []string{"read_file", "write_file", "list_dir"},
		},
	}
}

func (m *AppModel) listView() string {
	// Enhanced list view implementation
	s := "MCOP - Model Context Protocol Operations Monitor\n\n"
	s += fmt.Sprintf("%-4s %-30s %-12s %-10s %s\n", "ID", "NAME", "STATUS", "CONNS", "URL")
	s += fmt.Sprintf("%-4s %-30s %-12s %-10s %s\n", "--", "----", "------", "-----", "---")

	for i, server := range m.State.Servers {
		prefix := "  "
		if i == m.State.SelectedIndex {
			prefix = " >"
		}

		statusSymbol := "●" // running
		if server.Status == "stopped" {
			statusSymbol = "○" // stopped
		} else if server.Status == "error" {
			statusSymbol = "●" // error (red)
		}

		// Add color indicators by using symbols or other characters
		name := server.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		s += fmt.Sprintf("%s %s %-30s %-12s %3d      %s\n",
			prefix, statusSymbol, name, server.Status, server.ActiveConnections, server.URL)
	}

	// Add key bindings information
	s += "\nControls: ↑↓=Navigate | Enter=Details | S=Start/Stop | R=Refresh | C=Config | Q=Quit\n"
	return s
}

func (m *AppModel) detailView() string {
	if m.State.SelectedIndex >= len(m.State.Servers) {
		return "No server selected"
	}

	server := m.State.Servers[m.State.SelectedIndex]

	s := "MCOP - Server Details\n\n"
	s += "Name: " + server.Name + "\n"
	s += "URL: " + server.URL + "\n"
	s += "Status: " + server.Status + "\n"

	if !server.StartTime.IsZero() {
		s += "Start Time: " + server.StartTime.Format("2006-01-02 15:04:05") + "\n"
	} else {
		s += "Start Time: Not started\n"
	}

	s += "Response Time: " + server.ResponseTime.String() + "\n"
	s += "Active Connections: " + fmt.Sprintf("%d", server.ActiveConnections) + "\n"
	s += "Description: " + server.Description + "\n"
	s += "\nAvailable Tools:\n"
	for _, tool := range server.Tools {
		s += "  - " + tool + "\n"
	}

	// Add start/stop button based on current status
	action := "start"
	if server.Status == "running" {
		action = "stop"
	}

	s += fmt.Sprintf("\nPress 'esc' to return, 's' to %s, 'd' to disconnect\n", action)
	return s
}

func (m *AppModel) configView() string {
	s := "MCOP - Configuration\n\n"
	s += "Auto-refresh: " + boolToString(m.State.AutoRefresh) + "\n"
	s += "Refresh Rate: " + fmt.Sprintf("%ds", m.State.RefreshRate) + "\n"
	s += "\nPress 'esc' to return\n"
	return s
}

func boolToString(b bool) string {
	if b {
		return "enabled"
	}
	return "disabled"
}

func (m *AppModel) ToggleServer(index int) {
	if index >= len(m.State.Servers) {
		return
	}

	server := &m.State.Servers[index]

	if server.Status == "running" {
		// Disconnect from the MCP server
		client, exists := m.State.MCPConnections[server.ID]
		if exists && client != nil {
			client.Disconnect()
			delete(m.State.MCPConnections, server.ID)
		}
		server.Status = "stopped"
	} else if server.Status == "stopped" {
		// Connect to the MCP server
		client := mcp.NewMCPClient(*server)  // Pass value, not pointer
		err := client.Connect()
		if err != nil {
			server.Status = "error"
			return
		}
		m.State.MCPConnections[server.ID] = client
		server.Status = "running"
		server.StartTime = time.Now()
	}

	// Update active connections based on status
	if server.Status == "running" {
		server.ActiveConnections = 1 // Simulate one active connection when running
	} else {
		server.ActiveConnections = 0
	}
}

func (m *AppModel) DisconnectServer(index int) {
	if index >= len(m.State.Servers) {
		return
	}

	server := &m.State.Servers[index]
	if server.Status == "running" {
		client, exists := m.State.MCPConnections[server.ID]
		if exists && client != nil {
			client.Disconnect()
			delete(m.State.MCPConnections, server.ID)
		}
		server.ActiveConnections = 0
		server.Status = "stopped"
	}
}