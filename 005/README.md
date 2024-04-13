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
starting 30000 clients
sent delta 471212, received delta 41258
sent delta 401873, received delta 33887
sent delta 413246, received delta 32503
sent delta 413949, received delta 32823
sent delta 408071, received delta 32602
sent delta 414328, received delta 30892
sent delta 398925, received delta 30284
^C
received shutdown signal
shutting down
done.
```

Better.
