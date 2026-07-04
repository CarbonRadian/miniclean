// Command miniclean cleans structured data in CSV and JSON files.
package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "miniclean: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Println("miniclean", version)
		return nil
	}
	return fmt.Errorf("not implemented yet")
}
