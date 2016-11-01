PDB Server
============

Testing Shard Performance
------------------------

Shard performance can be tested to understand read and write throughputs at
different workloads and database sizes using the shard_test code.

An example is
```golang
go test -run 0 -bench BenchmarkShard -benchtime 1m
```

Workload / sizing parameters are controlled by the following environmental vars:

* `PIR_SOCKET` (default "pir.socket")
* `READS_PER_WRITE` (defualt 20)
* `NUM_BUCKETS` (default 512)
* `BUCKET_DEPTH` (default 4)
* `DATA_SIZE` (default 512)
* `BATCH_SIZE` (default 8)
