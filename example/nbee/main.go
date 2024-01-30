package main

import (
	"bytes"
	_ "embed"
	"flag"
	"io"
	"log"
	"os"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/bevzzz/nb"
	synth "github.com/bevzzz/nb-synth"
	"github.com/bevzzz/nb/extension"
	"github.com/bevzzz/nb/extension/adapter"
	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/render/html"
	jupyter "github.com/bevzzz/nb/extension/extra/goldmark-jupyter"
	"github.com/robert-nix/ansihtml"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

var (
	file = flag.String("f", "notebook.ipynb", "Jupyter notebook file")
)

//go:embed notebook.ipynb
var defaultNotebook []byte

func main() {
	flag.Parse()
	var err error

	b := defaultNotebook
	outFile := "notebook.html"

	if f := *file; file != nil {
		if b, err = os.ReadFile(f); err != nil {
			log.Fatal(err)
		}
		outFile = strings.ReplaceAll(f, ".ipynb", ".html")
	}

	_ = os.Remove(outFile)
	out, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	if err := convert(out, b); err != nil {
		log.Fatal(err)
	}

	log.Printf("Done! %s -> %s", *file, outFile)
}

func convert(w io.Writer, b []byte) error {
	var body, css bytes.Buffer

	md := goldmark.New(
		goldmark.WithExtensions(
			jupyter.Attachments(),
			highlighting.Highlighting,
		),
	)

	c := nb.New(
		nb.WithExtensions(
			jupyter.Goldmark(md),
			synth.NewHighlighting(
				synth.WithStyle("monokailight"),
				synth.WithFormatOptions(
					chromahtml.WithLineNumbers(true),
				),
			),
			extension.NewStream(
				adapter.AnsiHtml(ansihtml.ConvertToHTML),
			),
		),
		nb.WithRenderOptions(
			render.WithCellRenderers(
				html.NewRenderer(
					html.WithCSSWriter(&css),
				),
			),
		),
	)

	err := c.Convert(&body, b)
	if err != nil {
		return err
	}

	// Write styles to the final output
	io.WriteString(w, "<html><head><meta charset=\"UTF-8\"><style>")
	io.Copy(w, &css)
	io.WriteString(w, "</style></head>")

	// Copy notebook body
	io.WriteString(w, "<body>")
	io.Copy(w, &body)
	io.WriteString(w, "</body></html>")
	return nil
}
