# 001

First attempt. 

Keep it simple and just see how quickly we can naively send UDP packets from a client to server, and just reply with an 8 byte UDP packet containing the hash.

To run:

```console
go run server.go
```

then in another terminal:

```console
go run client.go
```

Results:

```console
gaffer@batman 001 % go run client.go
starting 100 clients
sent delta 9100, received delta 9100
sent delta 9100, received delta 9100
sent delta 9100, received delta 9100
sent delta 9200, received delta 9200
sent delta 9100, received delta 9100
sent delta 9100, received delta 9100
sent delta 9100, received delta 9100
sent delta 9114, received delta 9100
sent delta 9186, received delta 9192
^C
received shutdown signal
shutting down
done.
```

Results on localhost interface on an old iMacPro, I can easily do around 100 clients worth of packets without any drops (~10k packets per-second).

Each client is on its own thread, and sleeps for 10ms before sending each packet. I believe we see ~9100 packets per-second sent because the sleeps run a bit long, on average.

Somewhere around 400-500 clients I start to see some drops.

At 1000 clients, I see signficant packet loss:

```console
gaffer@batman 001 % go run client.go
starting 1000 clients
sent delta 99773, received delta 57616
sent delta 98627, received delta 56149
sent delta 98336, received delta 53192
^C
received shutdown signal
shutting down
done.
```
