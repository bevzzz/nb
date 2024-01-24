package extension_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/bevzzz/nb"
	"github.com/bevzzz/nb/extension"
	"github.com/bevzzz/nb/internal/test"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
	"github.com/stretchr/testify/require"
)

func TestMarkdown(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	want := []byte("Hi, mom!")
	c := nb.New(nb.WithExtensions(
		extension.NewMarkdown(func(w io.Writer, c schema.Cell) error {
			w.Write(want)
			return nil
		}),
	))

	// Override default CellWrapper to compare bare cell contents only.
	r := c.Renderer()
	r.AddOptions(render.WithCellRenderers(&fakeWrapper{}))

	// Act
	err := r.Render(&buf, test.Notebook(test.Markdown("Bye!")))
	require.NoError(t, err)

	// Assert
	if got := buf.Bytes(); !bytes.Equal(want, got) {
		t.Errorf("wrong content: want %q, got %q", want, got)
	}
}

// fakeWrapper calls the passed RenderCellFunc immediately without any additional writes to w.
type fakeWrapper struct{}

var _ render.CellWrapper = (*fakeWrapper)(nil)

func (*fakeWrapper) RegisterFuncs(render.RenderCellFuncRegistry)                    {}
func (*fakeWrapper) Wrap(w io.Writer, c schema.Cell, r render.RenderCellFunc) error { return r(w, c) }
func (*fakeWrapper) WrapInput(w io.Writer, c schema.Cell, r render.RenderCellFunc) error {
	return r(w, c)
}
func (*fakeWrapper) WrapOutput(w io.Writer, out schema.Outputter, r render.RenderCellFunc) error {
	for _, c := range out.Outputs() {
		if err := r(w, c); err != nil {
			return err
		}
	}
	return nil
}
