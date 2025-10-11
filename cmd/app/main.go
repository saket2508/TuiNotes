package main

import (
	"fmt"
	"os"
	"path/filepath"

	"markdown-note-taking-app/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Use a local database file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(homeDir, ".markdown-notes.db")

	// Create the app
	app, err := ui.NewApp(dbPath)
	if err != nil {
		fmt.Printf("Error creating app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	// Run the program
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
