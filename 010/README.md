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

But there are still the egress bandwidth charges, and these dominate. Surely at such a scale, the egress bandwidth price would be greatly reduced with sales negotiations with Google Cloud, but let's go a step further.

Let's run the system in bare metal.

<img width="1348" alt="Screenshot 2024-04-14 at 10 06 40 AM" src="https://github.com/mas-bandwidth/udp/assets/696656/23e29eee-645c-4e61-bf7e-c4471d33f4f4">

I love https://datapacket.com. They are an excellent bare metal hosting company. Picking the fattest bare metal server they have with a 40GB NIC, there are no ingress or egress bandwidth charges.

On bare metal with a 40GB NIC we can handle the amount of traffic easily with XDP. The XDP program runs in native mode, whereas on google cloud it runs in SKB mode, so it's slower. There's also no virtualization overhead. All the cores on the bare metal are dedicated exclusively to doing the XDP work and you can tune the system as needed.

More about XDP here: https://mas-bandwidth.com/xdp-for-game-programmers/

The total cost for 1M clients is now: $8,430 USD per-month.

$6,076,036 / $8,430 = 720X reduction in cost.

Can we take it even further? Yes!

If we needed to scale up more, at some point XDP is not fast enough. 

We could purchase and install a netronome NIC that would run the XDP hash function in hardware. Alternatively, we could explore implementing the hash with with programmable NIC using P4.

If we need to scale up even further, perhaps another 100 - 1000X, we could scale out horizontally with multiple bare metal machines with NICs that have onboard FPGA and implement the hash there. _Although, this is mildly insane._

What's the moral of the story here?

Rewriting it in rust won't help you. 

1. Work out how much it costs
2. Prototype it, load test it and really understand the problem
3. Break all the rules :)
