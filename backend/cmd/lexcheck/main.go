package main

import (
	"fmt"
	"os"

	"github.com/otherwisejunk/protocast/internal/lexcheck"
	"github.com/otherwisejunk/protocast/internal/nsid"
)

// The lexicon tree sits at the repo root, outside this module. Default is
// relative to backend/, where the Makefile invokes this; pass a root to override.
const defaultRoot = "../lexicons"

func main() {
	root := defaultRoot
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	if err := lexcheck.CheckSchemaDir(nsid.SchemaDir(root)); err != nil {
		fmt.Fprintln(os.Stderr, "lexcheck:", err)
		os.Exit(1)
	}
	fmt.Println("schemas OK")
}
