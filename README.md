# Talek
[![Build Status](https://travis-ci.org/privacylab/talek.svg?branch=master)](https://travis-ci.org/privacylab/talek)
[![Coverage Status](https://coveralls.io/repos/github/privacylab/talek/badge.svg?branch=master)](https://coveralls.io/github/privacylab/talek?branch=master)
[![GoDoc](https://godoc.org/github.com/privacylab/talek?status.svg)](https://godoc.org/github.com/privacylab/talek)

Talek is a privacy-preserving messaging system. User communication is stored on untrusted systems using PIR.

## Getting Started
A basic client can be found at
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


## Develop
Pull requests are welcome! Please run all tests (see below) before submitting a PR.

### System Dependencies
Depending on which PIR implementation you use, you may need to install OpenCL / CUDA.
Make sure you have the latest graphics drivers for your video card.

NVIDIA CUDA:
- [Drivers](http://www.nvidia.com/Download/index.aspx?lang=en-us)
- [CUDA](https://developer.nvidia.com/cuda-downloads)

OpenCL on Ubuntu:

    sudo apt-get install -y ocl-icd-libopencl1 ocl-icd-opencl-dev opencl-headers clinfo
    sudo ln -s /usr/lib/x86_64-linux-gnu/libOpenCL.so /usr/lib/x86_64-linux-gnu/libCL.so

OpenCL on macOS:
- OpenCL is included in the developer tools. See [here](https://developer.apple.com/opencl/)



### Tools
- [gometalinter](https://github.com/alecthomas/gometalinter) for linting

```bash
$ make get-tools
```

### Testing
All tests should pass before submitting a pull request

```bash
$ make test
```

The GPU backings are not built by default. Changes to `pir/`, where the
backing interface may be affected should ensure that code is tested with
`go test -tags 'cuda,opencl'` to include testing of all drivers.


## Following Along:
Join the mailing list: https://lists.riseup.net/www/info/talek


## Publication
Talek: a Private Publish-Subscribe Protocol.   
Raymond Cheng, Will Scott, Bryan Parno, Irene Zhang, Arvind Krishnamurthy, Tom Anderson.   
In Submission. 2017.   
[PDF](https://raymondcheng.net/download/papers/talek-tr.pdf)
