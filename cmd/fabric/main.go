package main

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"

	"github.com/danielmiessler/fabric/internal/cli"
)

var version = "dev" // This would normally be set at build time

func main() {
	err := cli.Cli(version)
	if err != nil && !flags.WroteHelp(err) {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
}
