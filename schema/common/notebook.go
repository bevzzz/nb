package common

import (
	"encoding/json"

	"github.com/bevzzz/nb/schema"
)

type Notebook struct {
	VersionMajor int               `json:"nbformat"`
	VersionMinor int               `json:"nbformat_minor"`
	Metadata     json.RawMessage   `json:"metadata"` // TODO: omitempty
}

func (n *Notebook) Version() schema.Version {
	return schema.Version{
		Major: n.VersionMajor,
		Minor: n.VersionMinor,
	}
}

const (
	PlainText    = "text/plain"
	MarkdownText = "text/markdown"
	Stdout       = "application/vnd.jupyter.stdout" // Custom mime-type for stream output to stdout.
	Stderr       = "application/vnd.jupyter.stderr" // Custom mime-type for stream output to stderr.
)
