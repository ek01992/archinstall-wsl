package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"archinstall-wsl/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	version = flag.Bool("version", false, "Show version information")
	help    = flag.Bool("help", false, "Show help information")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("ArchInstall WSL TUI Configurator v1.0.0")
		os.Exit(0)
	}

	if *help {
		fmt.Println("ArchInstall WSL TUI Configurator")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Initialize the TUI application
	program := tea.NewProgram(
		tui.NewModel(),
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the TUI application
	if _, err := program.Run(); err != nil {
		log.Printf("Error running TUI application: %v", err)
		os.Exit(1)
	}
}
