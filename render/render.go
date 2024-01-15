package render

import (
	"fmt"
	"io"
	"sync"

	"github.com/bevzzz/nb/schema"
)

// Renderer renders a decoded notebook in the format it implements.
type Renderer interface {
	// Render writes the contents of the notebook cells it supports.
	//
	// Implementations should not error on cell types, for which no RenderCellFunc is registered.
	// This is expected, as some [RawCells] will be rendered in some output formats and ignored in others.
	//
	// [RawCells]: https://nbformat.readthedocs.io/en/latest/format_description.html#raw-nbconvert-cells
	Render(io.Writer, schema.Notebook) error

	// AddOptions configures the editor after it has been constructured.
	// The renderer's configuration should not change between renders, and so, implementations should
	// ignore options added after the first call to Render().
	AddOptions(...Option)
}

// CellRenderer registers a RenderCellFunc for every cell type it supports.
//
// Reminiscent of the [Visitor] pattern, it allows extending the base renderer
// to support any number of arbitrary cell types.
//
// [Visitor]: https://refactoring.guru/design-patterns/visitor
type CellRenderer interface {
	// RegisterFuncs registers one or more RenderCellFunc with the passed renderer.
	RegisterFuncs(RenderCellFuncRegisterer)
}

// RenderCellFuncRegisterer is an interface that extendable Renderers should implement.
type RenderCellFuncRegisterer interface {
	// Register adds a RenderCellFunc that will be called for cells of this type.
	Register(schema.CellTypeMixed, RenderCellFunc)
}

// RenderCellFunc writes contents of a specific cell type.
type RenderCellFunc func(io.Writer, schema.Cell) error

type Option func(r *renderer)

// WithCellRenderers adds support for other cell types to the base renderer.
// If a renderer implements CellWrapper, it will be used to wrap input and output cells.
// Only one cell wrapper can be configured, and so the last implementor will take precedence.
func WithCellRenderers(crs ...CellRenderer) Option {
	return func(r *renderer) {
		for _, cr := range crs {
			cr.RegisterFuncs(r)

			if cw, ok := cr.(CellWrapper); ok {
				r.cellWrapper = cw
			}
		}
	}
}

// CellWrapper renders common wrapping elements for every cell type.
type CellWrapper interface {
	// Wrap the entire cell block.
	Wrap(io.Writer, schema.Cell, RenderCellFunc) error

	// Wrap input block.
	WrapInput(io.Writer, schema.Cell, RenderCellFunc) error

	// Wrap output block (code cells).
	WrapOutput(io.Writer, schema.Outputter, RenderCellFunc) error
}

// renderer is a base Renderer implementation.
// It does not support any cell types out of the box and should be extended by the client using the available Options.
type renderer struct {
	once               sync.Once
	cellWrapper        CellWrapper
	renderCellFuncsTmp map[schema.CellTypeMixed]RenderCellFunc
	renderCellFuncs    map[schema.CellTypeMixed]RenderCellFunc
}

// New extends the base renderer with the passed options.
func New(opts ...Option) Renderer {
	r := renderer{
		renderCellFuncsTmp: make(map[schema.CellTypeMixed]RenderCellFunc),
		renderCellFuncs:    make(map[schema.CellTypeMixed]RenderCellFunc),
		cellWrapper:        nil,
	}
	r.AddOptions(opts...)
	return &r
}

func (r *renderer) AddOptions(opts ...Option) {
	for _, opt := range opts {
		opt(r)
	}
}

// Register registers a new RenderCellFunc for the cell type.
//
// Any previously registered functions will be overridden. All configurations
// should be done the first call to Render(), as later changes will have no effect.
func (r *renderer) Register(t schema.CellTypeMixed, f RenderCellFunc) {
	r.renderCellFuncsTmp[t] = f
}

func (r *renderer) init() {
	r.once.Do(func() {
		for t, f := range r.renderCellFuncsTmp {
			r.renderCellFuncs[t] = f
		}
	})
}

// render the contents of a cell if a RenderCellFunc is registered for its type.
func (r *renderer) render(w io.Writer, cell schema.Cell) error {
	render, ok := r.renderCellFuncs[cell.Type()]
	if ok {
		if err := render(w, cell); err != nil {
			return fmt.Errorf("ipynb: render: %w", err)
		}
	}
	// TODO: currently we silently drop cells for which no render func is registered. Should we error?
	return nil
}

func (r *renderer) Render(w io.Writer, nb schema.Notebook) error {
	r.init()

	for _, cell := range nb.Cells() {
		var err error

		if r.cellWrapper != nil {
			err = r.cellWrapper.Wrap(w, cell, func(w io.Writer, c schema.Cell) error {
				if err := r.cellWrapper.WrapInput(w, cell, r.render); err != nil {
					return err
				}

				if out, ok := cell.(interface{ schema.Outputter }); ok {
					if err := r.cellWrapper.WrapOutput(w, out, r.render); err != nil {
						return err
					}
				}
				return nil
			})
		} else {
			err = r.render(w, cell)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
