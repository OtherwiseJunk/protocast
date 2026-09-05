package feedxml

import (
	"strings"
	"testing"
)

const rss20 = `<rss version="2.0"><channel>
  <title>Standard</title>
  <item><title>A</title></item>
  <item><title>B</title></item>
</channel></rss>`

// RDF-flavoured RSS 1.0: items are siblings of channel, not children.
// Libsyn and Fireside both ship this shape.
const rdf10 = `<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#"
                xmlns="http://purl.org/rss/1.0/">
  <channel><title>RDF Flavoured</title></channel>
  <item><title>A</title></item>
  <item><title>B</title></item>
</rdf:RDF>`
const atomFeed = `<feed xmlns="http://www.w3.org/2005/Atom"><title>Atom</title></feed>`
const emptyChannelFeed = `<rss><channel><title>Empty</title></channel></rss>`

func TestLocate(test *testing.T) {

	tests := []struct {
		name, doc, wantTitle string
		wantItems            int
	}{
		{"rss20", rss20, "Standard", 2},
		{"rdf10", rdf10, "RDF Flavoured", 2},
	}
	for _, testCase := range tests {
		test.Run(testCase.name, func(test *testing.T) {
			root, err := Parse(strings.NewReader(testCase.doc))
			if err != nil {
				test.Fatalf("Parse: %v", err)
			}
			channel, items, err := Locate(root)
			if err != nil {
				test.Fatalf("Locate: %v", err)
			}
			if got := channel.ChildText(AnyNS, "title"); got != testCase.wantTitle {
				test.Errorf("channel title = %q, want %q", got, testCase.wantTitle)
			}
			if len(items) != testCase.wantItems {
				test.Fatalf("found %d items, want %d", len(items), testCase.wantItems)
			}
			if got := items[0].ChildText(AnyNS, "title"); got != "A" {
				test.Errorf("first item = %q, want A", got)
			}
		})
	}
}

// A feed with a channel and no items is valid - a show that has published
// nothing yet - and must not be an error.
func TestLocateEmptyChannel(test *testing.T) {
	root, _ := Parse(strings.NewReader(emptyChannelFeed))
	channel, items, err := Locate(root)
	if err != nil {
		test.Fatalf("Locate: %v", err)
	}
	if channel == nil || len(items) != 0 {
		test.Errorf("want channel and zero items, got ch=%v items=%d", channel != nil, len(items))
	}
}

// An Atom-only feed has no <channel>. Fail with a message that says what is
// wrong rather than returning an empty show.
func TestLocateAtomIsAnError(test *testing.T) {
	root, _ := Parse(strings.NewReader(atomFeed))
	if _, _, err := Locate(root); err == nil {
		test.Fatal("Atom feed accepted, want error")
	} else if !strings.Contains(err.Error(), "channel") {
		test.Errorf("unhelpful error: %v", err)
	}
}
