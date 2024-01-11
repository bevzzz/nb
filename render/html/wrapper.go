package html

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/bevzzz/nb/render"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
	"golang.org/x/net/html"
)

/*
TODO:
	- make class prefixes configurable (probably on the html.Renderer level).
	- refactor to use tagger
	- add WrapAll / WrapNotebook that would add a <div class="jp-Notebook"> (then there's no need for WriterOnce)
*/

// Wrapper wraps cells in the HTML produced by the original Jupyter's nbconvert.
type Wrapper struct {
	Config
}

func (wr *Wrapper) Wrap(w io.Writer, cell schema.Cell, render render.RenderCellFunc) error {
	if wr.CSSWriter != nil {
		wr.CSSWriter.Write(jupyterCSS)
	}

	var ct string
	switch cell.CellType() {
	case schema.Markdown:
		ct = "jp-MarkdownCell"
	case schema.Code:
		ct = "jp-CodeCell"
		// TODO: if no outputs, add jp-mod-noOutputs class
	case schema.Raw:
		ct = "jp-RawCell"
	}

	div.Open(w, attributes{"class": {"jp-Cell", ct, "jp-Notebook-cell"}})
	render(w, cell)
	div.Close(w)
	return nil
}

func (wr *Wrapper) WrapInput(w io.Writer, cell schema.Cell, render render.RenderCellFunc) error {
	div.Open(w, attributes{
		"class":    {"jp-Cell-inputWrapper"},
		"tabindex": {0}})

	div.Open(w, attributes{"class": {"jp-Collapser", "jp-InputCollapser", "jp-Cell-inputCollapser"}})
	io.WriteString(w, " ")
	div.Close(w)

	// TODO: add collapser-child <div class="jp-Collapser-child"></div> and collapsing functionality
	// Pure CSS Collapsible: https://www.digitalocean.com/community/tutorials/css-collapsible

	div.Open(w, attributes{"class": {"jp-InputArea", "jp-Cell-inputArea"}})

	// Prompt In:[1]
	div.Open(w, attributes{"class": {"jp-InputPrompt", "jp-InputArea-prompt"}})
	if ex, ok := cell.(interface{ ExecutionCount() int }); ok {
		fmt.Fprintf(w, "In\u00a0[%d]:", ex.ExecutionCount())
	}
	div.Close(w)

	isCode := cell.CellType() == schema.Code
	isMd := cell.CellType() == schema.Markdown
	if isCode {
		div.Open(w, attributes{
			"class": {
				"jp-CodeMirrorEditor",
				"jp-Editor",
				"jp-InputArea-editor",
			},
			"data-type": {"inline"},
		})
	} else if isMd {
		div.Open(w, attributes{
			"class": {
				"jp-RenderedMarkdown",
				"jp-MarkdownOutput",
				"jp-RenderedHTMLCommon",
			},
			"data-mime-type": {common.MarkdownText},
		})
	}

	// Cell itself
	_ = render(w, cell)

	if isCode || isMd {
		div.Close(w)
	}

	div.Close(w)
	div.Close(w)
	return nil
}

func (wr *Wrapper) WrapOutput(w io.Writer, cell schema.Outputter, render render.RenderCellFunc) error {
	div.Open(w, attributes{"class": {"jp-Cell-outputWrapper"}})
	div.OpenClose(w, attributes{"class": {"jp-Collapser", "jp-OutputCollapser", "jp-Cell-outputCollapser"}})
	div.Open(w, attributes{"class": {"jp-OutputArea jp-Cell-outputArea"}})

	// TODO: see how application/json would be handled
	// TODO: jp-RenderedJavaScript is a thing and so is jp-RenderedLatex (but I don't think we need to do anything about the latter)

	var child bool
	var childClass = "jp-OutputArea-child"
	var datamimetype string
	var outputtypeclass string

	if outs := cell.Outputs(); len(outs) > 0 {
		datamimetype = outs[0].MimeType()
		first := outs[0]

		switch first.CellType() {
		case schema.ExecuteResult:
			outputtypeclass = "jp-OutputArea-executeResult"
			child = true
		case schema.Error:
			child = true
		case schema.Stream:
			datamimetype = common.PlainText
			child = true
		}
	}

	var renderedClass string
	if strings.HasPrefix(datamimetype, "text/") || datamimetype == "application/json" {
		childClass += " jp-OutputArea-executeResult"
		renderedClass = "jp-RenderedText"
		if datamimetype == "text/html" {
			renderedClass = "jp-RenderedHTMLCommon jp-RenderedHTML"
		}
	} else if strings.HasPrefix(datamimetype, "image/") {
		renderedClass = "jp-RenderedImage"
		child = true
	} else if datamimetype == "application/vnd.jupyter.stderr" {
		renderedClass = "jp-RenderedText"
	}

	// Looks like this will always wrap the whole output area!
	if child {
		div.Open(w, attributes{"class": {childClass}})
	}

	div.Open(w, attributes{"class": {"jp-OutputPrompt", "jp-OutputArea-prompt"}})
	for _, out := range cell.Outputs() {
		if ex, ok := out.(interface{ ExecutionCount() int }); ok {
			fmt.Fprintf(w, "Out\u00a0[%d]:", ex.ExecutionCount())
			break
		}
	}
	div.Close(w)

	div.Open(w, attributes{
		"class":          {renderedClass, "jp-OutputArea-output", outputtypeclass},
		"data-mime-type": {datamimetype},
	})
	for _, out := range cell.Outputs() {
		_ = render(w, out)
	}
	div.Close(w)

	if child {
		div.Close(w)
	}

	div.Close(w)
	div.Close(w)
	return nil
}

