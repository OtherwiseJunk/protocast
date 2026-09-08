package feedxml

import (
	"runtime"
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

func TestDefectRealChannelWinsOverForeignNamespace(test *testing.T) {
	doc := `<rss version="2.0" xmlns:x="http://example.invalid/x"><x:channel>` +
		`<title>Decoy</title></x:channel><channel><title>Real</title>` +
		`<item><title>Ep</title></item></channel></rss>`
	root, err := Parse(strings.NewReader(doc))
	if err != nil {
		test.Fatalf("Parse: %v", err)
	}
	ch, items, err := Locate(root)
	if err != nil {
		test.Fatalf("Locate: %v", err)
	}
	if got := ch.ChildText(AnyNS, "title"); got != "Real" {
		test.Errorf("Locate chose the channel titled %q; a foreign-namespace <x:channel> "+
			"declared first became the show", got)
	}
	if len(items) != 1 {
		test.Errorf("Locate found %d items; the real channel's item was lost with it", len(items))
	}
}

func TestDefectNamespacedAttrDoesNotShadowPlain(test *testing.T) {
	doc := `<rss version="2.0" xmlns:x="http://example.invalid/x"><channel><item>` +
		`<enclosure url="https://good.example/a.mp3" x:url="https://evil.example/a.mp3"/>` +
		`</item></channel></rss>`
	root, err := Parse(strings.NewReader(doc))
	if err != nil {
		test.Fatalf("Parse: %v", err)
	}
	_, items, err := Locate(root)
	if err != nil || len(items) != 1 {
		test.Fatalf("Locate: %v, %d items", err, len(items))
	}
	if got := items[0].Child(AnyNS, "enclosure").Attr("url"); got != "https://good.example/a.mp3" {
		test.Errorf("Attr(\"url\") = %q; a namespaced x:url overwrote the unnamespaced url", got)
	}
}

func TestDefectTextAccumulationIsLinear(test *testing.T) {
	alloc := func(tokens int) uint64 {
		doc := `<?xml version="1.0"?><rss version="2.0"><channel><description>` +
			strings.Repeat("A<!---->", tokens) + `</description></channel></rss>`
		var before, after runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&before)
		if _, err := Parse(strings.NewReader(doc)); err != nil {
			test.Fatalf("Parse: %v", err)
		}
		runtime.ReadMemStats(&after)
		return after.TotalAlloc - before.TotalAlloc
	}
	small := alloc(20000)
	large := alloc(40000)
	// Linear accumulation roughly doubles. Quadratic roughly quadruples.
	if ratio := float64(large) / float64(small); ratio > 3 {
		test.Errorf("doubling the token count allocated %.1fx the bytes (%d -> %d): "+
			"Text += string(t) is quadratic, so a feed at the 32 MiB fetch cap "+
			"occupies a core for tens of minutes", ratio, small, large)
	}
}

func TestDefectSurvivesOneBadByte(test *testing.T) {
	const head = `<?xml version="1.0" encoding="UTF-8"?><rss version="2.0"><channel>` +
		`<title>Probe</title><link>https://ex.com</link><description>d</description>`
	const tail = `</channel></rss>`
	tests := []struct{ name, doc string }{
		{"lone Latin-1 byte in the last item",
			head + "<item><title>Ok</title></item><item><title>Caf\xe9</title></item>" + tail},
		{"NUL character reference", head + "<item><title>&#0;</title></item>" + tail},
		{"C0 control character reference", head + "<item><title>&#8;</title></item>" + tail},
		{"bare < in text", head + "<item><title>5 < 6</title></item>" + tail},
		{"unknown charset label", `<?xml version="1.0" encoding="x-mac-croatian"?>` +
			`<rss version="2.0"><channel><title>Probe</title>` + tail},
	}
	for _, testCase := range tests {
		test.Run(testCase.name, func(test *testing.T) {
			root, err := Parse(strings.NewReader(testCase.doc))
			if err != nil && root == nil {
				test.Errorf("one malformed byte discarded the whole document: %v\n"+
					"a real feed must survive with the bad value degraded, not vanish", err)
			}
		})
	}
}
