# Talek

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

## Developing

Pull requests are welcome, as clich� as that sounds! Code should pass gometalint.

## Publication
Talek: a Private Publish-Subscribe Protocol.
Raymond Cheng, Will Scott, Bryan Parno, Irene Zhang, Arvind Krishnamurthy, Tom Anderson.
In Submission. 2017.
[PDF](https://raymondcheng.net/download/papers/talek-tr.pdf)
