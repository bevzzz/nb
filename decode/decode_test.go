package decode_test

import (
	"bytes"
	"testing"

	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
	_ "github.com/bevzzz/nb/schema/v4"

	"github.com/bevzzz/nb/decode"
	"github.com/stretchr/testify/require"
)

// Cell combines a set of common expectations for a cell.
type Cell struct {
	Type     schema.CellType
	MimeType string
	Text     []byte
}

func TestDecodeBytes(t *testing.T) {
	t.Run("notebook", func(t *testing.T) {
		for _, tt := range []struct {
			name   string
			json   string
			nCells int
		}{
			{
				name: "v4.4",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "markdown", "metadata": {}, "source": []},
						{"cell_type": "markdown", "metadata": {}, "source": []}
					]
				}`,
				nCells: 2,
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				nb, err := decode.Bytes([]byte(tt.json))
				require.NoError(t, err)

				got := nb.Cells()
				require.Lenf(t, got, tt.nCells, "expected %d cells", tt.nCells)
				for i := range got {
					require.NotNil(t, got[i], "cell #%d is nil", i+1)
				}
			})
		}
	})

	t.Run("markdown cells", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			json string
			want Cell
		}{
			{
				name: "v4.4",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "markdown", "metadata": {}, "source": ["Join", " ", "me"]}
					]
				}`,
				want: Cell{
					Type:     schema.Markdown,
					MimeType: common.MarkdownText,
					Text:     []byte("Join me"),
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				nb, err := decode.Bytes([]byte(tt.json))
				require.NoError(t, err)

				got := nb.Cells()
				require.Len(t, got, 1, "expected 1 cell")

				checkCell(t, got[0], tt.want)
			})
		}
	})

	t.Run("raw cells", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			json string
			want Cell
		}{
			{
				name: "v4.4: no explicit mime-type",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "raw", "source": ["Plain as the nose on your face"]}
					]
				}`,
				want: Cell{
					Type:     schema.Raw,
					MimeType: common.PlainText,
					Text:     []byte("Plain as the nose on your face"),
				},
			},
			{
				name: "v4.4: metadata.format has specific mime-type",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "raw", "metadata": {"format": "text/html"}, "source": ["<p>Hi, mom!</p>"]}
					]
				}`,
				want: Cell{
					Type:     schema.Raw,
					MimeType: "text/html",
					Text:     []byte("<p>Hi, mom!</p>"),
				},
			},
			{
				name: "v4.4: metadata.raw_mimetype has specific mime-type",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "raw", "metadata": {"raw_mimetype": "application/x-latex"}, "source": ["$$"]}
					]
				}`,
				want: Cell{
					Type:     schema.Raw,
					MimeType: "application/x-latex",
					Text:     []byte("$$"),
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				nb, err := decode.Bytes([]byte(tt.json))
				require.NoError(t, err)

				got := nb.Cells()
				require.Len(t, got, 1, "expected 1 cell")

				checkCell(t, got[0], tt.want)
			})
		}
	})

	t.Run("code cells", func(t *testing.T) {
		type outcome struct {
			Cell
			Language       string
			ExecutionCount int
			OutputLen      int
		}

		for _, tt := range []struct {
			name string
			json string
			want outcome
		}{
			{
				name: "v4.4",
				json: `{
					"nbformat": 4, "nbformat_minor": 4,
					"metadata": {"language_info": {"name": "javascript"}},
					"cells": [
						{
							"cell_type": "code", "execution_count": 5,
							"source": ["print('Hi, mom!')"],  "outputs": [
								{"output_type": "stream"}, {"output_type": "stream"}
							]
						}
					]
				}`,
				want: outcome{
					Cell: Cell{
						Type:     schema.Code,
						MimeType: "application/x-python", // FIXME: expect language-specific mime-type
						Text:     []byte("print('Hi, mom!')"),
					},
					Language:       "javascript",
					ExecutionCount: 5,
					OutputLen:      2,
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				nb, err := decode.Bytes([]byte(tt.json))
				require.NoError(t, err)

				cell := toCodeCell(t, nb.Cells()[0])

				require.Equal(t, tt.want.Language, cell.Language(), "language")
				require.Len(t, cell.Outputs(), tt.want.OutputLen, "number of outputs")
				checkCell(t, cell, tt.want.Cell)
			})
		}
	})

	t.Run("code cell outputs", func(t *testing.T) {
		type output struct {
			Cell
			ExecutionCount int // accept zero value for cells that don't implement interface{ ExecutionCount() int}
		}

		for _, tt := range []struct {
			name string
			json string
			want []output
		}{
			{
				name: "v4.4: stream output to stdout",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {},
					"cells": [
						{"cell_type": "code", "outputs": [
							{
								"output_type": "stream", "name": "stdout",
								"text": ["$> ls\n", ".\n", "..\n", "nb/"]
							}
						]}
					]
				}`,
				want: []output{
					{Cell: Cell{
						Type:     schema.Stream,
						MimeType: common.Stdout,
						Text:     []byte("$> ls\n.\n..\nnb/"),
					}},
				},
			},
			{
				name: "v4.4: stream output to stderr",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {},
					"cells": [
						{"cell_type": "code", "outputs": [
							{
								"output_type": "stream", "name": "stderr",
								"text": ["KeyError: ", "dict['unknown key']"]
							}
						]}
					]
				}`,
				want: []output{
					{Cell: Cell{
						Type:     schema.Stream,
						MimeType: common.Stderr,
						Text:     []byte("KeyError: dict['unknown key']"),
					}},
				},
			},
			{
				name: "v4.4: stream output to unrecognized target",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {},
					"cells": [
						{"cell_type": "code", "outputs": [
							{
								"output_type": "stream", "name": "unknown",
								"text": ["print me please..."]
							}
						]}
					]
				}`,
				want: []output{
					{Cell: Cell{
						Type:     schema.Stream,
						MimeType: common.PlainText,
						Text:     []byte("print me please..."),
					}},
				},
			},
			{
				name: "v4.4: display_data output with several images and a plain text",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {},
					"cells": [
						{"cell_type": "code", "outputs": [
							{"output_type": "display_data", "metadata": {},
								"data": {
									"image/png":  "base64-encoded-png-image",
									"text/plain": "<Figure size 640x480 with 1 Axes>"
								}
							},
							{"output_type": "display_data", "metadata": {},
								"data": {
									"image/jpeg":  "base64-encoded-jpeg-image",
									"text/plain": "<Figure size 100x500 with 2 Axes>"
								}
							},
							{"output_type": "display_data", "metadata": {},
								"data": {
									"text/plain": "<Image url='https://image.com/?id=123' height=500>"
								}
							}
						]}
					]
				}`,
				want: []output{
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: "image/png",
						Text:     []byte("base64-encoded-png-image"),
					}},
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: "image/jpeg",
						Text:     []byte("base64-encoded-jpeg-image"),
					}},
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: common.PlainText,
						Text:     []byte("<Image url='https://image.com/?id=123' height=500>"),
					}},
				},
			},
			{
				name: "v4.4: execute_result output with several images and a plain text",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {},
					"cells": [
						{"cell_type": "code", "outputs": [
							{"output_type": "execute_result", "metadata": {},
								"execution_count": 13,
								"data": {
									"text/html":  "<p>Base thirteen!</p>"
								}
							},
							{"output_type": "execute_result", "metadata": {},
								"execution_count": 42,
								"data": {
									"text/plain": "<MeaningOfLife question='???'>"
								}
							}
						]}
					]
				}`,
				want: []output{
					{ExecutionCount: 13, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "text/html",
						Text:     []byte("<p>Base thirteen!</p>"),
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: common.PlainText,
						Text:     []byte("<MeaningOfLife question='???'>"),
					}},
				},
			},
			{
				name: "v4.4: error output",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {},
					"cells": [
						{"cell_type": "code", "outputs": [
							{
								"output_type": "error", "ename": "ZeroDivisionError", "evalue": "division by zero",
								"traceback": [
									"Traceback (most recent call last):",
									"\tFile \"main.py\", line 3, in <module>",
									"\t\tprint(n/0)",
									"\tZeroDivisionError: division by zero"
								]
							}
						]}
					]
				}`,
				want: []output{
					{Cell: Cell{
						Type:     schema.Error,
						MimeType: common.Stderr,
						Text:     []byte("Traceback (most recent call last):\n\tFile \"main.py\", line 3, in <module>\n\t\tprint(n/0)\n\tZeroDivisionError: division by zero"),
					}},
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				nb, err := decode.Bytes([]byte(tt.json))
				require.NoError(t, err)
				cell := toCodeCell(t, nb.Cells()[0])

				outputs := cell.Outputs()

				require.Len(t, outputs, len(tt.want), "number of outputs")
				for i := range tt.want {
					got, want := outputs[i], tt.want[i]
					checkCell(t, got, want.Cell)
				}
			})
		}
	})
}

// checkCell compares the cell's type and content to expected.
func checkCell(tb testing.TB, got schema.Cell, want Cell) {
	tb.Helper()
	require.Equalf(tb, want.Type, got.Type(), "reported cell type: want %q, got %q", want.Type, got.Type())
	require.Equal(tb, want.MimeType, got.MimeType(), "reported mime type")
	if got, want := got.Text(), want.Text; !bytes.Equal(want, got) {
		tb.Errorf("text:\n(+want) %q\n(-got) %q", want, got)
	}
}

// toCodeCell fails the test if the cell does not implement schema.CodeCell.
func toCodeCell(tb testing.TB, cell schema.Cell) schema.CodeCell {
	tb.Helper()

	if code, ok := cell.(schema.CodeCell); ok {
		return code
	}
	tb.Errorf("cell %T is not schema.CodeCell", cell)
	return nil
}
