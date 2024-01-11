package html_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/render/html"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

func TestRenderer(t *testing.T) {
	t.Run("handles basic cell/mime types by default", func(t *testing.T) {
		// Arrange
		reg := make(funcRegistry)
		r := html.NewRenderer()

		// Act
		r.RegisterFuncs(reg)

		// Assert
		for _, ct := range []schema.CellTypeMixed{
			schema.CodeCellType,
			schema.HTML,
			schema.MarkdownCellType,
			schema.JSON,
			schema.PNG,
			schema.JPEG,
			schema.StdoutCellType,
			schema.StderrCellType,
			schema.PlainTextCellType,
		} {
			require.Contains(t, reg, ct, "expected a RenderCellFunc for cell type %q", ct)
		}
	})

	t.Run("renders expected html", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			cell schema.Cell
			want *node
		}{
			{
				name: "markdown cell",
				cell: markdown("# List:- One\n- Two\n -Three"),
				want: &node{tag: "pre", content: "# List:- One\n- Two\n -Three"},
			},
			{
				name: "raw text/html",
				cell: raw("text/html", "<h1>Hi, mom!</h1>"),
				want: &node{tag: "h1", content: "Hi, mom!"},
			},
			{
				name: "raw text/plain",
				cell: raw("text/html", "asdf"),
				want: &node{tag: "pre", content: "asdf"},
			},
			{
				name: "application/json",
				cell: displaydata("application/json", `{"one":1,"two":2}`),
				want: &node{tag: "pre", content: `{"one":1,"two":2}`},
			},
			{
				name: "stream to stdout",
				cell: stdout("Two o'clock, and all's well!"),
				want: &node{tag: "pre", content: "Two o'clock, and all's well!"},
			},
			{
				name: "stream to stderr",
				cell: stderr("Mayday!Mayday!"),
				want: &node{tag: "pre", content: "Mayday!Mayday!"},
			},
			{
				name: "image/png",
				cell: displaydata("image/png", "base64-encoded-image"),
				want: &node{tag: "img", attr: map[string][]string{
					"src": {"data:image/png;base64, base64-encoded-image"},
				}},
			},
			{
				name: "image/jpeg",
				cell: displaydata("image/jpeg", "base64-encoded-image"),
				want: &node{tag: "img", attr: map[string][]string{
					"src": {"data:image/jpeg;base64, base64-encoded-image"},
				}},
			},
			{
				name: "code cell",
				cell: &CodeCell{
					Cell: Cell{
						ct:     schema.Code,
						source: []byte("print('Hi, mom!')"),
					},
					language: "python",
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
				reg := make(funcRegistry)
				html.NewRenderer().RegisterFuncs(reg)

				ct := tt.cell.Type()
				rf, ok := reg[tt.cell.Type()]
				if !ok {
					t.Fatalf("no function registered for %q cell", ct)
				}

				// Act
				err := rf(&buf, tt.cell)
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
		err = r.Wrap(io.Discard, markdown(""), noopRender)
		require.NoError(t, err)

		// Assert
		if got := css.Bytes(); !bytes.Equal(got, want) {
			t.Errorf("wrong css (-got), (+want):\n%s", cmp.Diff(got, want))
		}
	})
}

// funcRegistry implements render.RenderCellFuncRegisterer for a plain map.
type funcRegistry map[schema.CellTypeMixed]render.RenderCellFunc

var _ render.RenderCellFuncRegisterer = (*funcRegistry)(nil)

func (r funcRegistry) Register(ct schema.CellTypeMixed, f render.RenderCellFunc) {
	r[ct] = f
}

func markdown(s string) schema.Cell {
	return &Cell{ct: schema.Markdown, mimeType: common.MarkdownText, source: []byte(s)}
}

func raw(mt string, s string) schema.Cell {
	return &Cell{ct: schema.Raw, mimeType: mt, source: []byte(s)}
}

func displaydata(mt string, s string) schema.Cell {
	return &Cell{ct: schema.DisplayData, mimeType: mt, source: []byte(s)}
}

func stdout(s string) schema.Cell {
	return &Cell{ct: schema.Stream, mimeType: common.Stdout, source: []byte(s)}
}

func stderr(s string) schema.Cell {
	return &Cell{ct: schema.Stream, mimeType: common.Stderr, source: []byte(s)}
}

// Cell is a test fixture to mock schema.Cell.
type Cell struct {
	ct       schema.CellType
	mimeType string
	source   []byte
}

var _ schema.Cell = (*Cell)(nil)

func (c *Cell) CellType() schema.CellType { return c.ct }
func (c *Cell) MimeType() string          { return c.mimeType }
func (c *Cell) Text() []byte              { return c.source }

// TODO: drop
func (c *Cell) Type() schema.CellTypeMixed {
	switch c.ct {
	case schema.Markdown:
		return schema.MarkdownCellType
	case schema.Code:
		return schema.CodeCellType
	case schema.Stream:
		if c.mimeType == common.Stdout {
			return schema.StdoutCellType
		}
		return schema.StderrCellType
	}
	return schema.CellTypeMixed(c.mimeType)
}

// CodeCell is a test fixture to mock schema.CodeCell.
type CodeCell struct {
	Cell
	language       string
	executionCount int
	outputs        []schema.Cell
}

var _ schema.CodeCell = (*CodeCell)(nil)

func (code *CodeCell) Language() string       { return code.language }
func (code *CodeCell) ExecutionCount() int    { return code.executionCount }
func (code *CodeCell) Outputs() []schema.Cell { return code.outputs }

// ExecuteResultOutput is a test fixture to mock cell outputs with ExecuteResult type.
type ExecuteResultOutput struct {
	Cell
	executionCount int
}

var _ schema.Cell = (*ExecuteResultOutput)(nil)
var _ interface{ ExecutionCount() int } = (*ExecuteResultOutput)(nil)

// TODO: drop
var _ interface{ TimesExecuted() int } = (*ExecuteResultOutput)(nil)

func (ex *ExecuteResultOutput) ExecutionCount() int { return ex.executionCount }
func (ex *ExecuteResultOutput) TimesExecuted() int  { return ex.executionCount }
