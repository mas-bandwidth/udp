# 009

Bring back the server <-> backend HTTP comms in google cloud.

To run:

```console
terraform init
terraform apply
```

Result:

BOOM.

We're simply asking too much of the HTTP client. It can't keep up with the UDP packets.

10k players, 100 packets per-second = 10,000 * 100 = 1,000,000 UDP packets per-second.

Batching 1000 UDP packets, per-HTTP request, we turn it into 1000 requests per-second, but each request is 1000 * 100 bytes = 100 kilobytes long.

At this point it's clear this is an impossible problem. Or at least a problem that is waaaaaaaaaaaay outside the scope of anything reasonable to implement for a take home programmer test.

As next steps, we could add a second virtual NIC to the server VM, and a virtual network and subnetwork just for HTTP traffic. This would fix the IO boundness, and then we have a solution that handles conservatively 10k clients with two c3-highcpu-44 VMs.

Since the expected load is 1M clients, we can thes scale this horizontally to get an estimate of the cost to run the VMs in google cloud:

100 * 20 * c3-highcpu-44 = 2000 * $2405.57 = $4,811,140 USD per-month just for VMs.
