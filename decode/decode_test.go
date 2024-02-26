package decode_test

import (
	"bytes"
	"testing"

	"github.com/bevzzz/nb/schema"
	"github.com/bevzzz/nb/schema/common"
	_ "github.com/bevzzz/nb/schema/v3"
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

type WithAttachments struct {
	Cell
	Filename string
	MimeType string
	Data     []byte
}

func (w WithAttachments) HasAttachments() bool {
	return w.Filename != ""
}

func TestDecodeBytes(t *testing.T) {
	t.Run("notebook", func(t *testing.T) {
		for _, tt := range []struct {
			name   string
			json   string
			nCells int
		}{
			{
				name: "v4.5",
				json: `{
					"nbformat": 4, "nbformat_minor": 5, "metadata": {}, "cells": [
						{"id": "a", "cell_type": "markdown", "metadata": {}, "source": []},
						{"id": "b", "cell_type": "markdown", "metadata": {}, "source": []}
					]
				}`,
				nCells: 2,
			},
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
			{
				name: "v4.3",
				json: `{
					"nbformat": 4, "nbformat_minor": 3, "metadata": {}, "cells": [
						{"cell_type": "markdown", "metadata": {}, "source": []},
						{"cell_type": "markdown", "metadata": {}, "source": []}
					]
				}`,
				nCells: 2,
			},
			{
				name: "v4.2",
				json: `{
					"nbformat": 4, "nbformat_minor": 2, "metadata": {}, "cells": [
						{"cell_type": "markdown", "metadata": {}, "source": []},
						{"cell_type": "markdown", "metadata": {}, "source": []}
					]
				}`,
				nCells: 2,
			},
			{
				name: "v4.1",
				json: `{
					"nbformat": 4, "nbformat_minor": 1, "metadata": {}, "cells": [
						{"cell_type": "markdown", "metadata": {}, "source": []},
						{"cell_type": "markdown", "metadata": {}, "source": []}
					]
				}`,
				nCells: 2,
			},
			{
				name: "v4.0",
				json: `{
					"nbformat": 4, "nbformat_minor": 0, "metadata": {}, "cells": [
						{"cell_type": "markdown", "metadata": {}, "source": []},
						{"cell_type": "markdown", "metadata": {}, "source": []}
					]
					}`,
				nCells: 2,
			},
			{
				name: "v3.0",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "markdown", "metadata": {}, "source": []},
							{"cell_type": "markdown", "metadata": {}, "source": []}
						]},
						{"cells": [
							{"cell_type": "markdown", "metadata": {}, "source": []}
						]}
					]
				}`,
				nCells: 3,
			},
			{
				name: "v2.0",
				json: `{
					"nbformat": 2, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "markdown", "metadata": {}, "source": []},
							{"cell_type": "markdown", "metadata": {}, "source": []}
						]},
						{"cells": [
							{"cell_type": "markdown", "metadata": {}, "source": []}
						]}
					]
				}`,
				nCells: 3,
			},
			{
				name: "v1.0",
				json: `{
					"nbformat": 1, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "markdown", "metadata": {}, "source": []},
							{"cell_type": "markdown", "metadata": {}, "source": []}
						]},
						{"cells": [
							{"cell_type": "markdown", "metadata": {}, "source": []}
						]}
					]
				}`,
				nCells: 3,
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
			want WithAttachments
		}{
			{
				name: "v4.4",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "markdown", "metadata": {}, "source": [
							"Look", " at ", "me: ![alt](attachment:photo.png)"
						], "attachments": {
							"photo.png": {
								"image/png": "base64-encoded-image-data"
							}
						}}
					]
				}`,
				want: WithAttachments{
					Cell: Cell{
						Type:     schema.Markdown,
						MimeType: common.MarkdownText,
						Text:     []byte("Look at me: ![alt](attachment:photo.png)"),
					},
					Filename: "photo.png",
					MimeType: "image/png",
					Data:     []byte("base64-encoded-image-data"),
				},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				nb, err := decode.Bytes([]byte(tt.json))
				require.NoError(t, err)

				got := nb.Cells()
				require.Len(t, got, 1, "expected 1 cell")

				checkCellWithAttachments(t, got[0], tt.want)
			})
		}
	})

	t.Run("raw cells", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			json string
			want WithAttachments
		}{
			{
				name: "v4.4: no explicit mime-type",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "raw", "source": ["Plain as the nose on your face"]}
					]
				}`,
				want: WithAttachments{Cell: Cell{
					Type:     schema.Raw,
					MimeType: common.PlainText,
					Text:     []byte("Plain as the nose on your face"),
				}},
			},
			{
				name: "v4.4: metadata.format has specific mime-type",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "raw", "metadata": {"format": "text/html"}, "source": ["<p>Hi, mom!</p>"]}
					]
				}`,
				want: WithAttachments{Cell: Cell{
					Type:     schema.Raw,
					MimeType: "text/html",
					Text:     []byte("<p>Hi, mom!</p>"),
				}},
			},
			{
				name: "v4.4: metadata.raw_mimetype has specific mime-type",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{"cell_type": "raw", "metadata": {"raw_mimetype": "application/x-latex"}, "source": ["$$"]}
					]
				}`,
				want: WithAttachments{Cell: Cell{
					Type:     schema.Raw,
					MimeType: "application/x-latex",
					Text:     []byte("$$"),
				}},
			},
			{
				name: "v4.4: with attachments",
				json: `{
					"nbformat": 4, "nbformat_minor": 4, "metadata": {}, "cells": [
						{
							"cell_type": "raw", "metadata": {},
							"source": ["![alt](attachment:photo.png)"], "attachments": {
								"photo.png": {
									"image/png": "base64-encoded-image-data"
								}
							}
						}
					]
				}`,
				want: WithAttachments{
					Cell: Cell{
						Type:     schema.Raw,
						MimeType: common.PlainText,
						Text:     []byte("![alt](attachment:photo.png)"),
					},
					Filename: "photo.png",
					MimeType: "image/png",
					Data:     []byte("base64-encoded-image-data"),
				},
			},
			{
				name: "v3.0: no explicit mime-type",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "raw", "source": ["sometimes you just want to rawdog sqweel"]}
						]}
					]
				}`,
				want: WithAttachments{Cell: Cell{
					Type:     schema.Raw,
					MimeType: common.PlainText,
					Text:     []byte("sometimes you just want to rawdog sqweel"),
				}},
			},
		} {
			t.Run(tt.name, func(t *testing.T) {
				nb, err := decode.Bytes([]byte(tt.json))
				require.NoError(t, err)

				got := nb.Cells()
				require.Len(t, got, 1, "expected 1 cell")

				checkCellWithAttachments(t, got[0], tt.want)
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
								{"output_type": "stream", "name": "stdout"},
								{"output_type": "stream", "name": "stderr"}
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
			{
				name: "v3.0",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{
								"cell_type": "code", "language": "javascript", "prompt_number": 5,
								"input": ["print('Hi, mom!')"],  "outputs": [
									{"output_type": "stream", "stream": "stdout"}, 
									{"output_type": "stream", "stream": "stderr"}
								]
							}
						]}
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
				name: "v3.0: stream output to stdout",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "code", "outputs": [
								{
									"output_type": "stream", "stream": "stdout",
									"text": ["$> ls\n", ".\n", "..\n", "nb/"]
								}
							]}
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
				name: "v3.0: stream output to stderr",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "code", "outputs": [
								{
									"output_type": "stream", "stream": "stderr",
									"text": ["KeyError: ", "dict['unknown key']"]
								}
							]}
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
									"image/png": "base64-encoded-png-image",
									"text/plain": "<Figure size 640x480 with 1 Axes>"
								}
							},
							{"output_type": "display_data", "metadata": {},
								"data": {
									"image/jpeg": "base64-encoded-jpeg-image",
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
				name: "v3.0: display_data output different recognized formats",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "code", "outputs": [
								{"output_type": "display_data", "metadata": {},
									"png": ["base64-encoded-png-image"],
									"text": ["<Figure size 640x480 with 1 Axes>"]
								},
								{"output_type": "display_data", "metadata": {},
									"jpeg": ["base64-encoded-jpeg-image"],
									"text": ["<Figure size 100x500 with 2 Axes>"]
								},
								{"output_type": "display_data", "metadata": {},
									"html": ["<img />"]
								},
								{"output_type": "display_data", "metadata": {},
									"svg": ["<svg />"]
								},
								{"output_type": "display_data", "metadata": {},
									"javascript": ["[,,,].length"]
								},
								{"output_type": "display_data", "metadata": {},
									"json": ["{\"foo\": \"bar\"}"]
								},
								{"output_type": "display_data", "metadata": {},
									"pdf": ["some-raw-pdf-data"]
								},
								{"output_type": "display_data", "metadata": {},
									"latex": ["c = \\sqrt{a^2 + b^2}"]
								},
								{"output_type": "display_data", "metadata": {},
									"text": ["<Image url='https://image.com/?id=123' height=500>"]
								}
							]}
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
						MimeType: "text/html",
						Text:     []byte(`<img />`),
					}},
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: "image/svg+xml",
						Text:     []byte(`<svg />`),
					}},
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: "text/javascript",
						Text:     []byte("[,,,].length"),
					}},
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: "application/json",
						Text:     []byte("{\"foo\": \"bar\"}"), // ????
					}},
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: "application/pdf",
						Text:     []byte("some-raw-pdf-data"), // ????
					}},
					{Cell: Cell{
						Type:     schema.DisplayData,
						MimeType: "application/x-latex",
						Text:     []byte("c = \\sqrt{a^2 + b^2}"), // ????
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
				name: "v3.0: pyout (execute_result) output different recognized formats",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "code", "outputs": [
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"png": ["base64-encoded-png-image"],
									"text": ["<Figure size 640x480 with 1 Axes>"]
								},
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"jpeg": ["base64-encoded-jpeg-image"],
									"text": ["<Figure size 100x500 with 2 Axes>"]
								},
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"html": ["<img />"]
								},
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"svg": ["<svg />"]
								},
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"javascript": ["[,,,].length"]
								},
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"json": ["{\"foo\": \"bar\"}"]
								},
								{"output_type": "pyout", "metadata": {},
									"pdf": ["some-raw-pdf-data"]
								},
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"latex": ["c = \\sqrt{a^2 + b^2}"]
								},
								{"output_type": "pyout", "metadata": {},
									"prompt_number": 42,
									"text": ["<Image url='https://image.com/?id=123' height=500>"]
								}
							]}
						]}
					]
				}`,
				want: []output{
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "image/png",
						Text:     []byte("base64-encoded-png-image"),
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "image/jpeg",
						Text:     []byte("base64-encoded-jpeg-image"),
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "text/html",
						Text:     []byte(`<img />`),
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "image/svg+xml",
						Text:     []byte(`<svg />`),
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "text/javascript",
						Text:     []byte("[,,,].length"),
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "application/json",
						Text:     []byte("{\"foo\": \"bar\"}"), // ????
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "application/pdf",
						Text:     []byte("some-raw-pdf-data"), // ????
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: "application/x-latex",
						Text:     []byte("c = \\sqrt{a^2 + b^2}"), // ????
					}},
					{ExecutionCount: 42, Cell: Cell{
						Type:     schema.ExecuteResult,
						MimeType: common.PlainText,
						Text:     []byte("<Image url='https://image.com/?id=123' height=500>"),
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
			{
				name: "v3.0: error output",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{"cell_type": "code", "outputs": [
								{
									"output_type": "pyerr", "ename": "ZeroDivisionError", "evalue": "division by zero",
									"traceback": [
										"Traceback (most recent call last):",
										"\tFile \"main.py\", line 3, in <module>",
										"\t\tprint(n/0)",
										"\tZeroDivisionError: division by zero"
									]
								}
							]}
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

	t.Run("heading cells", func(t *testing.T) {
		for _, tt := range []struct {
			name string
			json string
			want Cell
		}{
			{
				name: "v3.0 used to have dedicated heading cells",
				json: `{
					"nbformat": 3, "nbformat_minor": 0, "metadata": {}, "worksheets": [
						{"cells": [
							{
								"cell_type": "heading", "level": 2, 
								"source": ["Fun facts about Ronald McDonald"], "metadata": {}
							}
						]}
					]
				}`,
				want: Cell{
					Type:     schema.Markdown,
					MimeType: common.MarkdownText,
					Text:     []byte("## Fun facts about Ronald McDonald"),
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

// checkCellWithAttachments compares the cell's type, content, and attachments to expected.
func checkCellWithAttachments(tb testing.TB, got schema.Cell, want WithAttachments) {
	tb.Helper()
	checkCell(tb, got, want.Cell)
	if !want.HasAttachments() {
		return
	}

	cell, ok := got.(schema.HasAttachments)
	if !ok {
		tb.Fatal("cell has no attachments (does not implement schema.HasAttachments)")
	}

	var mb schema.MimeBundle
	att := cell.Attachments()
	if mb = att.MimeBundle(want.Filename); mb == nil {
		tb.Fatalf("no data for %s, want %q", want.Filename, want.Data)
	}

	require.Equal(tb, want.MimeType, mb.MimeType(), "reported mime-type")
	require.Equal(tb, want.Data, mb.Text(), "attachment data")
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
