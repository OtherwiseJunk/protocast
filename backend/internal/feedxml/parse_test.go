package feedxml

import (
	"bytes"
	"strings"
	"testing"
)

const miniRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"
     xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd"
     xmlns:podcast="https://podcastindex.org/namespace/1.0"
     xmlns:content="http://purl.org/rss/1.0/modules/content/">
  <channel>
    <title>TV &amp; Film Weekly</title>
    <link>https://tvfilm.example</link>
    <itunes:title>Ignored Alternate Title</itunes:title>
    <podcast:guid>917393e3-1b1e-5cef-ace4-edaa54e1f810</podcast:guid>
    <item>
      <title>Ep 1</title>
      <link>https://tvfilm.example/1</link>
      <enclosure url="https://cdn.example.com/1.mp3" length="123" type="audio/mpeg"/>
      <content:encoded><![CDATA[<p>Notes with a <a href="https://s.example">link</a></p>]]></content:encoded>
    </item>
  </channel>
</rss>`

// Ensure <itunes:title> and <title> are treated as distinct entries.
func TestNamespacesAreDistinguished(test *testing.T) {
	root, err := Parse(strings.NewReader(miniRSS))
	if err != nil {
		test.Fatalf("Parse: %v", err)
	}
	ch := root.Child(NSRSS, "channel")
	if got := ch.ChildText(NSRSS, "title"); got != "TV & Film Weekly" {
		test.Errorf("rss title = %q, want %q", got, "TV & Film Weekly")
	}
	if got := ch.ChildText(NSItunes, "title"); got != "Ignored Alternate Title" {
		test.Errorf("itunes title = %q", got)
	}
}

func TestEntitiesDecodeOnce(test *testing.T) {
	root, _ := Parse(strings.NewReader(miniRSS))
	got := root.Child(NSRSS, "channel").ChildText(NSRSS, "title")
	if strings.Contains(got, "&amp;") {
		test.Errorf("entity not decoded: %q", got)
	}
	if strings.Count(got, "&") != 1 {
		test.Errorf("entity decoded more than once: %q", got)
	}
}

// Attribute order differs by host. Order must not matter.
func TestAttributeOrderIrrelevant(test *testing.T) {
	tests := []struct{ name, doc string }{
		{"url first, as Art19 emits it",
			`<rss><channel><item><enclosure url="u" length="1" type="audio/mpeg"/></item></channel></rss>`},
		{"length first, as Simplecast emits it",
			`<rss><channel><item><enclosure length="1" type="audio/mpeg" url="u"/></item></channel></rss>`},
	}
	for _, testCase := range tests {
		test.Run(testCase.name, func(test *testing.T) {
			root, err := Parse(strings.NewReader(testCase.doc))
			if err != nil {
				test.Fatalf("Parse: %v", err)
			}
			enc := root.Child(NSRSS, "channel").Child(NSRSS, "item").Child(NSRSS, "enclosure")
			if enc.Attr("url") != "u" || enc.Attr("type") != "audio/mpeg" {
				test.Errorf("attrs lost for %q", testCase.doc)
			}
		})
	}
}

func TestCDATAPreservesMarkup(test *testing.T) {
	root, _ := Parse(strings.NewReader(miniRSS))
	it := root.Child(NSRSS, "channel").Child(NSRSS, "item")
	got := it.ChildText(NSContent, "encoded")
	if !strings.Contains(got, "<a href=") {
		test.Errorf("CDATA markup lost: %q", got)
	}
}

// The podcast namespace ships in http and https spellings. Both must land in
// the same canonical token or half the PC2.0 tags vanish.
func TestCanonNSFoldsEveryRealSpelling(test *testing.T) {
	tests := []struct{ name, uri, want string }{
		{"podcast namespace over https", "https://podcastindex.org/namespace/1.0", NSPodcast},
		{"podcast namespace over http", "http://podcastindex.org/namespace/1.0", NSPodcast},
		{"iTunes DTD in mixed case", "http://WWW.iTunes.com/DTDs/Podcast-1.0.dtd", NSItunes},
		{"atom with a trailing slash the alias lacks", "http://www.w3.org/2005/Atom/", NSAtom},
		{"mrss without the trailing slash the alias has", "http://search.yahoo.com/mrss", NSMedia},
		// An unknown namespace passes through unchanged - raw retention needs it.
		{"unknown namespace passes through", "https://example.com/custom", "https://example.com/custom"},
	}
	for _, testCase := range tests {
		test.Run(testCase.name, func(test *testing.T) {
			if got := CanonNS(testCase.uri); got != testCase.want {
				test.Errorf("CanonNS(%q) = %q, want %q", testCase.uri, got, testCase.want)
			}
		})
	}
}

// <link> is in encoding/xml's HTMLAutoClose list. If AutoClose is ever enabled,
// the URL is stranded as a sibling and the closing tag aborts the parse
func TestLinkTextSurvives(test *testing.T) {
	root, err := Parse(strings.NewReader(miniRSS))
	if err != nil {
		test.Fatalf("Parse: %v", err)
	}
	ch := root.Child(NSRSS, "channel")
	if got := ch.ChildText(NSRSS, "link"); got != "https://tvfilm.example" {
		test.Errorf("channel link = %q, want %q", got, "https://tvfilm.example")
	}
	it := ch.Child(NSRSS, "item")
	if it == nil {
		test.Fatal("parsing stopped before <item> - a <link> aborted the document")
	}
	if got := it.ChildText(NSRSS, "link"); got != "https://tvfilm.example/1" {
		test.Errorf("item link = %q", got)
	}
	if got := it.ChildText(NSRSS, "title"); got != "Ep 1" {
		test.Errorf("element after <link> lost: title = %q", got)
	}
}

// Accessors chain on absent elements without a nil check at every call site.
func TestAccessorsAreNilSafe(test *testing.T) {
	root, _ := Parse(strings.NewReader(miniRSS))
	tests := []struct {
		name string
		got  string
	}{
		{"Attr through two absent children", root.Child(NSRSS, "nope").Child(NSRSS, "alsonope").Attr("href")},
		{"ChildText of an absent child", root.ChildText(NSRSS, "nope")},
	}
	for _, testCase := range tests {
		test.Run(testCase.name, func(test *testing.T) {
			if testCase.got != "" {
				test.Errorf("%s returned %q, want empty", testCase.name, testCase.got)
			}
		})
	}
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
