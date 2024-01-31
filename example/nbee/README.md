# nbee

`nbee` ([ɛn-biː], nb extended example) is a small but diligent CLI tool that uses `nb` and several extensions packages to convert `.ipynb` files into beautiful Jupyter notebooks.

## Usage

Compile the package on the fly 🐝

```sh
go run github.com/bevzzz/nb/example/nbee@latest -f path/to/notebook.ipynb
```

Or, install a binary 🗑

```sh
# Build from source:
git clone https://github.com/bevzzz/nb.git
cd nb/example/nbee
go install

# Install from remote repository (Go 1.17+):
go install github.com/bevzzz/nb/example/nbee@latest
```

Try it out:

```sh
# If you already have bevzzz/nb checked out:
cd nb/testdata
nbee -f notebook.ipynb

# Convert your own notebooks
nbee -f my_jupyter.ipynb
```

## Disclaimer

This package is only an example of how `nb` can be extended with other packages.  
It's a showcase -- simple and minimal :)
