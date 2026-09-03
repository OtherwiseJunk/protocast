package guid

import "testing"

func TestCanonicalise(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://podnews.net/rss", "podnews.net/rss"},
		{"http://podnews.net/rss", "podnews.net/rss"},
		{"podnews.net/rss/", "podnews.net/rss"},
		{"PodNews.NET/rss", "podnews.net/rss"},
		{"https://PodNews.net/rss/", "podnews.net/rss"},
		{"https://podnews.net/RSS", "podnews.net/RSS"},
		{"https://www.podnews.net/rss", "www.podnews.net/rss"},
		{"https://podnews.net/rss?utm_source=x", "podnews.net/rss?utm_source=x"},
		{"https://podnews.net:443/rss", "podnews.net/rss"},
		{"http://podnews.net:80/rss", "podnews.net/rss"},
		{"https://podnews.net:8443/rss", "podnews.net:8443/rss"},
		{"PodNews.net/rss?u=http://x", "podnews.net/rss?u=http://x"},
		{"https://user:pw@example.com/rss", ""},
		{"", ""},
		{"not a url", ""},
		{"../relative/path", ""},
		{"example..com/rss", ""},
		{".example.com/rss", ""},
		{"https://[2001:db8::1]/rss", "[2001:db8::1]/rss"},
	}

	for _, test := range tests {
		got := Canonicalise(test.input)
		if got != test.want {
			t.Errorf("Canonicalise(%q) = %q; want %q", test.input, got, test.want)
		}
	}

}
