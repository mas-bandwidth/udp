# 004

Naive implementation of the server <-> backend over HTTP.

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
glenn@hulk:~/udp/004$ go run client.go
starting 30000 clients
sent delta 383894, received delta 3715
sent delta 239831, received delta 146
sent delta 377213, received delta 1595
sent delta 360799, received delta 595
sent delta 309412, received delta 8
sent delta 319754, received delta 0
sent delta 309425, received delta 112
^C
received shutdown signal
shutting down
done.
```

NOPE.

<img width="1540" alt="image" src="https://github.com/mas-bandwidth/udp/assets/696656/3e9e4ac7-7e3f-49ef-a94d-f41a049d8d47">
