package schema

type CellTypeMixed string

const (
	// TODO: drop these

	PlainTextCellType CellTypeMixed = "text/plain"
	MarkdownCellType  CellTypeMixed = "text/markdown"
	HTML              CellTypeMixed = "text/html"
	PNG               CellTypeMixed = "image/png"
	JPEG              CellTypeMixed = "image/jpeg"
	JSON              CellTypeMixed = "application/json"

	// Internal cell types

	CodeCellType   CellTypeMixed = "code"
	StdoutCellType CellTypeMixed = "stdout"
	StderrCellType CellTypeMixed = "stderr"
)
