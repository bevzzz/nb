package html_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	stdhtml "golang.org/x/net/html"

	"github.com/stretchr/testify/require"

	"github.com/bevzzz/nb/internal/test"
	"github.com/bevzzz/nb/render/html"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

func noopRender(w io.Writer, c schema.Cell) error { return nil }

func TestWrapper_Wrap(t *testing.T) {
	for _, tt := range []struct {
		name string
		cell schema.Cell
		want *node
	}{
		{
			name: "markdown cell has jp-MarkdownCell class",
			cell: test.Markdown(""),
			want: &node{
				tag: "div",
				attr: map[string][]string{
					"class": {
						"jp-MarkdownCell",
						"jp-Cell",
						"jp-Notebook-cell",
					},
				},
			},
		},
		{
			name: "code cell has jp-CodeCell class",
			cell: &test.Cell{CellType: schema.Code},
			want: &node{
				tag: "div",
				attr: map[string][]string{
					"class": {
						"jp-CodeCell",
						"jp-Cell",
						"jp-Notebook-cell",
					},
				},
			},
		},
		{
			name: "raw cell has jp-RawCell class",
			cell: test.Raw("", common.PlainText),
			want: &node{
				tag: "div",
				attr: map[string][]string{
					"class": {
						"jp-RawCell",
						"jp-Cell",
						"jp-Notebook-cell",
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var w html.Wrapper
			var buf bytes.Buffer

			// Act
			err := w.Wrap(&buf, tt.cell, noopRender)
			require.NoError(t, err)

			// Assert
			checkDOM(t, &buf, tt.want)
		})
	}
}

func TestWrapper_WrapInput(t *testing.T) {
	// Common elements
	collapser := func() *node {
		return &node{
			tag: "div",
			attr: map[string][]string{
				"class": {"jp-Collapser", "jp-InputCollapser", "jp-Cell-inputCollapser"},
			},
			content: " ",
		}
	}

	prompt := func(s string) *node {
		n := node{
			tag: "div",
			attr: map[string][]string{
				"class": {"jp-InputPrompt", "jp-InputArea-prompt"},
			},
		}
		if s != "" {
			n.content = "In\u00a0[" + s + "]:"
		}
		return &n
	}

	for _, tt := range []struct {
		name string
		cell schema.Cell
		want *node
	}{
		{
			name: "markdown input",
			cell: test.Markdown(""),
			want: &node{
				tag: "div",
				attr: map[string][]string{
					"class":    {"jp-Cell-inputWrapper"},
					"tabindex": {"0"},
				},
				children: []*node{
					collapser(),
					{
						tag: "div",
						attr: map[string][]string{
							"class": {"jp-InputArea", "jp-Cell-inputArea"},
						},
						children: []*node{
							prompt(""),
							{
								tag: "div",
								attr: map[string][]string{
									"class": {
										"jp-RenderedMarkdown",
										"jp-MarkdownOutput",
										"jp-RenderedHTMLCommon",
									},
									"data-mime-type": {common.MarkdownText},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "raw input",
			cell: test.Raw("", common.PlainText),
			want: &node{
				tag: "div",
				attr: map[string][]string{
					"class":    {"jp-Cell-inputWrapper"},
					"tabindex": {"0"},
				},
				children: []*node{
					collapser(),
					{
						tag: "div",
						attr: map[string][]string{
							"class": {"jp-InputArea", "jp-Cell-inputArea"},
						},
						children: []*node{
							prompt(""),
						},
					},
				},
			},
		},
		{
			name: "code cell has a div additional classes and a non-empty prompt",
			cell: &test.CodeCell{
				Cell:          test.Cell{CellType: schema.Code},
				TimesExecuted: 10,
			},
			want: &node{
				tag: "div",
				attr: map[string][]string{
					"class":    {"jp-Cell-inputWrapper"},
					"tabindex": {"0"},
				},
				children: []*node{
					collapser(),
					{
						tag: "div",
						attr: map[string][]string{
							"class": {"jp-InputArea", "jp-Cell-inputArea"},
						},
						children: []*node{
							prompt("10"),
							{
								tag: "div",
								attr: map[string][]string{
									"class": {
										"jp-CodeMirrorEditor",
										"jp-Editor",
										"jp-InputArea-editor",
									},
									"data-type": {"inline"},
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var buf bytes.Buffer
			var w html.Wrapper

			// Act
			err := w.WrapInput(&buf, tt.cell, noopRender)
			require.NoError(t, err)

			// Assert
			checkDOM(t, &buf, tt.want)
		})
	}
}

func TestWrapper_WrapOutput(t *testing.T) {
	// Common elements
	collapser := func() *node {
		return &node{
			tag: "div",
			attr: map[string][]string{
				"class": {"jp-Collapser", "jp-OutputCollapser", "jp-Cell-outputCollapser"},
			},
			content: " ",
		}
	}

	prompt := func(s string) *node {
		n := node{
			tag: "div",
			attr: map[string][]string{
				"class": {"jp-OutputPrompt", "jp-OutputArea-prompt"},
			},
		}
		if s != "" {
			n.content = "Out\u00a0[" + s + "]:"
		}
		return &n
	}

	// creates parent nodes from [OutputWrapper -> [collapser + OutputArea]].
	outputArea := func(children []*node) *node {
		return &node{
			tag: "div",
			attr: map[string][]string{
				"class": {"jp-Cell-outputWrapper"},
			},
			children: []*node{
				collapser(),
				{
					tag: "div",
					attr: map[string][]string{
						"class": {"jp-OutputArea", "jp-Cell-outputArea"},
					},
					children: children,
				},
			},
		}
	}

	for _, tt := range []struct {
		name string
		out  []schema.Cell
		want *node
	}{
		{
			name: "stream output to stdout",
			out: []schema.Cell{
				test.Stdout(""),
			},
			want: outputArea([]*node{
				{
					tag: "div",
					attr: map[string][]string{
						// needs "-executeResult" if mime-type is "text/*"
						"class": {"jp-OutputArea-child", "jp-OutputArea-executeResult"},
					},
					children: []*node{
						prompt(""),
						{
							tag: "div",
							attr: map[string][]string{
								"class":          {"jp-OutputArea-output", "jp-RenderedText"},
								"data-mime-type": {common.PlainText},
							},
						},
					},
				},
			}),
		},
		{
			name: "stream output to stderr",
			out: []schema.Cell{
				test.Stderr(""),
			},
			want: outputArea([]*node{
				{
					tag: "div",
					attr: map[string][]string{
						// needs "-executeResult" if mime-type is "text/*"
						"class": {"jp-OutputArea-child", "jp-OutputArea-executeResult"},
					},
					children: []*node{
						prompt(""),
						{
							tag: "div",
							attr: map[string][]string{
								"class":          {"jp-OutputArea-output", "jp-RenderedText"},
								"data-mime-type": {common.PlainText},
							},
						},
					},
				},
			}),
		},
		{
			name: "error output",
			out: []schema.Cell{
				test.ErrorOutput(""),
			},
			want: outputArea([]*node{
				{
					tag: "div",
					attr: map[string][]string{
						"class": {"jp-OutputArea-child"},
					},
					children: []*node{
						prompt(""),
						{
							tag: "div",
							attr: map[string][]string{
								"class":          {"jp-OutputArea-output", "jp-RenderedText"},
								"data-mime-type": {common.Stderr},
							},
						},
					},
				},
			}),
		},
		{
			name: "display data image/png",
			out: []schema.Cell{
				test.DisplayData("base64-encoded-image", "image/png"),
			},
			want: outputArea([]*node{
				{
					tag: "div",
					attr: map[string][]string{
						"class": {"jp-OutputArea-child"},
					},
					children: []*node{
						prompt(""),
						{
							tag: "div",
							attr: map[string][]string{
								"class":          {"jp-OutputArea-output", "jp-RenderedImage"},
								"data-mime-type": {"image/png"},
							},
						},
					},
				},
			}),
		},
		{
			name: "display data image/jpeg",
			out: []schema.Cell{
				test.DisplayData("base64-encoded-image", "image/jpeg"),
			},
			want: outputArea([]*node{
				{
					tag: "div",
					attr: map[string][]string{
						"class": {"jp-OutputArea-child"},
					},
					children: []*node{
						prompt(""),
						{
							tag: "div",
							attr: map[string][]string{
								"class":          {"jp-OutputArea-output", "jp-RenderedImage"},
								"data-mime-type": {"image/jpeg"},
							},
						},
					},
				},
			}),
		},
		{
			name: "execute result text/html",
			out: []schema.Cell{
				test.ExecuteResult(`<img src="https://images.unsplash.com/photo" height="300"/>`, "text/html", 10),
			},
			want: outputArea([]*node{
				{
					tag: "div",
					attr: map[string][]string{
						// needs "-executeResult" if mime-type is "text/*"
						"class": {"jp-OutputArea-child", "jp-OutputArea-executeResult"},
					},
					children: []*node{
						prompt("10"),
						{
							tag: "div",
							attr: map[string][]string{
								"class":          {"jp-OutputArea-output", "jp-OutputArea-executeResult", "jp-RenderedHTMLCommon", "jp-RenderedHTML"},
								"data-mime-type": {"text/html"},
							},
						},
					},
				},
			}),
		},
		{
			name: "execute result application/json",
			out: []schema.Cell{
				test.ExecuteResult(`{"one":1,"two":2}`, "application/json", 10),
			},
			want: outputArea([]*node{
				{
					tag: "div",
					attr: map[string][]string{
						// needs "-executeResult" if mime-type is "application/json"
						"class": {"jp-OutputArea-child", "jp-OutputArea-executeResult"},
					},
					children: []*node{
						prompt("10"),
						{
							tag: "div",
							attr: map[string][]string{
								"class":          {"jp-OutputArea-output", "jp-OutputArea-executeResult", "jp-RenderedText"},
								"data-mime-type": {"application/json"},
							},
						},
					},
				},
			}),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var w html.Wrapper
			var buf bytes.Buffer

			// Act
			err := w.WrapOutput(&buf, outputs(tt.out), noopRender)
			require.NoError(t, err)

			// Assert
			checkDOM(t, &buf, tt.want)
		})
	}
}

// checkDOM parses the HTML from r and validates it against the target tree.
// Pass nil as the expectation to check for an empty HTML output.
func checkDOM(tb testing.TB, r io.Reader, want *node) {
	tb.Helper()

	// Duplicate input byte stream to print in case of an error.
	var buf bytes.Buffer
	tee := io.TeeReader(r, &buf)

	dom, err := stdhtml.Parse(tee)
	if err != nil {
		tb.Fatal(err)
	}

	// html.Parse may add missing elements to ensure a valid tree, but we don't want to validate those.
	body := findFirst(dom, "body")
	if body == nil {
		tb.Fatal("invalid html: has no <body>")
	}

	got := body.FirstChild
	switch {
	case want == nil && got == nil:
		return
	case want == nil && got != nil:
		tb.Fatalf("expected empty html, got %q", got.Data)
	case want != nil && got == nil:
		tb.Fatalf("expected <%s>, got empty html", want.tag)
	}

	if diff := want.Compare(body.FirstChild); diff.Err() != nil {
		tb.Errorf("%v\n\n%s", diff.Err(), diff.Frames())
	}
}

// htmlDiff tracks the position in the tree in which the validation error occurred.
type htmlDiff struct {
	err    error
	frames []*stdhtml.Node
}

// Frames returns a user-friendly "traceback" of the checked nodes.
func (e *htmlDiff) Frames() string {
	var w strings.Builder
	indent := func(i int) {
		ind := strings.Repeat("\t", i)
		if i > 0 {
			w.WriteString("\n")
		}

		// Add -> to the last line and remove one tab
		if i == len(e.frames)-1 {
			if i > 0 {
				ind = ind[:i-1]
			}
			ind += "-> "
		}
		w.WriteString(ind)
	}

	for i, n := range e.frames {
		indent(i)
		// Only print the root element of the current node.
		stdhtml.Render(&w, dropChildElements(n))
	}
	return w.String()
}

// AddFrame appends a node to be printed in the error traceback.
func (e *htmlDiff) AddFrame(n *stdhtml.Node) {
	e.frames = append(e.frames, n)
}

// SetErr provides the semantics of fmt.Errorf to set the error's status.
func (e *htmlDiff) SetErr(format string, args ...interface{}) {
	e.err = fmt.Errorf(format, args...)
}

func (e *htmlDiff) SetMissing(nodes []*stdhtml.Node, format string, args ...interface{}) {

}

func (e *htmlDiff) Err() error {
	return e.err
}

// node describes the desired DOM structure.
type node struct {
	tag      string              // div, span, etc
	attr     map[string][]string // attributes such as class, height, tabindex
	content  string              // textual content of the HTML element e.g. "Hi, mom!" in <p>Hi, mom!</p>
	children []*node             // child nodes (only html.ElementNodes)
}

// toHTML returns an html.Node with the node's attributes and the *textual* content.
func (n *node) toHTML() *stdhtml.Node {
	var attr []stdhtml.Attribute
	for k, v := range n.attr {
		attr = append(attr, stdhtml.Attribute{
			Key: k,
			Val: strings.Join(v, " "),
		})
	}

	var content *stdhtml.Node
	if n.content != "" {
		content = &stdhtml.Node{
			Type: stdhtml.TextNode,
			Data: n.content,
		}
	}

	return &stdhtml.Node{
		Type:       stdhtml.ElementNode,
		Data:       n.tag,
		Attr:       attr,
		FirstChild: content,
	}
}

// Compare traverses the parse tree and compares it to its own structure.
// Commment nodes are ignored.
func (n *node) Compare(other *stdhtml.Node) htmlDiff {
	var diff htmlDiff
	n.cmp(other, &diff)
	return diff
}

// cmp updates the validation status as it walks the tree.
func (n *node) cmp(other *stdhtml.Node, status *htmlDiff) {
	status.AddFrame(other)

	switch other.Type {
	case stdhtml.TextNode: // this accounts for cases where TextNode is the top-most element
		if got := trim1(other.Data, "\n"); got != n.content {
			status.SetErr("want content %q, got %q", n.content, other.Data)
		}
		return // don't look further, text nodes have neither childern no attributes
	case stdhtml.ElementNode:
		if got, want := other.Data, n.tag; want != got {
			status.SetErr("wrong tag: want <%s>, got <%s>", want, got)
			return
		}
	}

	for key, values := range n.attr {

		var attrval string
		var got []string
		for _, attr := range other.Attr {
			if attr.Key == key {
				attrval = attr.Val
				got = strings.Split(attr.Val, " ")
				break
			}
		}

		// Check if the attribute matches as a whole, e.g. <img src=""> property.
		if len(values) == 1 && values[0] == attrval {
			continue
		}

		for _, v := range values {
			if !contains(got, v) {
				status.SetErr("missing or incorrect %s: want %q, got %q", key, v, attrval)
				return
			}
		}
	}

	var i int
	var c *stdhtml.Node
	l := len(n.children)
	for c = other.FirstChild; c != nil && i <= l; c = c.NextSibling {
		switch c.Type {
		case stdhtml.TextNode:
			if got := trim1(c.Data, "\n"); got != n.content {
				status.SetErr("want content %q, got %q", n.content, c.Data)
				return
			}
		case stdhtml.ElementNode:
			n.children[i].cmp(c, status)
			if status.Err() != nil {
				return
			}
			i++
		}
	}

	if i < l {
		var missing strings.Builder
		for j, m := range n.children[i:l] {
			if j > 0 {
				missing.WriteString("\n")
			}
			stdhtml.Render(&missing, m.toHTML())
		}
		status.SetErr("missing %d node(-s) after ->:\n%s", l-i, &missing)
	} else if c != nil && c.NextSibling != nil {
		// The last node we checked has more children, but we did not expect them.
		var extra strings.Builder
		stdhtml.Render(&extra, c.NextSibling)
		status.AddFrame(c.NextSibling)
		status.SetErr("unexpected node (at ->)")
	}
}

// findFirst fast-forwards to the first occurrence of the <target> node and returns it.
// If the node is not in the tree, it will return nil.
func findFirst(n *stdhtml.Node, target string) *stdhtml.Node {
	if n.Type == stdhtml.ElementNode && n.Data == target {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == stdhtml.ElementNode && c.NextSibling == nil {
			return findFirst(c, target)
		}
	}
	return nil
}

// dropChildElements removes the first child element if it is not a TextElement.
// Add modifications are done on the copy, so the original node is not modified.
func dropChildElements(n *stdhtml.Node) *stdhtml.Node {
	cp := *n
	if cp.FirstChild != nil && cp.FirstChild.Type != stdhtml.TextNode {
		cp.FirstChild = nil
	}
	cp.NextSibling = nil
	return &cp
}

// outputs implements schema.Outputter.
type outputs []schema.Cell

var _ schema.Outputter = (*outputs)(nil)

func (o outputs) Outputs() []schema.Cell { return o }

// contains returns true on the first occurrence of v in s,
// and false if not present.
func contains(s []string, v string) bool {
	for i := range s {
		if v == s[i] {
			return true
		}
	}
	return false
}

// trim1 trims 1 leading and 1 trailing character in cutset.
func trim1(s string, cutset string) string {
	s = strings.TrimPrefix(s, cutset)
	s = strings.TrimSuffix(s, cutset)
	return s
}
