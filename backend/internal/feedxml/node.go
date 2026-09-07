// Pacakge feedxml parses podcast RSS into a generic namespace-aware node tree.
package feedxml

import "strings"

// Canonical namespace tokens. RSS 2.0's elements are unnamespaced.
const (
	NSRSS     = ""
	NSItunes  = "itunes"
	NSPodcast = "podcast"
	NSAtom    = "atom"
	NSContent = "content"
	NSMedia   = "media"
	NSDC      = "dc"
	AnyNS     = "*"
)

var nsAliases = map[string]string{
	"http://www.itunes.com/dtds/podcast-1.0.dtd": NSItunes,
	"https://podcastindex.org/namespace/1.0":     NSPodcast,
	"http://podcastindex.org/namespace/1.0":      NSPodcast,
	"https://podcastindex.org/namespace/1.0/":    NSPodcast,
	"http://www.w3.org/2005/Atom":                NSAtom,
	"http://purl.org/rss/1.0/modules/content/":   NSContent,
	"http://search.yahoo.com/mrss/":              NSMedia,
	"http://purl.org/dc/elements/1.1/":           NSDC,
}

// CanonNS maps a namespace URI to a short stable token. Unknown namespaces pass
// through unchanged so nothing is silently lost.
func CanonNS(uri string) string {
	if uri == "" {
		return NSRSS
	}
	if str, ok := nsAliases[uri]; ok {
		return str
	}
	norm := strings.TrimSuffix(uri, "/")
	for key, value := range nsAliases {
		if strings.EqualFold(strings.TrimSuffix(key, "/"), norm) {
			return value
		}
	}
	return uri
}

// Node represents one XML element.
type Node struct {
	Space    string
	Local    string
	Attrs    map[string]string
	Text     string
	Children []*Node
}

func match(node *Node, space, local string) bool {
	return node.Local == local && (space == AnyNS || node.Space == space)
}

// Child returns the first matching child, or nil.
func (node *Node) Child(space, local string) *Node {
	if node == nil {
		return nil
	}
	for _, child := range node.Children {
		if match(child, space, local) {
			return child
		}
	}
	return nil
}

// ChildrenNamed returns every matching child, in document order.
func (node *Node) ChildrenNamed(space, local string) []*Node {
	if node == nil {
		return nil
	}
	var out []*Node
	for _, child := range node.Children {
		if match(child, space, local) {
			out = append(out, child)
		}
	}
	return out
}

// ChildText returns the trimmed text of the first matching child.
//
// If no match is found or the element is empty "" is returned.
func (node *Node) ChildText(space, local string) string {
	return strings.TrimSpace(node.Child(space, local).TextTrimmed())
}

// TextTrimmed is nil-safe access to a node's own text.
func (node *Node) TextTrimmed() string {
	if node == nil {
		return ""
	}
	return strings.TrimSpace(node.Text)
}

// Attr returns a trimmed attribute value by local name, or "".
func (node *Node) Attr(name string) string {
	if node == nil {
		return ""
	}
	return strings.TrimSpace(node.Attrs[name])
}
