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
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		var sf lexicon.SchemaFile
		if err := json.Unmarshal(raw, &sf); err != nil {
			return fmt.Errorf("%s: parse: %w", f, err)
		}
		if err := sf.FinishParse(); err != nil {
			return fmt.Errorf("%s: FinishParse: %w", f, err)
		}
		if err := sf.CheckSchema(); err != nil {
			return fmt.Errorf("%s: CheckSchema: %w", f, err)
		}
		if sf.Lexicon != 1 {
			return fmt.Errorf("%s: lexicon = %d, want 1", f, sf.Lexicon)
		}
	}
	return nil
}
