package main

import (
	"fmt"
	"os"

	"github.com/otherwisejunk/protocast/internal/lexcheck"
)

func main() {
	const dir = "lexicons/com/cacheblasters/protocast"
	if err := lexcheck.CheckSchemaDir(dir); err != nil {
		fmt.Fprintln(os.Stderr, "lexcheck:", err)
		os.Exit(1)
	}
	fmt.Println("schemas OK")
}
