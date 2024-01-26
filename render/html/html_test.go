package html_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/bevzzz/nb/pkg/test"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/render/html"
	"github.com/bevzzz/nb/schema"
)

func TestRenderer(t *testing.T) {
	t.Run("renders expected html", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			cell schema.Cell
			want *node
		}{
			{
				name: "markdown cell",
				cell: test.Markdown("# List:- One\n- Two\n -Three"),
				want: &node{tag: "pre", content: "# List:- One\n- Two\n -Three"},
			},
			{
				name: "raw text/html",
				cell: test.Raw("<h1>Hi, mom!</h1>", "text/html"),
				want: &node{tag: "h1", content: "Hi, mom!"},
			},
			{
				name: "raw text/plain",
				cell: test.Raw("asdf", "text/plain"),
				want: &node{tag: "pre", content: "asdf"},
			},
			{
				name: "application/json",
				cell: test.DisplayData(`{"one":1,"two":2}`, "application/json"),
				want: &node{tag: "pre", content: `{"one":1,"two":2}`},
			},
			{
				name: "stream to stdout",
				cell: test.Stdout("Two o'clock, and all's well!"),
				want: &node{tag: "pre", content: "Two o'clock, and all's well!"},
			},
			{
				name: "stream to stderr",
				cell: test.Stderr("Mayday!Mayday!"),
				want: &node{tag: "pre", content: "Mayday!Mayday!"},
			},
			{
				name: "image/png",
				cell: test.DisplayData("base64-encoded-image", "image/png"),
				want: &node{tag: "img", attr: map[string][]string{
					"src": {"data:image/png;base64, base64-encoded-image"},
				}},
			},
			{
				name: "image/jpeg",
				cell: test.DisplayData("base64-encoded-image", "image/jpeg"),
				want: &node{tag: "img", attr: map[string][]string{
					"src": {"data:image/jpeg;base64, base64-encoded-image"},
				}},
			},
			{
				name: "code cell",
				cell: &test.CodeCell{
					Cell: test.Cell{
						CellType: schema.Code,
						Source:   []byte("print('Hi, mom!')"),
					},
					Lang: "python",
				},
				want: &node{
					tag: "div",
					attr: map[string][]string{
						"class": {"cm-editor", "cm-s-jupyter"},
					},
					children: []*node{
						{
							tag: "div",
							attr: map[string][]string{
								"class": {"highlight"},
							},
							children: []*node{
								{
									tag: "pre",
									children: []*node{
										{
											tag: "code",
											attr: map[string][]string{
												"class": {"language-python"},
											},
											content: "print('Hi, mom!')",
										},
									}},
							},
						},
					},
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				// Arrange
				var buf bytes.Buffer
				r := render.NewRenderer()
				reg := r.(render.RenderCellFuncRegistry)
				html.NewRenderer().RegisterFuncs(reg)

				// Act
				err := r.Render(&buf, test.Notebook(tt.cell))
				require.NoError(t, err)

				// Assert
				checkDOM(t, &buf, tt.want)
			})
		}
	})
}

func TestRenderer_CSSWriter(t *testing.T) {
	t.Run("WithCSSWriter wraps in WriterOnce", func(t *testing.T) {
		// Arrange
		var cfg html.Config
		opt := html.WithCSSWriter(io.Discard)

		// Act
		opt(&cfg)

		// Assert
		if _, ok := cfg.CSSWriter.(*html.WriterOnce); !ok {
			t.Errorf("expected *html.WriterOnce, got %T", cfg.CSSWriter)
		}
	})

	t.Run("captures correct css", func(t *testing.T) {
		// Arrange
		var css bytes.Buffer
		var want []byte
		var err error

		r := html.NewRenderer(html.WithCSSWriter(&css))
		if want, err = os.ReadFile("styles/jupyter.css"); err != nil {
			t.Skip(err)
		}

		// Act
		err = r.Wrap(io.Discard, test.Markdown(""), noopRender)
		require.NoError(t, err)

		// Assert
		if got := css.Bytes(); !bytes.Equal(got, want) {
			t.Errorf("wrong css (-got), (+want):\n%s", cmp.Diff(got, want))
		}
	})
}
