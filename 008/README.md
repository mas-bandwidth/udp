# 008

Reduce server to c3-highcpu-44, since we only really get 16 threads per-NIC to receive packets with.

To run:

```console
terraform init
terraform apply
```

You'll need significant quota for n1 cores and c3 instances in your google cloud to be able to run. If you don't have it, edit the main.tf for different instance types.

Result:

We can still hit 10k clients, now at a much lower cost.
