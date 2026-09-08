// nsid is a package containing constants defining the NSID (Namespace Identifier) for various components in the system.
// It provides a centralized location for managing and referencing NSIDs, ensuring consistency across the codebase.
package nsid

import (
	"path/filepath"
	"strings"
)

const Prefix = "com.cacheblasters.protocast"

func Ref(name string) string {
	if name == "" {
		return ""
	}

	return Prefix + "." + name
}

// every schema file NSID, including non-record files.
func SchemaFiles() []string {
	return []string{Ref("defs"), Ref("show"), Ref("episode")}
}

// returns only NSIDs that define a record.
func RecordTypes() []string {
	return []string{Ref("show"), Ref("episode")}
}

// SchemaDir returns the directory holding this namespace's schema files, given
// the root of the lexicon tree. Lexicon layout mirrors the NSID, so the path is
// derived from Prefix rather than spelled a second time.
func SchemaDir(root string) string {
	return filepath.Join(root, filepath.FromSlash(strings.ReplaceAll(Prefix, ".", "/")))
}
