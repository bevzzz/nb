package markdown_test

import (
	"io"
	"strings"
	"testing"

	"github.com/bevzzz/nb/extension/markdown"
	"github.com/bevzzz/nb/internal/test"
)

func TestBlackfriday(t *testing.T) {
	// Arrange
	var sb strings.Builder
	want := "Hi, mom!"
	render := markdown.Blackfriday(func(b []byte) []byte { return b })

	// Act
	render(&sb, test.Markdown("Hi, mom!"))

	// Assert
	if got := sb.String(); got != want {
		t.Errorf("wrong content: want %q, got %q", want, got)
	}
}

func TestGoldmark(t *testing.T) {
	// Arrange
	var sb strings.Builder
	want := "Hi, mom!"
	render := markdown.Goldmark(func(b []byte, w io.Writer) error {
		w.Write(b)
		return nil
	})

	// Act
	render(&sb, test.Markdown("Hi, mom!"))

	// Assert
	if got := sb.String(); got != want {
		t.Errorf("wrong content: want %q, got %q", want, got)
	}
}