const (
	div tag = "div"
)

type tag string

// Open the tag with the attributes, e.g. <div class="container" checked>.
func (t tag) Open(w io.Writer, attrs attributes) {
	t._open(w, attrs, true)
}

func (t tag) _open(w io.Writer, attrs attributes, newline bool) {
	io.WriteString(w, "<")
	io.WriteString(w, string(t))
	attrs.WriteTo(w)
	io.WriteString(w, ">")
	if newline {
		io.WriteString(w, "\n")
	}
}

func (t tag) Close(w io.Writer) {
	fmt.Fprintf(w, "</%s>\n", t)
}

func (t tag) OpenClose(w io.Writer, attrs attributes) {
	t._open(w, attrs, false)
	t.Close(w)
}

// Empty writes the attributes in an empty-element tag, e.g. <div class="container" />.
func (t tag) Empty(w io.Writer, attrs attributes) {
	io.WriteString(w, "<")
	io.WriteString(w, string(t))
	attrs.WriteTo(w)
	io.WriteString(w, " />")
}

type tagger struct {
	opened []tag
}

// Open opens the tag with the attributes.
func (t *tagger) Open(tag tag, w io.Writer, attr attributes) {
	tag.Open(w, attr)
	t.opened = append(t.opened, tag)
}

// Close closes all opened tags in reverse order.
func (t *tagger) Close(w io.Writer) {
	l := len(t.opened)
	if l == 0 {
		return
	}
	for i := l - 1; i >= 0; i-- {
		t.opened[i].Close(w)
	}
}

type attributes map[string][]interface{}

// WriteTo writes values of each attribute in a space-separated list, e.g. class="container box jp-NotebookCell".
// TODO: refactor to use html.Attribute from the beginning
func (attrs attributes) WriteTo(w io.Writer) (n64 int64, err error) {
	type Attribute struct {
		html.Attribute
		IsBool bool
	}

	var sorted []Attribute
	for k, values := range attrs {
		var v string
		attr := Attribute{
			Attribute: html.Attribute{Key: k},
		}

		if len(values) == 1 {
			if _, isBool := values[0].(bool); isBool {
				attr.IsBool = true
				sorted = append(sorted, attr)
				continue
			}
		}

		for i := range values {
			if i > 0 {
				v += " "
			}
			v += fmt.Sprintf("%v", values[i])
		}

		attr.Val = v
		sorted = append(sorted, attr)
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	// class=""
	for _, attr := range sorted {
		s := " "
		if attr.IsBool {
			s += attr.Key
			continue
		}
		s += fmt.Sprintf("%s=\"%s\"", attr.Key, attr.Val)

		var n int
		n, err = io.WriteString(w, s)
		if err != nil {
			return
		}
		n64 += int64(n)
	}
	return
}

// WriterOnce writes to the writer once only.
// TODO: move to util
type WriterOnce struct {
	w    io.Writer
	once sync.Once
}

var _ io.Writer = (*WriterOnce)(nil)

func (w *WriterOnce) Write(p []byte) (n int, err error) {
	w.once.Do(func() {
		n, err = w.w.Write(p)
	})
	return
}
