package wildcard_test

import (
	"testing"

	"github.com/bevzzz/nb/render/internal/wildcard"
)

func TestMatch(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pattern   string
		wantMatch []string // strings that should match
		noMatch   []string // strings that should not match
	}{
		{
			name:      "matches anything",
			pattern:   "*",
			wantMatch: []string{"", "  ", "*", "*?*", "word"},
		},
		{
			name:      "no wildcard",
			pattern:   "word",
			wantMatch: []string{"word"},
			noMatch:   []string{"vord", "wort"},
		},
		{
			name:      "matches any prefix",
			pattern:   "*tion",
			wantMatch: []string{"caution", "notion", "tion"},
			noMatch:   []string{"extension", "onion"},
		},
		{
			name:      "matches any suffix",
			pattern:   "image/*",
			wantMatch: []string{"image/png", "image/jpeg"},
			noMatch:   []string{"image", "image*"},
		},
		{
			name:      "matches any middle part",
			pattern:   "application/*json",
			wantMatch: []string{"application/json", "application/x+json"},
			noMatch:   []string{"application/jsonc"},
		},
		{
			name:      "multiple wildcards",
			pattern:   "*/*",
			wantMatch: []string{"application/json", "text/plain"},
			noMatch:   []string{"text:csv"},
		},
		{
			name:      "wildcard in place of a repeated string",
			pattern:   "ba*gage",
			wantMatch: []string{"baggage"},
		},
		{
			name:      "redundant wildcards",
			pattern:   "b***k**",
			wantMatch: []string{"book", "books", "bookie", "back"},
			noMatch:   []string{"battle"},
		},
		{
			name:      "both pattern and s empty",
			pattern:   "",
			wantMatch: []string{""},
		},
		{
			name:    "empty input with non-trivial pattern",
			pattern: "s*mething",
			noMatch: []string{""},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, s := range tt.wantMatch {
				if !wildcard.Match(tt.pattern, s) {
					t.Errorf("expected string %q to match pattern %q", s, tt.pattern)
				}
			}
			for _, s := range tt.noMatch {
				if wildcard.Match(tt.pattern, s) {
					t.Errorf("pattern %q should not match string %q", tt.pattern, s)
				}
			}
		})
	}
}

func TestCount(t *testing.T) {
	for _, tt := range []struct {
		s    string
		want int
	}{
		{s: "word", want: 0},
		{s: "image/*", want: 1},
		{s: "*/*-*", want: 3},
	} {
		t.Run(tt.s, func(t *testing.T) {
			if got := wildcard.Count(tt.s); got != tt.want {
				t.Errorf("%q has %d wildcards, counted %d", tt.s, tt.want, got)
			}
		})
	}
}
