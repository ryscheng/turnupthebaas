# OpenCL PIR Daemon

## Compile
PIR daemon depends on OpenCL headers to compile. For example on Ubuntu:

```bash
$ sudo apt-get install opencl-headers
$ make    # creates ./pird binary
```

## Usage

```bash
$ ./pird [-l] [-d <device id>] [-s <socket>]
```

`pird` must be assigned dedicated GPU. 
To see what GPU devices exist, use `-l`.
Then pin `pird` to that GPU with the `-d` flag.
Applications interact with `pird` using the UNIX socket specified with `-s`.

## Optimization

Communication with the Daemon occurs over a unix socket interface.
To optmize capacity over this interace, transmitted communication should
be sized to fit in a single system call.

To figure out the size, perform the following calcuation:

1. Input PIR vectors will be sized as `[batch size] * [# DB entries] / 8` bytes.
2. Output responses will be sized as `[batch size] * [DB entry size]` bytes.
3. The unix socket buffer size should be sized at one power of two above the larger of these two numbers.

You can view the current unix socket buffer size on your machine at
`/proc/sys/net/core/wmem_max` and `/proc/sys/net/core/rmem_max`. You can set it using `sysctl net.core.wmem_max=VALUE` (and add to `/etc/sysctl.conf` to persist over reboots.)

