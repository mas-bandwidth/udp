# 010

The only way to pass this test is to break the rules of the test.

Drop the server <-> backend HTTP requests and switch from Golang to XDP.

To run:

```console
terraform init
terraform apply
```

Result:

<img width="1536" alt="image" src="https://github.com/mas-bandwidth/udp/assets/696656/5b491871-d1d7-4b10-9aca-12328b49b1f2">

We can now run 100k clients for each c3-highcpu-44 instance. This is a 20X cost saving from $4,811,140 USD per-month to $240,557 for the VMs.

But there are still the egress bandwidth charges, and these dominate. Surely at such a scale, the egress bandwidth would be greatly reduce, but let's go a step further.

Let's run the system in bare metal.

<img width="1348" alt="Screenshot 2024-04-14 at 10 06 40 AM" src="https://github.com/mas-bandwidth/udp/assets/696656/23e29eee-645c-4e61-bf7e-c4471d33f4f4">

I love https://datapacket.com. They are an excellent bare metal hosting company. Picking the fattest bare metal server they have with a 100GB NIC, I have no ingress or egress bandwidth charges.

The total cost: $
