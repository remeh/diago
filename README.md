# Diago

Diago is a visualization tool for profiles generated with `pprof`.

## Features

  - Visual interface with an easy-to-use read tree
  - Search in functions and filenames
  - Aggregate per functions or per function calls (lines)

![Screenshot of Diago](https://github.com/remeh/diago/raw/master/screenshot.png)

## Installation

```
go get -u github.com/remeh/diago
```

The `diago` binary should be available. Note that the build could take a few
seconds to complete due to the dependencies.

## Usage

```
./diago -file <profile-to-visualize>
```

## Roadmap

  - Read a profile from HTTP
  - Test profiles not generated with Go `http/pprof`

## Author

Rémy MATHIEU - @remeh

## License

Apache License 2.0
