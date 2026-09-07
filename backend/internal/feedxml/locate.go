package feedxml

import "fmt"

// Locate finds the channel element and the item elements.
//
// Handles both RSS 2.0 which nests <item> inside <channel>,
// and RSS 1.0 which makes <item> a sibling of <channel>.
func Locate(root *Node) (*Node, []*Node, error) {
	const channelStr = "channel"
	const itemStr = "item"
	if root == nil {
		return nil, nil, fmt.Errorf("feedxml.Locate: empty document")
	}

	channel := root.Child(AnyNS, channelStr)
	if channel == nil && root.Local == channelStr {
		channel = root
	}
	if channel == nil {
		return nil, nil, fmt.Errorf(
			"feedxml: no <channel> under <%s>: not an RSS feed (Atom is not supported)\n\n"+
				"Developer's Note: If you are actually attempting to ingest an Atom feed please "+
				"open an issue on https://github.com/otherwisejunk/protocast/issues including "+
				"the feed URL you submitted", root.Local)
	}

	items := channel.ChildrenNamed(AnyNS, "item")
	if len(items) == 0 {
		items = root.ChildrenNamed(AnyNS, "item")
	}

	return channel, items, nil
}
