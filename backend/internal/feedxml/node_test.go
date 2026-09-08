package feedxml

import "testing"

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
		{"itunes over https", "https://www.itunes.com/dtds/podcast-1.0.dtd", NSItunes},
		{"atom over https", "https://www.w3.org/2005/Atom", NSAtom},
		{"content over https", "https://purl.org/rss/1.0/modules/content/", NSContent},
		{"mrss over https", "https://search.yahoo.com/mrss/", NSMedia},
		{"dublin core over https", "https://purl.org/dc/elements/1.1/", NSDC},
	}
	for _, testCase := range tests {
		test.Run(testCase.name, func(test *testing.T) {
			if got := CanonNS(testCase.uri); got != testCase.want {
				test.Errorf("CanonNS(%q) = %q, want %q", testCase.uri, got, testCase.want)
			}
		})
	}
}

func TestNSAliasKeysAreCanonical(test *testing.T) {
	for key, token := range nsAliases {
		if got := nsKey(key); got != key {
			test.Errorf("alias key %q is not canonical (nsKey gives %q); "+
				"CanonNS can never match it, so %q is unreachable", key, got, token)
		}
	}
}
