package feedxml

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html/charset"
)

const (
	bomUTF8    = "\xef\xbb\xbf"
	bomUTF16BE = "\xfe\xff"
	bomUTF16LE = "\xff\xfe"
)

var xmlDeclEncoding = regexp.MustCompile(`^\s*<\?xml\s[^>]*?encoding\s*=\s*["']([^"']+)["']`)

// declaredEncoding returns the encoding a document names for itself, by byte
// order mark or XML declaration, or "" for neither.
func declaredEncoding(raw []byte) string {
	switch {
	case bytes.HasPrefix(raw, []byte(bomUTF8)):
		return "utf-8"
	case bytes.HasPrefix(raw, []byte(bomUTF16BE)):
		return "utf-16be"
	case bytes.HasPrefix(raw, []byte(bomUTF16LE)):
		return "utf-16le"
	}
	if match := xmlDeclEncoding.FindSubmatch(raw); match != nil {
		return string(match[1])
	}
	return ""
}

func isUTF8Label(label string) bool {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "", "utf-8", "utf8", "unicode-1-1-utf-8", "us-ascii", "ascii":
		return true
	}
	return false
}

// decodeUTF8 converts a document to UTF-8 using the encoding it declares.
func decodeUTF8(raw []byte) []byte {
	label := declaredEncoding(raw)
	if isUTF8Label(label) {
		return bytes.TrimPrefix(raw, []byte(bomUTF8))
	}

	converted, err := charset.NewReaderLabel(label, bytes.NewReader(raw))
	if err != nil {
		return raw
	}
	decoded, err := io.ReadAll(converted)
	if err != nil {
		return raw
	}
	// A UTF-16 byte order mark survives conversion as U+FEFF.
	return bytes.TrimPrefix(decoded, []byte("\ufeff"))
}
