package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"mcop/src/model"
)

// AppInterface combines model and styling functionality
type AppInterface struct {
	AppModel    *model.AppModel
	Width       int
	Height      int
	// Dialog state
	ShowDialog    bool
	DialogType    string
	DialogMessage string
	// Log state
	LogMessages   []string
}

// Styled components - using lipgloss for theming
var (
	// Base window style
	BaseStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1).
		Margin(1)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Background(lipgloss.Color("57")).
		Padding(0, 1).
		MarginBottom(1).
		Bold(true)

	// Header styles for table headers
	HeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Underline(true)

	// Selected item style
	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("220")).
		Background(lipgloss.Color("62")).
		PaddingLeft(1).
		Bold(true)

	// Regular item style
	ItemStyle = lipgloss.NewStyle().
		PaddingLeft(2)

	// Status running style
	StatusRunningStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")). // Green
		Padding(0, 1)

	// Status stopped style
	StatusStoppedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("203")). // Red
		Padding(0, 1)

	// Status error style
	StatusErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")). // Bright red
		Padding(0, 1)

	// Help text style
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		MarginTop(1)

	// Detail view styles
	DetailTitleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")).
		Bold(true).
		Underline(true).
		MarginBottom(1)

	DetailValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")).
		MarginLeft(2)

	// Dialog styles
	DialogStyle = lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("220")).
		Padding(2).
		Background(lipgloss.Color("235"))

	// Log styles
	LogStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Height(8).
		Padding(1).
		MarginTop(1)

	// Status bar style
	StatusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("255")).
		Padding(0, 1).
		MarginTop(1)
)

// NewAppModel creates a new instance of the styled application model
func NewAppModel() *AppInterface {
	return &AppInterface{
		AppModel:    model.NewAppModel(),
		Width:       80,
		Height:      24,
		LogMessages: []string{},
	}
}

// SetInitialServerURL sets the initial server URL
func (a *AppInterface) SetInitialServerURL(url string) {
	a.AppModel.SetInitialServerURL(url)
}

// View returns the styled view of the application with full layout
func (a *AppInterface) View() string {
	var content string

	// Render main content based on view
	switch a.AppModel.State.View {
	case "detail":
		content = a.renderServerDetail()
	case "config":
		content = a.renderConfigView()
	default:
		content = a.renderServerList()
	}

	// Add log console to content
	logContent := a.renderLogConsole()
	content += "\n" + logContent

	// Add status bar
	statusBar := a.renderStatusBar()
	content += "\n" + statusBar

	// Wrap in base style
	finalContent := BaseStyle.Render(content)

	// If dialog is visible, overlay it
	if a.ShowDialog {
		dialog := a.renderDialog()
		return a.overlayDialog(finalContent, dialog)
	}

	return finalContent
}

// renderStatusBar renders the status bar
func (a *AppInterface) renderStatusBar() string {
	statusText := fmt.Sprintf("MCOP | Servers: %d | View: %s | Press 'H' for Help | Q: Quit",
		len(a.AppModel.State.Servers), a.AppModel.State.View)
	return StatusBarStyle.Render(statusText)
}

// renderLogConsole renders the log console
func (a *AppInterface) renderLogConsole() string {
	logContent := "Operation Logs:\n"

	// Show last 5 log messages
	startIdx := 0
	if len(a.LogMessages) > 5 {
		startIdx = len(a.LogMessages) - 5
	}

	for i := startIdx; i < len(a.LogMessages); i++ {
		logContent += a.LogMessages[i] + "\n"
	}

	// If no logs, show a placeholder
	if len(a.LogMessages) == 0 {
		logContent += "[No recent logs]"
	}

	return LogStyle.Render(logContent)
}

// renderDialog renders the dialog box
func (a *AppInterface) renderDialog() string {
	dialog := a.DialogMessage
	if a.DialogType == "help" {
		dialog = a.DialogMessage
	}
	return DialogStyle.Render(dialog)
}

