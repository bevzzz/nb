package test

import (
	"io"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
)

// NoWrapper overrides the default cell wrapper so that the cell content could be compared
// directly without parsing the surrounding wrap. Useful for testing extensions.
var NoWrapper = render.WithCellRenderers(&fakeWrapper{})

// fakeWrapper calls the passed RenderCellFunc immediately without any additional writes to w.
type fakeWrapper struct{}

var _ render.CellWrapper = (*fakeWrapper)(nil)

func (*fakeWrapper) RegisterFuncs(render.RenderCellFuncRegistry)                    {}
func (*fakeWrapper) WrapAll(io.Writer, func(io.Writer) error) error                 { return nil }
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
