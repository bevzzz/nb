package v3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bevzzz/nb/decode"
	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
)

func init() {
	decode.RegisterDecoder(schema.Version{Major: 3, Minor: 0}, new(decoder))
}

// decoder decodes cell contents and metadata for nbformat v3.0.
type decoder struct{}

var _ decode.Decoder = (*decoder)(nil)

func (d *decoder) ExtractCells(data []byte) ([]json.RawMessage, error) {
	var raw struct {
		Worksheets []struct {
			Cells []json.RawMessage `json:"cells"`
		} `json:"worksheets"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	var cells []json.RawMessage
	for i := range raw.Worksheets {
		cells = append(cells, raw.Worksheets[i].Cells...)
	}
	return cells, nil
}

func (d *decoder) DecodeMeta(data []byte) (schema.NotebookMetadata, error) {
	return nil, nil
}

func (d *decoder) DecodeCell(m map[string]interface{}, data []byte, meta schema.NotebookMetadata) (schema.Cell, error) {
	var ct interface{}
	var c schema.Cell
	switch ct = m["cell_type"]; ct {
	case "markdown":
		c = &Markdown{}
	case "heading":
		c = &Heading{}
	case "raw":
		c = &Raw{}
	case "code":
		c = &Code{}
	default:
		return nil, fmt.Errorf("unknown cell type %q", ct)
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", ct, err)
	}
	return c, nil
}

// Markdown defines the schema for a "markdown" cell.
type Markdown struct {
	Source common.MultilineString `json:"source"`
}

var _ schema.Cell = (*Markdown)(nil)

func (md *Markdown) Type() schema.CellType {
	return schema.Markdown
}

func (md *Markdown) MimeType() string {
	return common.MarkdownText
}

func (md *Markdown) Text() []byte {
	return md.Source.Text()
}

// Heading is a dedicated cell type which represent a heading in a Jupyter notebook.
// This type is deprecated in the later versions and the content is stored as markdown instead.
//
// Heading cell behaves exactly like a markdown cell, decorating its source with the
// appropriate number of heading signs (#).
type Heading struct {
	Source common.MultilineString `json:"source"`
	Level  int
}

var _ schema.Cell = (*Heading)(nil)

func (h *Heading) Type() schema.CellType {
	return schema.Markdown
}

func (h *Heading) MimeType() string {
	return common.MarkdownText
}

func (h *Heading) Text() []byte {
	hashes := append(bytes.Repeat([]byte("#"), h.Level), " "...)
	return append(hashes, h.Source.Text()...)
}

// Raw defines the schema for a "raw" cell.
type Raw struct {
	Source   common.MultilineString `json:"source"`
	Metadata RawCellMetadata        `json:"metadata"`
}

var _ schema.Cell = (*Raw)(nil)

func (raw *Raw) Type() schema.CellType {
	return schema.Raw
}

func (raw *Raw) MimeType() string {
	return raw.Metadata.MimeType()
}

func (raw *Raw) Text() []byte {
	return raw.Source.Text()
}

// RawCellMetadata may specify a target conversion format.
type RawCellMetadata struct {
	Format      *string `json:"format"`
	RawMimeType *string `json:"raw_mimetype"`
}

// MimeType returns a more specific mime-type if one is provided and "text/plain" otherwise.
func (raw *RawCellMetadata) MimeType() string {
	switch {
	case raw.Format != nil:
		return *raw.Format
	case raw.RawMimeType != nil:
		return *raw.RawMimeType
	default:
		return common.PlainText
	}
}

// Code defines the schema for a "code" cell.
type Code struct {
	Source        common.MultilineString `json:"input"`
	TimesExecuted int                    `json:"prompt_number"`
	Out           []Output               `json:"outputs"`
	Lang          string                 `json:"language"`
}

var _ schema.CodeCell = (*Code)(nil)
var _ schema.Outputter = (*Code)(nil)

func (code *Code) Type() schema.CellType {
	return schema.Code
}

// FIXME: return correct mime type (add a function to common)
func (code *Code) MimeType() string {
	return "application/x-python"
}

func (code *Code) Text() []byte {
	return code.Source.Text()
}

func (code *Code) Language() string {
	return code.Lang
}

func (code *Code) ExecutionCount() int {
	return code.TimesExecuted
}

func (code *Code) Outputs() (cells []schema.Cell) {
	for i := range code.Out {
		cells = append(cells, code.Out[i].cell)
	}
	return
}

// Outputs unmarshals cell outputs into schema.Cell based on their type.
type Output struct {
	cell schema.Cell
}

func (out *Output) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("code outputs: %w", err)
	}

	var t interface{}
	var c schema.Cell
	switch t = v["output_type"]; t {
	case "stream":
		c = &StreamOutput{}
	case "display_data":
		c = &DisplayDataOutput{}
	case "pyout":
		c = &ExecuteResultOutput{}
	case "pyerr":
		c = &ErrorOutput{}
	default:
		return fmt.Errorf("unknown output type %q", t)
	}

	if err := json.Unmarshal(data, &c); err != nil {
		return fmt.Errorf("%q output: %w", t, err)
	}
	out.cell = c
	return nil
}

// StreamOutput is a plain, text-based output of the executed code.
// Depending on the stream "target", Type() can report "text/plain" (stdout) or "error" (stderr).
// The output is often decorated with ANSI-color sequences, which should be handled separately.
type StreamOutput struct {
	// Target can be stdout or stderr.
	Target string                 `json:"stream"`
	Source common.MultilineString `json:"text"`
}

var _ schema.Cell = (*StreamOutput)(nil)

func (stream *StreamOutput) Type() schema.CellType {
	return schema.Stream
}

func (stream *StreamOutput) MimeType() string {
	switch stream.Target {
	case "stdout":
		return common.Stdout
	case "stderr":
		return common.Stderr
	}
	return common.PlainText
}

func (stream *StreamOutput) Text() []byte {
	return stream.Source.Text()
}

// DisplayDataOutput are rich-format outputs generated by running the code in the parent cell.
type DisplayDataOutput struct {
	MimeBundle
	Metadata map[string]interface{} `json:"metadata"`
}

var _ schema.Cell = (*DisplayDataOutput)(nil)

func (dd *DisplayDataOutput) Type() schema.CellType {
	return schema.DisplayData
}

// MimeBundle contains rich output data keyed by mime-type.
type MimeBundle struct {
	PNG        common.MultilineString `json:"png,omitempty"`
	JPEG       common.MultilineString `json:"jpeg,omitempty"`
	HTML       common.MultilineString `json:"html,omitempty"`
	SVG        common.MultilineString `json:"svg,omitempty"`
	Javascript common.MultilineString `json:"javascript,omitempty"`
	JSON       common.MultilineString `json:"json,omitempty"`
	PDF        common.MultilineString `json:"pdf,omitempty"`
	LaTeX      common.MultilineString `json:"latex,omitempty"`
	Txt        common.MultilineString `json:"text,omitempty"`
}

var _ schema.MimeBundle = (*MimeBundle)(nil)

// MimeType returns the richer of the mime-types present in the bundle,
// and falls back to "text/plain" otherwise.
func (mb MimeBundle) MimeType() string {
	switch {
	case mb.PNG != nil:
		return "image/png"
	case mb.JPEG != nil:
		return "image/jpeg"
	case mb.HTML != nil:
		return "text/html"
	case mb.SVG != nil:
		return "image/svg+xml"
	case mb.Javascript != nil:
		return "text/javascript"
	case mb.JSON != nil:
		return "application/json"
	case mb.PDF != nil:
		return "application/pdf"
	case mb.LaTeX != nil:
		return "application/x-latex"
	}
	return common.PlainText
}

// Text returns data with the richer mime-type.
func (mb MimeBundle) Text() []byte {
	return mb.Data(mb.MimeType())
}

// Data returns mime-type-specific content if present and a nil slice otherwise.
func (mb MimeBundle) Data(mime string) []byte {
	switch mime {
	case "image/png":
		return mb.PNG.Text()
	case "image/jpeg":
		return mb.JPEG.Text()
	case "text/html":
		return mb.HTML.Text()
	case "image/svg+xml":
		return mb.SVG.Text()
	case "text/javascript":
		return mb.Javascript.Text()
	case "application/json":
		return mb.JSON.Text()
	case "application/pdf":
		return mb.PDF.Text()
	case "application/x-latex":
		return mb.LaTeX.Text()
	case common.PlainText:
		return mb.Txt.Text()
	}
	return nil
}

// PlainText returns data for "text/plain" mime-type and a nil slice otherwise.
func (mb MimeBundle) PlainText() []byte {
	return mb.Data(common.PlainText)
}

// ExecuteResultOutput is the result of executing the code in the cell.
// Its contents are identical to those of DisplayDataOutput with the addition of the execution count.
type ExecuteResultOutput struct {
	DisplayDataOutput
	TimesExecuted int `json:"prompt_number"`
}

var _ schema.Cell = (*ExecuteResultOutput)(nil)
var _ schema.ExecutionCounter = (*ExecuteResultOutput)(nil)

func (ex *ExecuteResultOutput) Type() schema.CellType {
	return schema.ExecuteResult
}

func (ex *ExecuteResultOutput) ExecutionCount() int {
	return ex.TimesExecuted
}

// ErrorOutput stores the output of a failed code execution.
type ErrorOutput struct {
	ExceptionName  string   `json:"ename"`
	ExceptionValue string   `json:"evalue"`
	Traceback      []string `json:"traceback"`
}

var _ schema.Cell = (*ErrorOutput)(nil)

func (err *ErrorOutput) Type() schema.CellType {
	return schema.Error
}

func (err *ErrorOutput) MimeType() string {
	return common.Stderr
}

func (err *ErrorOutput) Text() (txt []byte) {
	s := strings.Join(err.Traceback, "\n")
	return []byte(s)
}
