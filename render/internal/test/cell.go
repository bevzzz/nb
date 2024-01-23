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
