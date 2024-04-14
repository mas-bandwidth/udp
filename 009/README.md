# 009

Bring back the server <-> backend HTTP comms in google cloud.

To run:

```console
terraform init
terraform apply
```

Result:

BOOM.

We're firing off too many HTTP requests. Each request opens a new socket. We quickly run out of open sockets.
