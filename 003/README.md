# 003

Third attempt. 

SO_REUSEPORT

https://medium.com/high-performance-network-programming/performance-optimisation-using-so-reuseport-c0fe4f2d3f88

To run:

```console
go get
go run server.go
```

then in another terminal:

```console
go run client.go
```

Results:

```console
glenn@hulk:~/udp/003$ go run client.go
starting 30000 clients
sent delta 416614, received delta 416501
sent delta 342716, received delta 342561
sent delta 342897, received delta 342660
sent delta 341856, received delta 341719
sent delta 343937, received delta 344361
sent delta 342054, received delta 342165
sent delta 339302, received delta 339368
sent delta 340573, received delta 340552
sent delta 338369, received delta 338167
sent delta 340355, received delta 340550
sent delta 338247, received delta 338135
^C
received shutdown signal
shutting down
done.
```

Can handle around 30k clients on loopback which is 300,000 packets per-second, and we're still using really naive sendto/recvfrom equivalents (syscall for every packet send and receive).

Nothing amazing, but decent enough to start exploring the server <-> backend part over HTTP...
