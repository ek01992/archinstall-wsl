package main

import (
	"flag"
	"fmt"
	"os"

	"archwsl-tui-configurator/internal/app"
	"archwsl-tui-configurator/pkg/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(version.Version)
		return
	}

	application := app.New()
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
