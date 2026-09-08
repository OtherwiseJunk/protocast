package feedxml

import (
	"bytes"
	"encoding/binary"
	"slices"
	"testing"
	"unicode/utf16"
)

const (
	showTitle    = "日本のポッドキャスト — Café"
	episodeTitle = "エピソード1"
	feedTail     = `</item></channel></rss>`
)

// multibyteFeed is left open at <item> so cases can append another one.
const multibyteFeed = `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel>` +
	`<title>` + showTitle + `</title><item><title>` + episodeTitle

// utf16Doc encodes a document as UTF-16 behind a byte order mark.
func utf16Doc(source string, order binary.ByteOrder) []byte {
	units := utf16.Encode([]rune("\ufeff" + source))
	out := make([]byte, len(units)*2)
	for index, unit := range units {
		order.PutUint16(out[index*2:], unit)
	}
	return out
}

// latin1Doc builds a document whose non-ASCII bytes are Latin-1 rather than
// UTF-8, so nothing but its declaration can decode it.
func latin1Doc() []byte {
	doc := []byte(`<?xml version="1.0" encoding="ISO-8859-1"?><rss><channel><title>Caf`)
	doc = append(doc, 0xE9) // é
	doc = append(doc, `</title><item><title>Episodio N`...)
	doc = append(doc, 0xFA) // ú
	doc = append(doc, `m 1`+feedTail...)
	return doc
}

// Charset reader must handle non-UTF-8 encodings. ISO-8859-1 (Latin-1) is
// common in older feeds. If the charset reader is not wired correctly, non-ASCII
// bytes are silently lost or mangled, which is a parser bug. This test asserts
// that the charset reader is active and working by parsing a document declared
// as ISO-8859-1 and containing bytes that are invalid UTF-8 on their own.
func TestNonUTF8EncodingDecoded(test *testing.T) {
	// XML document with ISO-8859-1 encoding declaration, containing the byte 0xE9
	// which is 'é' in ISO-8859-1 but is not valid UTF-8 on its own.
	doc := []byte{
		// <?xml version="1.0" encoding="ISO-8859-1"?>
		0x3C, 0x3F, 0x78, 0x6D, 0x6C, 0x20, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6F,
		0x6E, 0x3D, 0x22, 0x31, 0x2E, 0x30, 0x22, 0x20, 0x65, 0x6E, 0x63, 0x6F,
		0x64, 0x69, 0x6E, 0x67, 0x3D, 0x22, 0x49, 0x53, 0x4F, 0x2D, 0x38, 0x38,
		0x35, 0x39, 0x2D, 0x31, 0x22, 0x3F, 0x3E,
		// <rss><channel><title>Caf
		0x3C, 0x72, 0x73, 0x73, 0x3E, 0x3C, 0x63, 0x68, 0x61, 0x6E, 0x6E, 0x65,
		0x6C, 0x3E, 0x3C, 0x74, 0x69, 0x74, 0x6C, 0x65, 0x3E, 0x43, 0x61, 0x66,
		// é (0xE9 in ISO-8859-1)
		0xE9,
		// </title></channel></rss>
		0x3C, 0x2F, 0x74, 0x69, 0x74, 0x6C, 0x65, 0x3E, 0x3C, 0x2F, 0x63, 0x68,
		0x61, 0x6E, 0x6E, 0x65, 0x6C, 0x3E, 0x3C, 0x2F, 0x72, 0x73, 0x73, 0x3E,
	}

	root, err := Parse(bytes.NewReader(doc))
	if err != nil {
		test.Fatalf("Parse: %v", err)
	}
	got := root.Child(NSRSS, "channel").ChildText(NSRSS, "title")
	if got != "Café" {
		test.Errorf("non-UTF-8 text = %q, want %q", got, "Café")
	}
}

// A document's own encoding declaration is authoritative. Sniffing the bytes
// instead falls back to windows-1252 on any UTF-8 failure, turning one corrupt
// byte into mojibake across the whole feed.
func TestEncodingDeclarationIsAuthoritative(test *testing.T) {
	const utf16Head = `<?xml version="1.0" encoding="UTF-16"?><rss><channel><title>` +
		showTitle + `</title><item><title>` + episodeTitle + feedTail

	tests := []struct {
		name      string
		doc       []byte
		wantTitle string
		wantItems []string
	}{
		{"UTF-8 as declared", []byte(multibyteFeed + feedTail), showTitle, []string{episodeTitle}},
		{"UTF-8 with one stray Latin-1 byte",
			[]byte(multibyteFeed + "</item><item><title>bad \xe9 byte" + feedTail),
			showTitle, []string{episodeTitle, "bad \ufffd byte"}},
		{"UTF-8 behind a byte order mark",
			[]byte("\xef\xbb\xbf" + multibyteFeed + feedTail), showTitle, []string{episodeTitle}},
		{"no declaration at all, the XML default",
			[]byte(`<rss><channel><title>` + showTitle + `</title><item><title>` +
				episodeTitle + feedTail), showTitle, []string{episodeTitle}},
		{"declared ISO-8859-1", latin1Doc(), "Café", []string{"Episodio Núm 1"}},
		{"UTF-16 big endian", utf16Doc(utf16Head, binary.BigEndian),
			showTitle, []string{episodeTitle}},
		{"UTF-16 little endian", utf16Doc(utf16Head, binary.LittleEndian),
			showTitle, []string{episodeTitle}},
		// An encoding no charset table knows is not a reason to discard a show.
		{"encoding label no table knows",
			[]byte(`<?xml version="1.0" encoding="x-mac-croatian"?><rss><channel>` +
				`<title>Probe</title><item><title>Kept` + feedTail),
			"Probe", []string{"Kept"}},
	}

	for _, testCase := range tests {
		test.Run(testCase.name, func(test *testing.T) {
			root, err := ParseBytes(testCase.doc)
			if err != nil {
				test.Fatalf("ParseBytes: %v", err)
			}
			channel, items, err := Locate(root)
			if err != nil {
				test.Fatalf("Locate: %v", err)
			}
			if got := channel.ChildText(AnyNS, "title"); got != testCase.wantTitle {
				test.Errorf("channel title = %q, want %q", got, testCase.wantTitle)
			}
			var got []string
			for _, item := range items {
				got = append(got, item.ChildText(AnyNS, "title"))
			}
			if !slices.Equal(got, testCase.wantItems) {
				test.Errorf("item titles = %q, want %q", got, testCase.wantItems)
			}
		})
	}
}
