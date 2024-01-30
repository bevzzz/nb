# nbee

`nbee` ([ɛn-biː], nb extended example) is a small but diligent CLI tool that uses `nb` and several extensions packages to convert `.ipynb` files into beautiful Jupyter notebooks.

## Usage

Compile the package on the fly 🐝

```sh
go run github.com/bevzzz/nb/example/nbee
```

Or, install a binary 🗑

```sh
go install github.com/bevzzz/nb/example/nbee
```

Try it out:

```sh
nbee # convert the default notebook to HTML
nbee -f "my_notebook.ipynb" # convert you own notebooks
```

## Disclaimer

This package is only an example of how `nb` can be extended with other packages.  
It's a showcase -- simple and minimal :)