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
starting 1000 clients
sent delta 89355, received delta 1502
sent delta 74633, received delta 332
sent delta 82034, received delta 166
sent delta 84410, received delta 4
sent delta 93893, received delta 0
sent delta 93719, received delta 0
sent delta 90368, received delta 0
sent delta 82220, received delta 5
sent delta 75389, received delta 24
sent delta 64667, received delta 13
sent delta 60097, received delta 179
sent delta 52373, received delta 181
sent delta 70531, received delta 47
sent delta 83631, received delta 2
sent delta 60504, received delta 351
sent delta 60579, received delta 82
sent delta 64227, received delta 75
sent delta 61369, received delta 2
sent delta 94000, received delta 0
sent delta 94689, received delta 0
sent delta 94000, received delta 0
sent delta 94716, received delta 0
sent delta 93285, received delta 0
sent delta 94000, received delta 0
sent delta 93835, received delta 0
sent delta 94132, received delta 0
sent delta 94035, received delta 0
^C
received shutdown signal
shutting down
done.
```

NOPE.

<img width="1540" alt="image" src="https://github.com/mas-bandwidth/udp/assets/696656/3e9e4ac7-7e3f-49ef-a94d-f41a049d8d47">
