package nb

import (
	"io"

	"github.com/bevzzz/nb/decode"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/render/html"
)

// Convert a Jupyter notebook using default converter.
func Convert(w io.Writer, source []byte) error {
	return defaultNotebook.Convert(w, source)
}

var defaultNotebook = New()

// Converter converts raw Jupyter Notebook JSON to the selected format.
type Converter interface {
	Convert(w io.Writer, source []byte) error
}

// WithRenderer sets a new notebook renderer.
// This option should be supplied before passing any WithRenderOptions.
func WithRenderer(r render.Renderer) Option {
	return func(n *Notebook) {
		n.renderer = r
	}
}

// WithRendererOptions adds configuration to the notebook renderer.
func WithRenderOptions(opts ...render.Option) Option {
	return func(n *Notebook) {
		n.renderer.AddOptions(opts...)
	}
}

// Notebook is an extensible Converter implementation.
type Notebook struct {
	renderer render.Renderer
}

var _ Converter = (*Notebook)(nil)

type Option func(*Notebook)

// New returns a converter with default HTML renderer and extensions.
func New(opts ...Option) *Notebook {
	nb := Notebook{
		renderer: DefaultRenderer(),
	}
	for _, opt := range opts {
		opt(&nb)
	}
	return &nb
}

// DefaultRenderer configures an HTML renderer.
func DefaultRenderer() render.Renderer {
	return render.NewRenderer(
		render.WithCellRenderers(html.NewRenderer()),
	)
}

// Сonvert raw Jupyter Notebook JSON and write the output.
func (n *Notebook) Convert(w io.Writer, source []byte) error {
	// nb, err := decode.Decode(source)
	nb, err := decode.Bytes(source)
	if err != nil {
		return err
	}
	return n.renderer.Render(w, nb)
}
