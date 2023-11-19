package main

import (
	"os"
	"sshtools/internal/rexec"
)

func main() {
	cmd := rexec.NewRExecCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
