# 009

Bring back the server <-> backend HTTP comms in google cloud.

To run:

```console
terraform init
terraform apply
```

Result:

BOOM.

<img width="1613" alt="image" src="https://github.com/mas-bandwidth/udp/assets/696656/b89211ec-6230-47ca-8bdc-03611b95f262">

We're simply asking too much of the HTTP client. It can't keep up with the UDP packets.

10k players, 100 packets per-second = 10,000 * 100 = 1,000,000 UDP packets per-second.

We've already fixed it so it's not a HTTP requests per-second problem. Batching 1000 UDP packets, per-HTTP request, we turn it into 1000 requests per-second, but each request is 1000 * 100 bytes = 100 kilobytes long.

At this point it's clear this is an impossible problem. Or at least a problem that is waaaaaaaaaaaay outside the scope of anything reasonable to implement for a take home programmer test.

It's entirely IO bound at this point, not anything we can fix by writing code. Indeed, attempting to write code past this point would be an exercise in frustration without understanding that it's IO bound.

As next steps, we could add a second virtual NIC to the server VM, and a virtual network and subnetwork just for HTTP traffic. This would fix the IO boundness, and then we have a solution that handles conservatively 10k clients with two c3-highcpu-44 VMs.

Since the expected load is 1M clients, we can then scale this horizontally to get an estimate of the cost to run the VMs in google cloud:

100 * 20 * c3-highcpu-44 = 2000 * $2405.57 = $4,811,140 USD per-month just for VMs.

But this won't be all. The cost of load balancing this traffic, even internally in google cloud is $0.01 USD per-GB.

1000 requests per-second @ 100k per-request = 1000 * 100k per-second = 1000 * 100,000 bytes/sec = 
