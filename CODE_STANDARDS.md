# Code Standards

This project uses the following code standards and specifications from:
- [effective go](https://golang.org/doc/effective_go.html)
- [go benchmarks](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [go examples](https://golang.org/pkg/testing/#hdr-Examples)
- [go tests](https://golang.org/pkg/testing/)
- [godoc](https://godoc.org/golang.org/x/tools/cmd/godoc)
- [gofmt](https://golang.org/cmd/gofmt/)
- [golangci-lint](https://golangci-lint.run/)
- [report card](https://goreportcard.com/)

### *effective go* standards
View the [effective go](https://golang.org/doc/effective_go.html) standards documentation.

### *golangci-lint* specifications
The package [golangci-lint](https://golangci-lint.run/usage/quick-start) runs several linters in one package/cmd.

View the active linters in the [configuration file](.golangci.yml).

Install via macOS:
```shell
brew install golangci-lint
```

Install via Linux and Windows:
```shell
# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.31.0
golangci-lint --version
```

### *godoc* specifications
All code is written with documentation in mind. Follow the best practices with naming, examples and function descriptions.