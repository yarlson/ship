package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"ship/cli"
	"ship/workflow"
)

func main() {
	cfg, err := cli.Parse(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Print(cli.HelpText)
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	if err := workflow.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
