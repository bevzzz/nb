package adapter_test

import (
	"io"
	"strings"
	"testing"

	"github.com/bevzzz/nb/extension/adapter"
	"github.com/bevzzz/nb/pkg/test"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
)

func TestAdapter(t *testing.T) {
	for _, tt := range []struct {
		name   string
		render render.RenderCellFunc
		cell   schema.Cell
		want   string
	}{
		{
			name: "Goldmark",
			render: adapter.Goldmark(func(b []byte, w io.Writer) error {
				w.Write(b)
				return nil
			}),
			cell: test.Markdown("Hi, mom!"),
			want: "Hi, mom!",
		},
		{
			name:   "Blackfriday",
			render: adapter.Blackfriday(func(b []byte) []byte { return b }),
			cell:   test.Markdown("Hi, mom!"),
			want:   "Hi, mom!",
		},
		{
			name:   "AnsiHtml",
			render: adapter.AnsiHtml(func(b []byte) []byte { return b }),
			cell:   test.Stdout("Hi, mom!"),
			want:   "Hi, mom!",
		},
		{
			name:   "AnsiHtml",
			render: adapter.AnsiHtml(func(b []byte) []byte { return b }),
			cell:   test.Stderr("Hi, mom!"),
			want:   "Hi, mom!",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var sb strings.Builder

			// Act
			tt.render(&sb, tt.cell)

			// Assert
			if got := sb.String(); got != tt.want {
				t.Errorf("wrong content: want %q, got %q", tt.want, got)
			}
		})
	}
}
