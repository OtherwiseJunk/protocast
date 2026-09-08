package feedxml

import (
	"strconv"
	"unicode/utf8"
)

const replacement = "�"

// sanitizeXML rewrites the byte sequences that make encoding/xml abandon a
// document into ones it tolerates: invalid UTF-8, characters XML forbids, and
// stray markup delimiters in text. Feeds ship all three, and without this one
// bad byte in the last item throws away every item before it.
//
// src must already be UTF-8. The rewrite is value-local: a broken character
// degrades to U+FFFD and the surrounding elements parse normally.
func sanitizeXML(src []byte) []byte {
	out := make([]byte, 0, len(src))
	for i := 0; i < len(src); {
		switch src[i] {
		case '<':
			if end, ok := verbatimSpan(src, i); ok {
				out = append(out, src[i:end]...)
				i = end
				continue
			}
			if startsMarkup(src, i) {
				out = append(out, '<')
				i++
				continue
			}
			out = append(out, "&lt;"...)
			i++

		case '&':
			end, illegal, ok := scanReference(src, i)
			switch {
			case !ok:
				out = append(out, "&amp;"...)
				i++
			case illegal:
				out = append(out, replacement...)
				i = end
			default:
				out = append(out, src[i:end]...)
				i = end
			}

		default:
			char, size := utf8.DecodeRune(src[i:])
			switch {
			case char == utf8.RuneError && size == 1:
				out = append(out, replacement...)
				i++
			case !legalXMLRune(char):
				i += size
			default:
				out = append(out, src[i:i+size]...)
				i += size
			}
		}
	}
	return out
}

// legalXMLRune reports whether a rune may appear in an XML 1.0 document.
func legalXMLRune(candidate rune) bool {
	switch {
	case candidate == 0x09, candidate == 0x0A, candidate == 0x0D:
		return true
	case candidate >= 0x20 && candidate <= 0xD7FF:
		return true
	case candidate >= 0xE000 && candidate <= 0xFFFD:
		return true
	case candidate >= 0x10000 && candidate <= 0x10FFFF:
		return true
	}
	return false
}

// verbatimSpan returns the end of a CDATA section or comment starting at
// src[start]. Their contents are literal, so escaping inside them would corrupt
// show notes rather than repair them.
func verbatimSpan(src []byte, start int) (int, bool) {
	for open, close := range map[string]string{"<![CDATA[": "]]>", "<!--": "-->"} {
		if !hasPrefixAt(src, start, open) {
			continue
		}
		for scan := start + len(open); scan+len(close) <= len(src); scan++ {
			if hasPrefixAt(src, scan, close) {
				return scan + len(close), true
			}
		}
		return len(src), true // unterminated: pass the rest through untouched
	}
	return 0, false
}

func hasPrefixAt(src []byte, at int, prefix string) bool {
	return at+len(prefix) <= len(src) && string(src[at:at+len(prefix)]) == prefix
}

// startsMarkup reports whether the '<' at src[start] opens a tag, a
// declaration or a processing instruction rather than being literal text.
func startsMarkup(src []byte, start int) bool {
	if start+1 >= len(src) {
		return false
	}
	switch next := src[start+1]; {
	case next == '/', next == '!', next == '?', next == '_', next == ':':
		return true
	case next >= 'a' && next <= 'z', next >= 'A' && next <= 'Z':
		return true
	case next >= utf8.RuneSelf: // a non-ASCII name start
		return true
	}
	return false
}

// scanReference examines the reference beginning at src[start] ('&'). ok is
// false when it is a bare ampersand rather than a well-formed reference;
// illegal is true for a numeric reference to a codepoint XML forbids.
func scanReference(src []byte, start int) (end int, illegal, ok bool) {
	scan := start + 1
	if scan < len(src) && src[scan] == '#' {
		scan++
		base := 10
		if scan < len(src) && (src[scan] == 'x' || src[scan] == 'X') {
			base = 16
			scan++
		}
		digits := scan
		for scan < len(src) && isDigitIn(src[scan], base) {
			scan++
		}
		if scan == digits || scan >= len(src) || src[scan] != ';' {
			return 0, false, false
		}
		code, err := strconv.ParseInt(string(src[digits:scan]), base, 32)
		return scan + 1, err != nil || !legalXMLRune(rune(code)), true
	}

	name := scan
	for scan < len(src) && isNameByte(src[scan]) {
		scan++
	}
	if scan == name || scan >= len(src) || src[scan] != ';' {
		return 0, false, false
	}
	return scan + 1, false, true
}

func isDigitIn(candidate byte, base int) bool {
	switch {
	case candidate >= '0' && candidate <= '9':
		return true
	case base == 16 && (candidate|0x20) >= 'a' && (candidate|0x20) <= 'f':
		return true
	}
	return false
}

func isNameByte(candidate byte) bool {
	switch {
	case candidate >= 'a' && candidate <= 'z', candidate >= 'A' && candidate <= 'Z':
		return true
	case candidate >= '0' && candidate <= '9':
		return true
	case candidate == '.', candidate == '-', candidate == '_', candidate == ':':
		return true
	case candidate >= utf8.RuneSelf:
		return true
	}
	return false
}
