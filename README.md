# Stock State

*Stock State* is a centralized stock manager for flash sale activities inspired by the Timestamp Oracle in Percolator
and TiDB's Placement Driver.

It's essentially a server that hands out the remained stocks for the common flash sales on the Internet. Stock State
periodically allocates a range of stocks by writing a cap of the allocated stocks into etcd for fault-tolerance. Stock
State then can handle the future client requests from memory. If the leader of the Stock State cluster fails, a new leader
will be elected and continue handling the client requests from the cap stored in etcd, which make the system highly available
in spite of leaving some stocks not handed out. Typically, we must keep the handled number of stocks not exceeding
the stock limit, but leaving some stocks not handled out is acceptable.

This project demonstrates how to implement such a server, while many details remain to be optimized to 
make it production-ready.

## Run this program

Before we run the program, an etcd server should be installed and running at local port 2379.

At the root path, run
```shell
go run main.go
```
to start the server.

Then we run the client program where we send some requests to the running server:
```shell
cd example
go run main.go
```

Note that currently the data stored in etcd will not be auto-cleanup, users may manually delete it before a new run.
