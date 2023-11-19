package main

import (
	"os"
	"sshtools/internal/rcp"
)

func main() {
	cmd := rcp.NewRCopyCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
