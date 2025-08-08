package main

import (
	"fmt"
	"os"

	"archwsl-tui-configurator/internal/app"
	"archwsl-tui-configurator/pkg/version"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	_ = version.Version
}
