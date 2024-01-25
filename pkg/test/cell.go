// Package test provides test doubles that implement some of nb interfaces.
// Authors of nb-extension packages are encouraged to use them as they make
// for a uniform test code across different packages.
// See it's example usages in schema/**/*_test.go files.
package test

import (
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

// Markdown creates schema.Markdown cell with source s.
func Markdown(s string) schema.Cell {
	return &Cell{CellType: schema.Markdown, Mime: common.MarkdownText, Source: []byte(s)}
}

// Raw creates schema.Raw cell with source s and reported mime-type mt.
func Raw(s, mt string) schema.Cell {
	return &Cell{CellType: schema.Raw, Mime: mt, Source: []byte(s)}
}

// DisplayData creates schema.DisplayData cell with source s and reported mime-type mt.
func DisplayData(s, mt string) schema.Cell {
	return &Cell{CellType: schema.DisplayData, Mime: mt, Source: []byte(s)}
}

// ExecuteResult creates schema.ExecuteResult cell with source s, reported mime-type mt and execution count n.
func ExecuteResult(s, mt string, n int) schema.Cell {
	return &ExecuteResultOutput{
		Cell:          Cell{CellType: schema.ExecuteResult, Mime: mt, Source: []byte(s)},
		TimesExecuted: n,
	}
}

// ErrorOutput creates schema.Error cell with source s and mime-type common.Stderr.
func ErrorOutput(s string) schema.Cell {
	return &Cell{CellType: schema.Error, Mime: common.Stderr, Source: []byte(s)}
}

// Stdout creates schema.Stream cell with source s and mime-type common.Stdout.
func Stdout(s string) schema.Cell {
	return &Cell{CellType: schema.Stream, Mime: common.Stdout, Source: []byte(s)}
}

// Stderr creates schema.Stream cell with source s and mime-type common.Stderr.
func Stderr(s string) schema.Cell {
	return &Cell{CellType: schema.Stream, Mime: common.Stderr, Source: []byte(s)}
}

// Cell is a test fixture to mock schema.Cell.
type Cell struct {
	CellType schema.CellType
	Mime     string // mime-type (avoid name-clash with the interface method)
	Source   []byte
}

var _ schema.Cell = (*Cell)(nil)

func (c *Cell) Type() schema.CellType { return c.CellType }
func (c *Cell) MimeType() string      { return c.Mime }
func (c *Cell) Text() []byte          { return c.Source }

// CodeCell is a test fixture to mock schema.CodeCell.
// Use cases which only require schema.Cell, should create &test.Cell{CT: schema.Code} instead.
type CodeCell struct {
	Cell
	Lang          string
	TimesExecuted int
	Out           []schema.Cell
}

var _ schema.CodeCell = (*CodeCell)(nil)

func (code *CodeCell) Language() string       { return code.Lang }
func (code *CodeCell) ExecutionCount() int    { return code.TimesExecuted }
func (code *CodeCell) Outputs() []schema.Cell { return code.Out }

// ExecuteResultOutput is a test fixture to mock cell outputs with ExecuteResult type.
type ExecuteResultOutput struct {
	Cell
	TimesExecuted int
}

var _ schema.Cell = (*ExecuteResultOutput)(nil)
var _ interface{ ExecutionCount() int } = (*ExecuteResultOutput)(nil)

func (ex *ExecuteResultOutput) ExecutionCount() int { return ex.TimesExecuted }

// Notebook wraps a slice of cells into a simple schema.Notebook implementation.
func Notebook(cs ...schema.Cell) schema.Notebook {
	return cells(cs)
}

// cells implements schema.Notebook for a slice of cells.
type cells []schema.Cell

var _ schema.Notebook = (*cells)(nil)

func (n cells) Version() (v schema.Version) { return }

func (n cells) Cells() []schema.Cell { return n }

// WithAttachments creates a cell that has an attachment.
//
// The underlying test implementation for schema.MimeBundle accesses
// its keys in a random order and should always be created with 1 element only
// to keep test outcomes stable and predictable.
//
// Example:
//
//	test.WithAttachments(
//		test.Markdown("![img](attachment:photo:png)"),
//		"photo.png",
//		map[string]interface{"image/png": "base64-encoded-image"}
//	)
func WithAttachment(c schema.Cell, filename string, mimebundle map[string]interface{}) interface {
	schema.Cell
	schema.HasAttachments
} {
	return &struct {
		schema.Cell
		schema.HasAttachments
	}{
		Cell: c,
		HasAttachments: &cellAttachment{
			filename: filename,
			mb:       mimebundle,
		},
	}
}

// cellWithAttachment fakes a single cell attachment.
type cellAttachment struct {
	filename string
	mb       mimebundle
}

var _ schema.HasAttachments = (*cellAttachment)(nil)
var _ schema.Attachments = (*cellAttachment)(nil)

func (c *cellAttachment) Attachments() schema.Attachments {
	return c
}

// MimeBundle returns the underlying mime-bundle if the filename matches.
func (c *cellAttachment) MimeBundle(filename string) schema.MimeBundle {
	if filename != c.filename {
		return nil
	}
	return c.mb
}

// mimebundle is a mock implementation of schema.MimeBundle, which always
// returns the mime-type and content of its first (random access) element.
// It does not differentiate between "richer" mime-types and should not be
// created with more than one entry to keep the tests stable and reproducible.
type mimebundle map[string]interface{}

var _ schema.MimeBundle = new(mimebundle)

func (mb mimebundle) MimeType() string {
	for mt := range mb {
		return mt
	}
	return common.PlainText
}

func (mb mimebundle) Text() []byte {
	if txt, ok := mb[mb.MimeType()]; ok {
		switch v := txt.(type) {
		case []byte:
			return v
		case string:
			return []byte(v)
		}
	}
	return nil
}

func (mb mimebundle) PlainText() []byte {
	if mb.MimeType() == common.PlainText {
		return mb.Text()
	}
	return nil
}
