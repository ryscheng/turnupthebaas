# Talek
[![Build Status](https://travis-ci.org/privacylab/talek.svg?branch=master)](https://travis-ci.org/privacylab/talek)
[![GoDoc](https://godoc.org/github.com/privacylab/talek?status.svg)](https://godoc.org/github.com/privacylab/talek)

Talek is a privacy-preserving messaging system. User communication is stored on untrusted systems using PIR.

## Getting Started

A basic client (which is not resistant to traffic analysis!) can be found at   
```go get github.com/privacylab/talek/cli/talekclient```

Talek uses a construct called topic handles. Topics represent a stream of
messages from one author to a few readers. The author who creates a topic can
provide a handle to it to allow others to "follow along". A longer description
of the specific guarantees of a topic are provided in the academic paper linked
below.

### Basic Usage:

    talekclient --config=talek.conf --create --topic=newhandle
    talekclient --config=talek.conf --topic=newhandle --write "Hello World"
    talekclient --config=talek.conf --topic=newhandle --share=readOnlyHandle
    talekclient --config=talek.conf --topic=readOnlyHandle --read

### Following Along:

Join the mailing list: https://lists.riseup.net/www/info/talek

## Develop
Pull requests are welcome! Please run all tests (see below) before submitting a PR.

### Tools

- [govendor](https://github.com/kardianos/govendor) for vendoring
- [gometalinter](https://github.com/alecthomas/gometalinter) for linting

```bash
$ make get-tools
```

### Testing

All tests should pass before submitting a pull request

```bash
$ make lint
$ make test
```

### Vendoring

Talek vendors all of its dependencies into the local `vendor/` directory.
To add or update dependencies to the latest in `vendor/`, use the `govendor` tool, as follows:
- `govendor fetch github.com/foo/bar`

To see a list and status of dependencies:
- `govendor list`


## Publication

Talek: a Private Publish-Subscribe Protocol.   
Raymond Cheng, Will Scott, Bryan Parno, Irene Zhang, Arvind Krishnamurthy, Tom Anderson.   
In Submission. 2017.   
[PDF](https://raymondcheng.net/download/papers/talek-tr.pdf)
