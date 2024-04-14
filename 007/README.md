# 007

Go to google cloud with terraform, run 10 n1-standard-8 VMs with 1000 clients each to get 10K players each against a c3-highcpu-176

To run:

```console
terraform init
terraform apply
```

You'll need significant quota for n1 cores and c3 instances in your google cloud to be able to run. If you don't have it, edit the main.tf for different instance types.

Result:

![image](https://github.com/mas-bandwidth/udp/assets/696656/3db5afa7-ad3a-46c4-8f70-b702054f74fb)

Even with the c3-highcpu-176 for the server, we can only get 10-25k clients max. Above this, UDP packets start to get dropped.

You can see these drops on the server with:

```
sudo apt install net-tools -y
netstat -anus
```

The problem is that with c3 class machines: 

"Using gVNIC, the maximum number of queues per vNIC is 16. If the calculated number is greater than 16, ignore the calculated number, and assign each vNIC 16 queues instead."

https://cloud.google.com/compute/docs/network-bandwidth

This means even though we have a massive amount of cores, only 16 are actively available to receive packets. When we overload these, UDP packets are dropped on receive, even though we are only using a fraction of the CPU available.
