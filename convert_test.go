package nb_test

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/bevzzz/nb"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/render/html"
	"github.com/bevzzz/nb/schema"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

// update allows updating golden files via `go test -update`.
var update = flag.Bool("update", false, "update .golden files in testdata/")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestConvert(t *testing.T) {
	for _, tt := range []struct {
		name   string
		ipynb  string
		golden string
		c      nb.Converter
	}{
		{
			name:   "complete notebook",
			ipynb:  "notebook",
			golden: "notebook",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ipynb, err := os.ReadFile("testdata/" + tt.ipynb + ".ipynb")
			require.NoError(t, err)

			// Act
			var got bytes.Buffer
			if tt.c == nil {
				err = nb.Convert(&got, ipynb)
			} else {
				err = tt.c.Convert(&got, ipynb)
			}
			require.NoError(t, err)

			// Assert
			cmpGolden(t, "testdata/"+tt.golden+".golden", got.Bytes(), *update)
		})
	}
}

// cmpGolden compares the result of the test run with a golden file.
// If the contents don't match and upd == true, it will update the golden file
// with the current value instead of failing the test.
func cmpGolden(tb testing.TB, goldenFile string, got []byte, upd bool) {
	gf, err := os.OpenFile(goldenFile, os.O_RDWR, 0644)
	require.NoError(tb, err)
	defer gf.Close()

	want, err := io.ReadAll(gf)
	require.NoError(tb, err)

	dotnew := gf.Name() + ".new"
	del := func() {
		files, _ := filepath.Glob("testdata/*.golden.new")
		for i := range files {
			if err := os.Remove(files[i]); err != nil {
				tb.Log(err)
				continue
			}
			log.Printf("deleted previous %s file", dotnew)
		}
	}

	if bytes.Equal(want, got) {
		del()
		return
	}

	if upd {
		err = gf.Truncate(0)
		require.NoError(tb, err)

		gf.Seek(0, 0)
		_, err := gf.Write(got)
		require.NoError(tb, err)

		log.Printf("updated %s", goldenFile)
		del()
		return
	}

	tb.Errorf("mismatched output (+want) (-got):\n%s", cmp.Diff(string(want), string(got)))

	if err := os.WriteFile(dotnew, got, 0644); err == nil {
		tb.Logf("saved to %s (the file will be deleted on the next `-update` or successful test run)", dotnew)
	} else {
		tb.Logf("failed to save %s: %v", dotnew, err)
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithRenderer", func(t *testing.T) {
		// Arrange
		r := render.NewRenderer()

		// Act
		n := nb.New(nb.WithRenderer(r))

		// Assert
		if n.Renderer() != r {
			t.Error("option was not applied")
		}
	})

	t.Run("WithRendererOptions", func(t *testing.T) {
		// Arrange
		var spy spyRenderer

		// Act
		_ = nb.New(
			nb.WithRenderer(&spy),
			nb.WithRenderOptions(
				render.WithCellRenderers(html.NewRenderer()),
				render.WithCellRenderers(html.NewRenderer()),
			),
		)

		// Assert
		if l := len(spy.AddedOptions); l != 2 {
			t.Errorf("expected %d options applied, got %d", 2, l)
		}
	})

	t.Run("WithExtensions", func(t *testing.T) {
		// Arrange
		var spy spyRenderer
		ext := mockExtension{options: []render.Option{
			render.WithCellRenderers(html.NewRenderer()),
		}}

		// Act
		_ = nb.New(
			nb.WithRenderer(&spy),
			nb.WithExtensions(&ext),
		)

		// Assert
		if len(spy.AddedOptions) == 0 {
			t.Errorf("option not applied or applied incorrectly")
		}
	})
}

// spyRenderer records info about that calls that were made to its methods.
type spyRenderer struct{ AddedOptions []render.Option }

func (r *spyRenderer) Render(io.Writer, schema.Notebook) error { return nil }

func (r *spyRenderer) AddOptions(opts ...render.Option) {
	r.AddedOptions = append(r.AddedOptions, opts...)
}

// mockExtension extends Notebook's renderer with options.
type mockExtension struct {
	options []render.Option
}

var _ nb.Extension = (*mockExtension)(nil)

func (ext *mockExtension) Extend(n *nb.Notebook) {
	n.Renderer().AddOptions(ext.options...)
}
