package app

import (
	"reflect"
	"testing"
)

func TestExtractURLs(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"empty", "", nil},
		{"no urls", "fix the bug in auth flow", nil},
		{"single http", "see http://example.com for details", []string{"http://example.com"}},
		{"single https", "https://example.com/path?q=1", []string{"https://example.com/path?q=1"}},
		{"trailing period", "visit https://example.com.", []string{"https://example.com"}},
		{"trailing comma", "see https://example.com, then https://other.com", []string{"https://example.com", "https://other.com"}},
		{"trailing paren", "(see https://example.com)", []string{"https://example.com"}},
		{"trailing question", "have you seen https://example.com?", []string{"https://example.com"}},
		{"multiple", "https://a.com and https://b.com/x", []string{"https://a.com", "https://b.com/x"}},
		{"dedupe", "https://a.com and https://a.com again", []string{"https://a.com"}},
		{"ignores ftp", "ftp://example.com is not http", nil},
		{"strips angle brackets", "<https://example.com>", []string{"https://example.com"}},
		{"strips backticks", "see `https://example.com` here", []string{"https://example.com"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractURLs(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("extractURLs(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}