// overlayDialog overlays a dialog on top of the main content
func (a *AppInterface) overlayDialog(content, dialog string) string {
	// Calculate center position for dialog
	contentLines := strings.Split(content, "\n")
	contentHeight := len(contentLines)

	dialogLines := strings.Split(dialog, "\n")
	dialogHeight := len(dialogLines)

	dialogPos := (contentHeight - dialogHeight) / 2
	if dialogPos < 0 {
		dialogPos = 0
	}

	// Create overlay with centered dialog
	var result []string
	for i, line := range contentLines {
		if i >= dialogPos && i < dialogPos+len(dialogLines) {
			dialogLineIdx := i - dialogPos
			if dialogLineIdx < len(dialogLines) {
				// Center the dialog line
				dialogLine := dialogLines[dialogLineIdx]
				padding := (len(line) - len(stripAnsi(dialogLine))) / 2
				if padding < 0 {
					padding = 0
				}
				result = append(result, line)
			} else {
				result = append(result, line)
			}
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n") + "\n" + dialog
}

// stripAnsi removes ANSI color codes from a string (simplified version)
func stripAnsi(s string) string {
	return s
}

// renderServerList renders the server list with styling
func (a *AppInterface) renderServerList() string {
	var sb strings.Builder

	// Render title
	title := TitleStyle.Render("MCOP - Model Context Protocol Operations Monitor")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Render table header
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		lipgloss.NewStyle().Width(4).Padding(0).Render("ID"),
		lipgloss.NewStyle().Width(30).Padding(0).Render("NAME"),
		lipgloss.NewStyle().Width(12).Padding(0).Render("STATUS"),
		lipgloss.NewStyle().Width(8).Padding(0).Render("CONNS"),
		"URL",
	)
	sb.WriteString(HeaderStyle.Render(header))
	sb.WriteString("\n")

	// Render server list
	for i, server := range a.AppModel.State.Servers {
		var rowStyle lipgloss.Style
		if i == a.AppModel.State.SelectedIndex {
			rowStyle = SelectedItemStyle
		} else {
			rowStyle = ItemStyle
		}

		var statusStyle lipgloss.Style
		switch server.Status {
		case "running":
			statusStyle = StatusRunningStyle
		case "stopped":
			statusStyle = StatusStoppedStyle
		case "error":
			statusStyle = StatusErrorStyle
		default:
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
		}

		var indicator string
		if server.Status == "running" {
			indicator = "●"
		} else {
			indicator = "○"
		}

		name := server.Name
		if len(name) > 28 {
			name = name[:25] + "..."
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Width(4).Padding(0).Render(fmt.Sprintf("%d", i+1)),
			lipgloss.NewStyle().Width(30).Padding(0).Render(name),
			lipgloss.NewStyle().Width(12).Padding(0).Render(statusStyle.Render(server.Status)),
			lipgloss.NewStyle().Width(8).Padding(0).Render(fmt.Sprintf("%d", server.ActiveConnections)),
			fmt.Sprintf("%s %s", indicator, server.URL),
		)

		sb.WriteString(rowStyle.Render(row))
		sb.WriteString("\n")
	}

	// Add controls help
	help := HelpStyle.Render("↑↓=Navigate | Enter=Details | S=Start/Stop | R=Refresh | C=Config | Q=Quit")
	sb.WriteString("\n")
	sb.WriteString(help)

	return sb.String()
}

// renderServerDetail renders the server detail view with styling
func (a *AppInterface) renderServerDetail() string {
	if a.AppModel.State.SelectedIndex >= len(a.AppModel.State.Servers) {
		return "No server selected"
	}

	server := a.AppModel.State.Servers[a.AppModel.State.SelectedIndex]
	var sb strings.Builder

	// Title
	title := TitleStyle.Render("MCOP - Server Details")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Server info
	sb.WriteString(DetailTitleStyle.Render("Name:"))
	sb.WriteString("\n")
	sb.WriteString(DetailValueStyle.Render(server.Name))
	sb.WriteString("\n\n")

	sb.WriteString(DetailTitleStyle.Render("URL:"))
	sb.WriteString("\n")
	sb.WriteString(DetailValueStyle.Render(server.URL))
	sb.WriteString("\n\n")

	sb.WriteString(DetailTitleStyle.Render("Status:"))
	sb.WriteString("\n")
	var statusStyle lipgloss.Style
	switch server.Status {
	case "running":
		statusStyle = StatusRunningStyle
	case "stopped":
		statusStyle = StatusStoppedStyle
	default:
		statusStyle = StatusErrorStyle
	}
	sb.WriteString(DetailValueStyle.Render(statusStyle.Render(server.Status)))
	sb.WriteString("\n\n")

	sb.WriteString(DetailTitleStyle.Render("Start Time:"))
	sb.WriteString("\n")
	if !server.StartTime.IsZero() {
		sb.WriteString(DetailValueStyle.Render(server.StartTime.Format("2006-01-02 15:04:05")))
	} else {
		sb.WriteString(DetailValueStyle.Render("Not started"))
	}
	sb.WriteString("\n\n")

	sb.WriteString(DetailTitleStyle.Render("Response Time:"))
	sb.WriteString("\n")
	sb.WriteString(DetailValueStyle.Render(server.ResponseTime.String()))
	sb.WriteString("\n\n")

	sb.WriteString(DetailTitleStyle.Render("Active Connections:"))
	sb.WriteString("\n")
	sb.WriteString(DetailValueStyle.Render(fmt.Sprintf("%d", server.ActiveConnections)))
	sb.WriteString("\n\n")

	sb.WriteString(DetailTitleStyle.Render("Description:"))
	sb.WriteString("\n")
	sb.WriteString(DetailValueStyle.Render(server.Description))
	sb.WriteString("\n\n")

	if len(server.Tools) > 0 {
		sb.WriteString(DetailTitleStyle.Render("Available Tools:"))
		sb.WriteString("\n")
		for _, tool := range server.Tools {
			sb.WriteString(DetailValueStyle.Render(fmt.Sprintf("  - %s", tool)))
			sb.WriteString("\n")
		}
	}

	// Add action instructions
	action := "start"
	if server.Status == "running" {
		action = "stop"
	}
	help := HelpStyle.Render(fmt.Sprintf("Press 'Esc' to return, 'S' to %s, 'D' to disconnect", action))
	sb.WriteString("\n")
	sb.WriteString(help)

	return sb.String()
}

// renderConfigView renders the configuration view with styling
func (a *AppInterface) renderConfigView() string {
	var sb strings.Builder

	// Title
	title := TitleStyle.Render("MCOP - Configuration")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Config options
	sb.WriteString(DetailTitleStyle.Render("Auto-refresh:"))
	sb.WriteString("\n")
	enabledStr := "Disabled"
	if a.AppModel.State.AutoRefresh {
		enabledStr = "Enabled"
	}
	sb.WriteString(DetailValueStyle.Render(enabledStr))
	sb.WriteString("\n\n")

	sb.WriteString(DetailTitleStyle.Render("Refresh Rate:"))
	sb.WriteString("\n")
	sb.WriteString(DetailValueStyle.Render(fmt.Sprintf("%ds", a.AppModel.State.RefreshRate)))
	sb.WriteString("\n\n")

	// Help text
	help := HelpStyle.Render("Press 'Esc' to return")
	sb.WriteString("\n")
	sb.WriteString(help)

	return sb.String()
}

// Update handles updates for the application
func (a *AppInterface) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window size changes
	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		a.Width = msg.Width
		a.Height = msg.Height
	}

	// Update the underlying model for non-key messages
	// But we need to intercept key messages to handle UI-specific functionality
	if _, ok := msg.(tea.KeyMsg); !ok {
		// For non-key messages (like resize), update the model directly
		updatedModel, cmd := a.AppModel.Update(msg)
		if newModel, ok := updatedModel.(*model.AppModel); ok {
			a.AppModel = newModel
		}
		return a, cmd
	}

	// Handle key messages for dialog interaction and other actions
	if msg, ok := msg.(tea.KeyMsg); ok {
		if a.ShowDialog {
			// Handle dialog keys based on dialog type
			if a.DialogType == "download" {
				if msg.String() == "y" || msg.String() == "Y" {
					// Simulate download process
					a.addLogMessage("Starting download of MCP server...")
					a.addLogMessage("Download complete!")
					a.ShowDialog = false
				} else if msg.String() == "n" || msg.String() == "n" || msg.String() == "esc" {
					a.addLogMessage("Download cancelled")
					a.ShowDialog = false
				}
			} else if a.DialogType == "help" {
				// Any key closes help dialog
				a.ShowDialog = false
			} else {
				// Handle generic dialog keys
				if msg.String() == "y" || msg.String() == "Y" {
					// Handle yes for confirmations, etc.
					a.ShowDialog = false
					// Add log message
					a.addLogMessage("Dialog confirmed")
				} else if msg.String() == "n" || msg.String() == "n" || msg.String() == "esc" {
					// Handle no/cancel
					a.ShowDialog = false
					a.addLogMessage("Dialog cancelled")
				}
			}
		} else {
			// Handle normal keys
			switch msg.String() {
			case "h":
				// Show comprehensive help dialog
				a.ShowDialog = true
				a.DialogType = "help"
				a.DialogMessage = "MCOP - MCP Operations Monitor\n\n" +
					"Navigation:\n" +
					"  ↑/↓  - Move between servers\n" +
					"  Enter - View server details\n" +
					"  Esc   - Return to list view\n\n" +
					"Server Management:\n" +
					"  S     - Start/Stop selected server\n" +
					"  D     - Disconnect selected server\n" +
					"  C     - Configuration view\n" +
					"  R     - Refresh server list\n\n" +
					"Tools:\n" +
					"  X     - Download/Configure MCP Servers\n" +
					"  H     - Show this help\n" +
					"  Q     - Quit MCOP\n\n" +
					"Press any key to close..."
			case "s":
				// Handle start/stop for servers
				if a.AppModel.State.View == "list" && a.AppModel.State.SelectedIndex < len(a.AppModel.State.Servers) {
					server := a.AppModel.State.Servers[a.AppModel.State.SelectedIndex]
					originalStatus := server.Status
					a.AppModel.ToggleServer(a.AppModel.State.SelectedIndex)
					// Add log message about the action
					if originalStatus == "running" {
						a.addLogMessage(fmt.Sprintf("Stopped server: %s", server.Name))
					} else {
						a.addLogMessage(fmt.Sprintf("Started server: %s", server.Name))
					}
				}
			case "d":
				// Handle disconnect
				if a.AppModel.State.View == "detail" && a.AppModel.State.SelectedIndex < len(a.AppModel.State.Servers) {
					server := a.AppModel.State.Servers[a.AppModel.State.SelectedIndex]
					a.AppModel.DisconnectServer(a.AppModel.State.SelectedIndex)
					a.addLogMessage(fmt.Sprintf("Disconnected from server: %s", server.Name))
				}
			case "x":
				// Show download/configure dialog
				a.ShowDialog = true
				a.DialogType = "download"
				a.DialogMessage = "MCP Server Manager:\n\n- Download new server\n- Configure existing servers\n\n[y/N] to download example server?"
			}
		}
	}

	// Return the UI wrapper with the updated model
	return a, nil
}

// addLogMessage adds a message to the log console
func (a *AppInterface) addLogMessage(message string) {
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	a.LogMessages = append(a.LogMessages, logEntry)

	// Limit log size to prevent memory issues
	if len(a.LogMessages) > 50 {
		a.LogMessages = a.LogMessages[len(a.LogMessages)-50:]
	}
}

// addServerLog adds server operation logs
func (a *AppInterface) addServerLog(serverName, operation string) {
	message := fmt.Sprintf("Server '%s' %s", serverName, operation)
	a.addLogMessage(message)
}

// Init initializes the application
func (a *AppInterface) Init() tea.Cmd {
	return a.AppModel.Init()
}