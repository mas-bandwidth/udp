# 006

Increase to batches of 1000, don't use channels, and reduce copies.

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
starting 1000 clients
sent delta 96544, received delta 54116
sent delta 96771, received delta 92884
sent delta 95436, received delta 95000
sent delta 95456, received delta 94000
sent delta 95951, received delta 97174
sent delta 95473, received delta 93094
sent delta 96212, received delta 96732
sent delta 95523, received delta 104000
sent delta 95964, received delta 94615
sent delta 96412, received delta 96624
sent delta 93194, received delta 81761
sent delta 96762, received delta 102685
sent delta 96001, received delta 100033
sent delta 96520, received delta 91282
sent delta 95962, received delta 98000
sent delta 95105, received delta 95000
sent delta 95542, received delta 97493
sent delta 95896, received delta 94507
sent delta 96073, received delta 94000
sent delta 96125, received delta 99000
^C
received shutdown signal
shutting down
done.
```

As close as I can get. I'm out of ideas for making the HTTP stuff any faster. I think I'm hitting IO limits on my machine.

It's time to start going horizontal for the clients, so we have different source IP addresses hashing to cores on the server.

