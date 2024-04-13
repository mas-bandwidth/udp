# 005

Reuse HTTP connections.

https://blog.cubieserver.de/2022/http-connection-reuse-in-go-clients/

To run:

```console
go get
go run backend.go
```

in another terminal:

```console
go run server.go
```

in another terminal:

```console
go run client.go
```

Results:

```console
glenn@hulk:~/udp/005$ go run client.go
starting 1000 clients
sent delta 95851, received delta 1401
sent delta 95510, received delta 1445
sent delta 95254, received delta 1419
sent delta 95256, received delta 1423
sent delta 94706, received delta 1469
sent delta 95750, received delta 1342
sent delta 94960, received delta 1433
sent delta 94764, received delta 1418
^C
received shutdown signal
shutting down
done.
```

Better...
