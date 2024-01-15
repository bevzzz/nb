# nb

	Render Jupyter Notebooks in pure Go 📔

This package is inspired by @yuin's [`goldmark`](https://github.com/yuin/goldmark) and is designed to be as clear and extensible.

The implementation follows the official [Jupyter Notebook format spec](https://nbformat.readthedocs.io/en/latest/format_description.html#the-notebook-file-format) (`nbformat`) and produces an output similar to that of [`nbconvert`](https://github.com/jupyter/nbconvert) (Jupyter's team own reference implementation) both structurally and visually. 

Although the current release only supports `v4.4` notebooks, support for other formats will be added soon (see the [**Roadmap**](#roadmap)).    
The package comes with an HTML renderer out of the box and can be extended to convert notebooks to other formats, such as LaTeX or PDF.

> 🏗 This package is being actively developed: its structure and APIs might change overtime.  
> If you find any bugs, please consider opening an issue or submitting a PR.

## Installation

```sh
go get github.com/bevzzz/nb
```

## Usage

`nb`'s default, no-frills converter can render markdown, code, and raw cells out of the box:

```go
b, err := os.ReadFile("notebook.ipynb")
if err != nil {
	panic(err)
}
err := nb.Convert(os.Stdout, b)
```

Without extensions or CSS the output may look rather bland and messy. The package comes with the Jupyter's classic light theme, and you can capture it by creating a new `html.Renderer` with a `WithCSSWriter` option:

```go
// Write both CSS and notebook's HTML to intermediate destinations
var body, css bytes.Buffer

// Configure your converter
c := nb.New(
  nb.WithRenderOptions(
		render.WithCellRenderers(
			html.NewRenderer(
				html.WithCSSWriter(&css),
			),
		),
	),
)

err := c.Convert(&body, b)
if er != nil {
	panic(err)
}

// Create the final output (could be a file)
f, _ := os.OpenFile("notebook.html", os.O_RDWR, 0644)
defer f.Close()

f.WriteString("<html><head><style>")
io.Copy(f, &css)
f.WriteString("</style></head>")

f.WriteString("<body>")
io.Copy(f, &body)
f.WriteString("</body></html>")
```

Finally, starting with `v0.2.0` you will be able to choose from a number of extensions to create neater notebooks.

## Roadmap 🗺

- **v0.2.0:**
	- Extension API + several extension packages (will be published separately)
		- Beautiful Markdown cells with [`goldmark`](https://github.com/yuin/goldmark)
		- Colourful `stream` outputs with [`ansihtml`](https://github.com/robert-nix/ansihtml)
		- JSON/XML output highlighting with [`chroma`](https://github.com/alecthomas/chroma)
	- Support for [cell attachments](https://nbformat.readthedocs.io/en/latest/format_description.html#cell-attachments) in markdown cells
	- Better granularity in the cell renderer registry API
- **v0.3.0:**
	- Support for older `nbformat` versions:
		- `v4.*`
		- `v3.0`
		- `v2.0`
	- Custom CSS
		-  Custom class prefix / class names
		-  Custom CSS with Go templates
- **v0.4.0**:
	- Support for `v1.0` notebooks
	- Built-in pretty-printing for JSON outputs
- Other:
	- I am curious about how `nb`'s performance measures against other popular libraries like [`nbconvert`](https://github.com/jupyter/nbconvert) (Python) and [`quarto`](https://github.com/quarto-dev/quarto-cli) (Javascript), so I want to do some benchmarking later.
	- As of now, I am not planning on adding converters to other formats (LaTeX, PDF, reStructuredText), but I will gladly consider this if there's a need for those.

If you have any other ideas or requests, please feel welcome to add a proposal in a new issue.

## Miscellaneous

### Math

Since Jupyter notebooks are often used for scientific work, you may want to display mathematical notation in your output.  
[MathJax](https://www.mathjax.org) is a powerful tool for that and [adding it to your  HTML header](https://www.mathjax.org/#gettingstarted) is the simplest way to get started.

Notice, that we want to _remove_ `<pre>` from the the list of skipped tags, as default HTML renderer will wrap raw and markdown cells in a `<pre>` tag.

```html
<html>
	<head>
		<script>
		    MathJax = {
		      options: {
		        skipHtmlTags: [ // includes "pre" by default
			        "script",
			        "noscript",
			        "style",
			        "textarea",
			        "code",
			        "annotation",
			        "annotation-xml"
			    ],
		      }
		    };
		</script>
	</head>
</html>
```

MathJax is very configurable and you can read more about that [here](https://docs.mathjax.org/en/latest/options/document.html#document-options).  
You may also find the [official MathJax config](https://nbformat.readthedocs.io/en/latest/markup.html#mathjax-configuration) used in the Jupyter project useful.

## License

This software is released under [the MIT License](https://opensource.org/license/mit/).
