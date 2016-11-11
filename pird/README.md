OpenCL PIR Daemon
=================

Usage
-----

Optimization
------------

Communication with the Daemon occurs over a unix socket interface.
To optmize capacity over this interace, transmitted communication should
be sized to fit in a single system call.

To figure out the size, perform the following calcuation:

1. Input PIR vectors will be sized as `[batch size] * [# DB entries] / 8` bytes.
2. Output responses will be sized as `[batch size] * [DB entry size]` bytes.
3. The unix socket buffer size should be sized at one power of two above the larger of these two numbers.

You can view the current unix socket buffer size on your machine at
`/proc/sys/net/core/wmem_max` and `/proc/sys/net/core/rmem_max`. You can set it using `sysctl net.core.wmem_max=VALUE` (and add to `/etc/sysctl.conf` to persist over reboots.)

