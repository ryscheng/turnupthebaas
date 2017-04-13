# Talek

Talek is system for privacy-preserving publish-subscribe that shares user data through untrusted servers.

## Getting Started
Please check back when we are ready for release

## Develop

### Tools
- [govendor](https://github.com/kardianos/govendor) for vendoring
- [gometalinter](https://github.com/alecthomas/gometalinter) for linting
```bash
$ go get -u github.com/kardianos/govendor
$ go get -u github.com/alecthomas/gometalinter
$ gometalinter --install
```

### Vendoring
Talek vendors all of its dependencies into the local `vendor/` directory.
To add or update dependencies to the latest in `vendor/`, use the `govendor` tool, as follows:
- `govendor fetch github.com/foo/bar`

To see a list and status of dependencies:
- `govendor list`

### Testing
All tests should pass before submitting a pull request
```bash
$ govendor test +local          # Run unit tests
$ gometalinter --vendor ./...   # Run linter
```

## Publication
Talek: a Private Publish-Subscribe Protocol.   
Raymond Cheng, Will Scott, Bryan Parno, Irene Zhang, Arvind Krishnamurthy, Tom Anderson.   
In Submission. 2017.   
[PDF](https://raymondcheng.net/download/papers/talek-tr.pdf)
