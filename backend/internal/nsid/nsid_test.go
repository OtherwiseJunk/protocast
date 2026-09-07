package nsid

import "testing"

func TestRef(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"basic", "show", "com.cacheblasters.protocast.show"},
		{"empty", "", ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := Ref(test.input)
			if result != test.expected {
				t.Errorf("Ref(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}
