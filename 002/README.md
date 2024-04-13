# 002

Second attempt. 

Increase socket send and receive buffer sizes to 2MB.

https://medium.com/@CameronSparr/increase-os-udp-buffers-to-improve-performance-51d167bb1360

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
gaffer@batman 002 % go run client.go
starting 1000 clients
sent delta 99583, received delta 54445
sent delta 96432, received delta 57645
sent delta 98198, received delta 54189
sent delta 98059, received delta 56692
sent delta 98449, received delta 55664
sent delta 98163, received delta 54361
^C
received shutdown signal
shutting down
done.
```

No change. On MacOS it is not the socket buffer sizes. They're already big enough (at this small scale) by default.
