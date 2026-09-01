// wraps indigo's lexicon catalog for schema self-checking
// and record validation.
package lexcheck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bluesky-social/indigo/atproto/lexicon"
)

// CheckSchemaDir parses every .json file in dir as a Lexicon schema file and
// runs indigo's own structural checks over it.
func CheckSchemaDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no schema files in %s", dir)
	}
	for _, file := range files {
		raw, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		var schema lexicon.SchemaFile
		if err := json.Unmarshal(raw, &schema); err != nil {
			return fmt.Errorf("%s: parse: %w", file, err)
		}
		if err := schema.FinishParse(); err != nil {
			return fmt.Errorf("%s: FinishParse: %w", file, err)
		}
		if err := schema.CheckSchema(); err != nil {
			return fmt.Errorf("%s: CheckSchema: %w", file, err)
		}
		if schema.Lexicon != 1 {
			return fmt.Errorf("%s: lexicon = %d, want 1", file, schema.Lexicon)
		}
	}
	return nil
}
