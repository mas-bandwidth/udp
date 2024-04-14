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

We've already fixed it so it's not a HTTP requests per-second problem. Batching 1000 UDP packets, per-HTTP request, we turn it into 1000 requests per-second, but... each request is 1000 * 100 bytes = 100 kilobytes long.

At this point it's clear this is an impossible problem. Or at least a problem that is waaaaaaaaaaaay outside the scope of anything reasonable to implement for a take home programmer test.

It's entirely IO bound at this point, not anything we can fix by writing code. Indeed, attempting to write code past this point would be an exercise in frustration without understanding that it's IO bound.

As next steps, the logical next step is to fix the IO issue by adding a second virtual NIC to the server VM, and a virtual network and subnetwork just for HTTP traffic. This would fix the IO boundness, and then we have a solution that conservatively handles 10k clients with two c3-highcpu-44 VMs, one for the server and one for the backend.

Since the expected load is 1M clients, we can then scale this horizontally to get an estimate of the cost to run the VMs in google cloud:

100 * 20 * c3-highcpu-44 = 2000 * $2405.57 = $4,811,140 USD per-month just for VMs.

We also need to consider egress bandwidth. It's roughly 10c per-GB egress. The response packets are just 8 bytes for the hash, but we need to add 28 bytes for IP and UDP header. Not sure if I should add ethernet header or not to the calculation, so let's just go with 36 bytes per-response UDP packet.

1M players * 100 response packets per-second * 36 bytes = 100M * 36 bytes = 100,000,000 * 36 bytes/sec = 3,600,000,000 bytes/sec = 3.6GB/sec.

2,592,000 seconds in a month, so 2,592,000 * 3.6GB = 9,331,200 GB per-month.

At $0.1 per-GB, we get an egress bandwidth charge of: $933,120 USD per-month.

But this is not all, we also need to consider that we'll put a load balancer in front of the UDP server. (Assume we can pin each load balancer to a backend instance, and those are not load balanced).

Ingress traffic to load balancers is billed at ~1c per-GB on google cloud.

We have 28+100 bytes per-UDP packet, and 1M players sending 100 packets per-second. 

1M players * 100 request packets per-second * 128 bytes = 12.8GB/sec.

2,592,000 seconds in a month, so 2,592,000 * 12.8GB = 9,331,200 GB per-month.

Ingress traffic to the UDP load balance is $331,776 USD per-month.

Total cost: $6,076,036 USD per-month.
