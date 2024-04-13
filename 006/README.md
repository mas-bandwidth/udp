# 006

Increase to batches of 1000, increase timeout to 10 sec, reduce copies and don't use channels.

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
glenn@hulk:~/udp/006$ go run client.go
starting 30000 clients
sent delta 332252, received delta 285135
sent delta 288638, received delta 223007
sent delta 258424, received delta 249822
sent delta 269752, received delta 255171
sent delta 288963, received delta 216977
sent delta 235732, received delta 281624
sent delta 261674, received delta 241142
sent delta 233617, received delta 261964
sent delta 259070, received delta 256569
sent delta 261217, received delta 260470
sent delta 243065, received delta 273405
sent delta 274137, received delta 203934
sent delta 257131, received delta 279429
^C
received shutdown signal
shutting down
done.
```

As close as I can get. I think I'm hitting io limits on the one machine.

Dropping down to 25k clients, we have a solution that has no dropped packets:


