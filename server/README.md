Talek Server
============

Server code is divided between a 'Frontend', which receives requests from users,
and relays them to the disparate trust domains, and 'Replica', which processes
requests for an individual trust domain, by maintaining a copy of the database,
which is updated and read by one or more 'Shard's.

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
* `READS_PER_WRITE` (default 20)
* `NUM_BUCKETS` (default 512)
* `BUCKET_DEPTH` (default 4)
* `DATA_SIZE` (default 512)
* `BATCH_SIZE` (default 8)
