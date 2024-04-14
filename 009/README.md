# 008

Reduce server to c3-highcpu-44, since we only really get 16 threads per-NIC to receive packets with.

To run:

```console
terraform init
terraform apply
```

Result:

We can still hit 10k clients, but at a much lower CPU cost.
