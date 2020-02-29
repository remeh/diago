# Diago

Diago is a visualization tool for profiles generated with `pprof`.

## Features

  - Visual interface with an easy-to-use read tree
  - Search in functions and filenames
  - Aggregate per functions or per function calls (lines)

![Screenshot of Diago](https://github.com/remeh/diago/raw/master/screenshot.png)

## Installation

Due to the underlying usage of `go-gl/glfw`, there is a few system dependencies (i.e. some Xorg libraries on Linux or headers/libraries on macOS). See [this link](https://github.com/go-gl/glfw#installation) for detailed information.

You'll need Go installed (only tested with Go >= 1.12), then:

```
go get -u github.com/remeh/diago
```

The `diago` binary should be available in `$GOPATH/bin` or `$HOME/go/bin` if the `$GOPATH` environment variable is not set.

Note that the build could take a few seconds to complete due to the dependencies.

## Usage

```
./diago -file <profile-to-visualize>
```

## Roadmap

  - Read a profile from HTTP
  - Test profiles not generated with Go `http/pprof`
  - Heap visualization

## Author

Rémy MATHIEU - @remeh

## License

Apache License 2.0
