package render_test

import (
	"io"
	"strings"
	"testing"

	"github.com/bevzzz/nb/internal/test"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
	"github.com/stretchr/testify/require"
)

func TestRenderer_Render(t *testing.T) {
	// More than anything, this test ensures that renderer
	// will correctly deduplicate and prioritize RenderCellFuncs
	// registered by extensions. That way, concrete CellRenderer implementations
	// will only need to test that their Prefs capture all their target cells.

	r, ok := render.NewRenderer().(render.RenderCellFuncRegistry)
	if !ok {
		t.Errorf("%T does not implement render.RenderCellFuncRegisterer", r)
	}

	// writeString returns a render.RenderCellFunc that writes s to w.
	writeString := func(s string) render.RenderCellFunc {
		return func(w io.Writer, c schema.Cell) error {
			io.WriteString(w, s)
			return nil
		}
	}

	for _, tt := range []struct {
		name     string
		standard renderCellFuncs // standard functions immitate existing (default) render cell funcs
		prefs    renderCellFuncs // functions expected to be used in the Act step
		cell     schema.Cell
		want     string
	}{
		{
			name: "any markdown cell",
			standard: renderCellFuncs{
				render.Pref{Type: schema.Markdown}: writeString("default markdown"),
			},
			prefs: renderCellFuncs{
				render.Pref{Type: schema.Markdown}: writeString("custom markdown"),
			},
			cell: test.Markdown(""),
			want: "custom markdown",
		},
		{
			name: "exact mime-type overrides wildcard",
			standard: renderCellFuncs{
				render.Pref{MimeType: "text/*"}: writeString("any text"),
			},
			prefs: renderCellFuncs{
				render.Pref{MimeType: common.MarkdownText}: writeString("custom markdown"),
			},
			cell: test.Markdown(""),
			want: "custom markdown",
		},
		{
			name: "cell type + mime-type overrides exact mime-type",
			standard: renderCellFuncs{
				render.Pref{MimeType: "image/png"}: writeString("any PNG image"),
			},
			prefs: renderCellFuncs{
				render.Pref{Type: schema.DisplayData, MimeType: "image/png"}: writeString("display data PNG"),
			},
			cell: test.DisplayData("", "image/png"),
			want: "display data PNG",
		},
		{
			name: "mime-type will less wildcards is prioritized",
			standard: renderCellFuncs{
				render.Pref{MimeType: "*/*"}: writeString("any mime-type"),
			},
			prefs: renderCellFuncs{
				render.Pref{MimeType: "text/*"}: writeString("any text"),
			},
			cell: test.Raw("", "text/html"),
			want: "any text",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			r := render.NewRenderer()
			reg := r.(render.RenderCellFuncRegistry)
			tt.standard.RegisterFuncs(reg)

			tt.prefs.RegisterFuncs(reg)
			var sb strings.Builder

			// Act
			err := r.Render(&sb, test.Notebook(tt.cell))
			require.NoError(t, err)

			// Assert
			if got := sb.String(); got != tt.want {
				t.Errorf("wrong content: want %q, got %q", tt.want, got)
			}
		})
	}

}

// renderCellFuncs implements render.CellRenderer for a map[render.Pref]render.RenderCellFunc.
type renderCellFuncs map[render.Pref]render.RenderCellFunc

var _ render.CellRenderer = new(renderCellFuncs)

func (sr renderCellFuncs) RegisterFuncs(reg render.RenderCellFuncRegistry) {
	for pref := range sr {
		reg.Register(pref, sr[pref])
	}
}

func TestPref_Match(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pref      render.Pref
		wantMatch []schema.Cell
		noMatch   []schema.Cell
	}{
		{
			name:      "only cell type",
			pref:      render.Pref{Type: schema.Markdown},
			wantMatch: []schema.Cell{test.Markdown("")},
			noMatch: []schema.Cell{
				test.Raw("", "text/markdown"),
				&test.Cell{CellType: schema.Code},
			},
		},
		{
			name: "only mime-type",
			pref: render.Pref{MimeType: "image/*"},
			wantMatch: []schema.Cell{
				test.Raw("", "image/jpeg"),
				test.DisplayData("", "image/png"),
				test.ExecuteResult("", "image/svg+xml", 0),
			},
			noMatch: []schema.Cell{
				test.Raw("", "text/html"),
				test.Markdown(""),
				test.Stdout(""),
			},
		},
		{
			name: "cell type and mime-type",
			pref: render.Pref{Type: schema.Code, MimeType: "*/javascript"},
			wantMatch: []schema.Cell{
				&test.Cell{CellType: schema.Code, Mime: "text/javascript"},
				&test.Cell{CellType: schema.Code, Mime: "application/javascript"},
			},
			noMatch: []schema.Cell{
				&test.Cell{CellType: schema.Code, Mime: "text/js"},
				&test.Cell{CellType: schema.Code, Mime: "application/x+javascript"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, cell := range tt.wantMatch {
				if got := tt.pref.Match(cell); !got {
					t.Errorf("%+v should match cell %+v", tt.pref, cell)
				}
			}

			for _, cell := range tt.noMatch {
				if got := tt.pref.Match(cell); got {
					t.Errorf("%+v should not match cell %+v", tt.pref, cell)
				}
			}
		})
	}
}
