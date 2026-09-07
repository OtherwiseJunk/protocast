// nsid is a package containing constants defining the NSID (Namespace Identifier) for various components in the system.
// It provides a centralized location for managing and referencing NSIDs, ensuring consistency across the codebase.
package nsid

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
