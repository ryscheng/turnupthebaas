PIR Server Library and Bindings
===============================

This code represents a Go language binding for a PIR server.

Benchmarks of PIR performance can be gathered using the
test harness, which is parameterized using environmental
variables.

An example benchmark might look like
```shell
PIR_SOCKET=../pird/pir.socket PIR_CELL_LENGTH=2048 PIR_CELL_COUNT=262144 PIR_BATCH_SIZE=8 go test -run x -bench .
```
