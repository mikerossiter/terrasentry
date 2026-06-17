// Command terrasentry is the CLI entrypoint. It delegates all argument
// handling to the cmd package so the binary stays a thin wrapper.
package main

import (
	"os"

	"github.com/mikejrossiter/terrasentry/cmd"
)

func main() {
	os.Exit(cmd.Main(os.Args[1:]))
}
