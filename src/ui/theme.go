package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	"github.com/charmbracelet/lipgloss"
)

// VSCodeTheme represents a VSCode theme structure
type VSCodeTheme struct {
	Name string `json:"name"`
	Colors struct {
		Foreground string `json:"foreground"`
		Background string `json:"background"`
		EditorBackground string `json:"editor.background"`
		EditorForeground string `json:"editor.foreground"`
		StatusBarBackground string `json:"statusBar.background"`
		StatusBarForeground string `json:"statusBar.foreground"`
		TabActiveForeground string `json:"tab.activeForeground"`
		TabInactiveBackground string `json:"tab.inactiveBackground"`
		ListActiveSelectionBackground string `json:"list.activeSelectionBackground"`
		ListInactiveSelectionBackground string `json:"list.inactiveSelectionBackground"`
	} `json:"colors"`
}

// loadVSCodeTheme tries to load the current VSCode theme for styling consistency
func loadVSCodeTheme() *VSCodeTheme {
	var configPath string

	switch runtime.GOOS {
	case "windows":
		configPath = filepath.Join(os.Getenv("USERPROFILE"), ".vscode", "User", "settings.json")
	case "darwin":
		configPath = filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Code", "User", "settings.json")
	case "linux":
		configPath = filepath.Join(os.Getenv("HOME"), ".config", "Code", "User", "settings.json")
	default:
		return nil
	}

	// Attempt to read the settings file
	data, err := os.ReadFile(configPath)
	if err == nil {
		var theme VSCodeTheme
		if json.Unmarshal(data, &theme) == nil {
			return &theme
		}
	}

	return nil
}

// GetThemeColors returns theme colors based on system settings (VSCode or default)
func GetThemeColors() map[string]string {
	theme := loadVSCodeTheme()

	colors := make(map[string]string)

	if theme != nil {
		// Use VSCode theme colors if available
		if theme.Colors.Background != "" {
			colors["background"] = theme.Colors.Background
		}
		if theme.Colors.Foreground != "" {
			colors["foreground"] = theme.Colors.Foreground
		}
		if theme.Colors.EditorBackground != "" {
			colors["editorBackground"] = theme.Colors.EditorBackground
		}
		if theme.Colors.EditorForeground != "" {
			colors["editorForeground"] = theme.Colors.EditorForeground
		}
		if theme.Colors.ListActiveSelectionBackground != "" {
			colors["selectionBackground"] = theme.Colors.ListActiveSelectionBackground
		}
		if theme.Colors.StatusBarBackground != "" {
			colors["statusBarBackground"] = theme.Colors.StatusBarBackground
		}
	}

	// Fallback to default colors if theme info is not available
	defaults := map[string]string{
		"background":           "235",
		"foreground":           "252",
		"editorBackground":     "235",
		"editorForeground":     "252",
		"selectionBackground":  "62",
		"statusBarBackground":  "240",
		"titleBackground":      "57",
		"titleForeground":      "212",
		"runningStatus":        "46",
		"stoppedStatus":        "203",
		"errorStatus":          "196",
		"headerBackground":     "235",
		"headerForeground":     "246",
	}

	// Merge defaults with theme colors
	for key, value := range defaults {
		if _, exists := colors[key]; !exists {
			colors[key] = value
		}
	}

	return colors
}

// ApplyThemeToStyle updates a lipgloss style based on theme colors
func ApplyThemeToStyle(style lipgloss.Style, themeType string) lipgloss.Style {
	colors := GetThemeColors()

	switch themeType {
	case "title":
		return style.Foreground(lipgloss.Color(colors["titleForeground"])).Background(lipgloss.Color(colors["titleBackground"]))
	case "header":
		return style.Foreground(lipgloss.Color(colors["headerForeground"])).Background(lipgloss.Color(colors["headerBackground"]))
	case "selected":
		return style.Foreground(lipgloss.Color(colors["foreground"])).Background(lipgloss.Color(colors["selectionBackground"]))
	case "running":
		return style.Foreground(lipgloss.Color(colors["runningStatus"]))
	case "stopped":
		return style.Foreground(lipgloss.Color(colors["stoppedStatus"]))
	case "error":
		return style.Foreground(lipgloss.Color(colors["errorStatus"]))
	default:
		return style.Foreground(lipgloss.Color(colors["foreground"]))
	}
}