package main

import (
	"fmt"
	"os"

	"notion-cli/internal/cli"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCodeFromError(err))
	}
}
