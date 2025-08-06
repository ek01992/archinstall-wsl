package main

import (
	"flag"
	"fmt"
	"os"
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

	fmt.Println("ArchInstall WSL TUI Configurator")
	fmt.Println("Starting application...")
	os.Exit(0)
}
