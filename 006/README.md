# 006

Batch http requests.

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
sent delta 408826, received delta 248558
sent delta 363857, received delta 197634
sent delta 359795, received delta 224175
sent delta 352296, received delta 221782
sent delta 361684, received delta 220733
sent delta 348469, received delta 248425
sent delta 321361, received delta 249453
sent delta 361405, received delta 231084
^C
received shutdown signal
shutting down
done.
```

Better.
