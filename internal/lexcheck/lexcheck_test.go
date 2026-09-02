package lexcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckSchemaDir(t *testing.T) {
	validSchema := `{
		"lexicon": 1,
		"id": "com.example.valid",
		"defs": {
			"main": {
				"type": "record",
				"key": "tid",
				"record": {
					"type": "object",
					"properties": {
						"createdAt": {"type": "string", "format": "datetime"}
					}
				}
			}
		}
	}`

	badJSON := `{`

	badLexicon := `{
		"lexicon": 2,
		"id": "com.example.badlexicon",
		"defs": {
			"main": {
				"type": "record",
				"key": "tid",
				"record": {"type": "object"}
			}
		}
	}`

	badNSID := `{
		"lexicon": 1,
		"id": "not a valid nsid",
		"defs": {
			"main": {
				"type": "record",
				"key": "tid",
				"record": {"type": "object"}
			}
		}
	}`

	badDefName := `{
		"lexicon": 1,
		"id": "com.example.badname",
		"defs": {
			"bad#name": {
				"type": "string"
			}
		}
	}`

	tests := []struct {
		name      string
		files     map[string]string
		wantErr   string
		wantNoDir bool
	}{
		{
			name: "valid schema directory",
			files: map[string]string{
				"schema.json": validSchema,
			},
		},
		{
			name:      "empty directory",
			wantNoDir: true,
			wantErr:   "no schema files",
		},
		{
			name: "invalid json",
			files: map[string]string{
				"broken.json": badJSON,
			},
			wantErr: "parse",
		},
		{
			name: "unsupported lexicon version",
			files: map[string]string{
				"bad-lexicon.json": badLexicon,
			},
			wantErr: "unsupported lexicon language version",
		},
		{
			name: "invalid schema nsid",
			files: map[string]string{
				"invalid-nsid.json": badNSID,
			},
			wantErr: "invalid lexicon schema NSID",
		},
		{
			name: "invalid schema definition name",
			files: map[string]string{
				"invalid-def.json": badDefName,
			},
			wantErr: "schema name invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			if tc.wantNoDir {
				err := CheckSchemaDir(dir)
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("CheckSchemaDir() error = %v, want substring %q", err, tc.wantErr)
				}
				return
			}

			// Write test files to the temporary directory
			for name, content := range tc.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
					t.Fatalf("WriteFile(%s): %v", name, err)
				}
			}

			err := CheckSchemaDir(dir)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckSchemaDir() unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("CheckSchemaDir() error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}
