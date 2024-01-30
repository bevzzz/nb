package jupyter_test

import (
	"strings"
	"testing"

	"github.com/bevzzz/nb"
	"github.com/bevzzz/nb/pkg/test"
	"github.com/bevzzz/nb/schema"
	jupyter "github.com/bevzzz/nb/extension/extra/goldmark-jupyter"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func Test(t *testing.T) {
	for _, tt := range []struct {
		name string
		cell schema.Cell
		opts []html.Option
		want string
	}{
		{
			name: "with attachment",
			cell: test.WithAttachment(
				test.Markdown("![alt](attachment:photo.jpeg)"),
				"photo.jpeg",
				map[string]interface{}{
					"image/jpeg": "base64-image-data",
				},
			),
			want: `<p><img src="data:image/jpeg;base64, base64-image-data" alt="alt"></p>`,
		},
		{
			name: "with html.Options",
			cell: test.WithAttachment(
				test.Markdown("![alt](attachment:photo.jpeg)"),
				"photo.jpeg",
				map[string]interface{}{
					"image/jpeg": "base64-image-data",
				},
			),
			opts: []html.Option{
				html.WithXHTML(), // closes image tag with "/>"
			},
			want: `<p><img src="data:image/jpeg;base64, base64-image-data" alt="alt" /></p>`,
		},
		{
			name: "regular image",
			cell: test.Markdown("![alt](https://example.com/photo)"),
			want: `<p><img src="https://example.com/photo" alt="alt"></p>`,
		},
		{
			name: "regular image with title",
			cell: test.Markdown("![alt](https://example.com/photo \"Title\")"),
			want: `<p><img src="https://example.com/photo" alt="alt" title="Title"></p>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var sb strings.Builder
			md := goldmark.New(
				goldmark.WithExtensions(
					jupyter.Attachments(tt.opts...),
				),
			)

			c := nb.New(
				nb.WithExtensions(
					jupyter.Goldmark(md),
				),
				nb.WithRenderOptions(test.NoWrapper),
			)

			r := c.Renderer()

			// Act
			err := r.Render(&sb, test.Notebook(tt.cell))
			require.NoError(t, err)

			// Assert
			got := strings.Trim(sb.String(), "\n")
			require.Equal(t, tt.want, got, "rendered markdown")
		})
	}
}
