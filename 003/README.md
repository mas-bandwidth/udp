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
starting 1000 clients
sent delta 97545, received delta 97545
sent delta 94000, received delta 94000
sent delta 94110, received delta 94003
sent delta 94559, received delta 94666
sent delta 95927, received delta 95869
sent delta 95014, received delta 95072
sent delta 94000, received delta 93992
sent delta 94000, received delta 94008
^C
received shutdown signal
shutting down
done.
```

Once again, no change. 

The reason is that we're sending all the packets from the same source IP address to the same dest IP address, so the hash is always the same, and they end up getting processed on one core only.
