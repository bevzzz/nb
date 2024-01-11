package html

import (
	"html"
	"io"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
)

type Config struct {
	CSSWriter io.Writer
}

type Option func(*Config)

// WithCSSWriter
func WithCSSWriter(w io.Writer) Option {
	return func(c *Config) {
		c.CSSWriter = &WriterOnce{w: w}
	}
}

// Renderer renders the notebook as HTML.
// It supports "markdown", "code", and "raw" cells with different mime-types of the their data.
type Renderer struct {
	render.CellWrapper
	cfg Config
}

// NewRenderer configures a new HTML renderer.
// By default, it embeds a *Wrapper and will panic if it is set to nil by one of the options.
func NewRenderer(opts ...Option) *Renderer {
	var cfg Config
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Renderer{
		CellWrapper: &Wrapper{
			Config: cfg,
		},
		cfg: cfg,
	}
}

func (r *Renderer) RegisterFuncs(reg render.RenderCellFuncRegisterer) {
	reg.Register(schema.MarkdownCellType, r.renderMarkdown)
	reg.Register(schema.CodeCellType, r.renderCode)
	reg.Register(schema.PNG, r.renderImage)
	reg.Register(schema.JPEG, r.renderImage)
	reg.Register(schema.HTML, r.renderRawHTML)
	reg.Register(schema.JSON, r.renderRaw)
	reg.Register(schema.StdoutCellType, r.renderRaw)
	reg.Register(schema.StderrCellType, r.renderRaw)
	reg.Register(schema.PlainTextCellType, r.renderRaw)
}

func (r *Renderer) renderMarkdown(w io.Writer, cell schema.Cell) error {
	io.WriteString(w, "<pre>")
	w.Write(cell.Text())
	io.WriteString(w, "</pre>")
	return nil
}

// renderCode renders the code blob and the code outputs.
func (r *Renderer) renderCode(w io.Writer, cell schema.Cell) error {
	code, ok := cell.(schema.CodeCell)
	if !ok {
		io.WriteString(w, "<pre><code>")
		w.Write(cell.Text())
		io.WriteString(w, "</code></pre>")
		return nil
	}

	div.Open(w, attributes{"class": {"cm-editor", "cm-s-jupyter"}})
	div.Open(w, attributes{"class": {"highlight", "hl-ipython3"}})

	io.WriteString(w, "<pre><code class=\"language-") // TODO: not sure if that's useful here
	io.WriteString(w, code.Language())
	io.WriteString(w, "\">")
	w.Write(code.Text())
	io.WriteString(w, "</code></pre>")

	div.Close(w)
	div.Close(w)

	return nil
}

// renderRawHTML writers raw contents of the cell directly to the document.
func (r *Renderer) renderRawHTML(w io.Writer, cell schema.Cell) error {
	w.Write(cell.Text())
	return nil
}

func (r *Renderer) renderImage(w io.Writer, cell schema.Cell) error {
	io.WriteString(w, "<img src=\"data:")
	io.WriteString(w, string(cell.Type()))
	io.WriteString(w, ";base64, ")
	w.Write(cell.Text())
	io.WriteString(w, "\" />\n")
	return nil
}

// renderRaw writes raw contents of the cell in a new container.
func (r *Renderer) renderRaw(w io.Writer, cell schema.Cell) error {
	io.WriteString(w, "<pre>")
	txt := cell.Text()
	// Escape, because raw text may contain special HTML characters.
	escaped := html.EscapeString(string(txt[:]))
	w.Write([]byte(escaped))
	io.WriteString(w, "</pre>")
	return nil
}
