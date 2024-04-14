/*
    UDP server XDP program

    Replies to 100 byte UDP packets sent to port 40000 with the fnv1a 64bit hash (8 bytes)

    USAGE:

        clang -Ilibbpf/src -g -O2 -target bpf -c server_xdp.c -o server_xdp.o
        sudo cat /sys/kernel/debug/tracing/trace_pipe
*/

#include <linux/in.h>
#include <linux/if_ether.h>
#include <linux/if_packet.h>
#include <linux/if_vlan.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/udp.h>
#include <linux/bpf.h>
#include <linux/string.h>
#include <bpf/bpf_helpers.h>

#if defined(__BYTE_ORDER__) && defined(__ORDER_LITTLE_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_ntohs(x)        __builtin_bswap16(x)
#define bpf_htons(x)        __builtin_bswap16(x)
#elif defined(__BYTE_ORDER__) && defined(__ORDER_BIG_ENDIAN__) && \
    __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_ntohs(x)        (x)
#define bpf_htons(x)        (x)
#else
# error "Endianness detection needs to be set up for your compiler?!"
#endif

//#define DEBUG 1

#if DEBUG
#define debug_printf bpf_printk
#else // #if DEBUG
#define debug_printf(...) do { } while (0)
#endif // #if DEBUG

static void reflect_packet( void * data, int payload_bytes )
{
    struct ethhdr * eth = data;
    struct iphdr  * ip  = data + sizeof( struct ethhdr );
    struct udphdr * udp = (void*) ip + sizeof( struct iphdr );

    __u16 a = udp->source;
    udp->source = udp->dest;
    udp->dest = a;
    udp->check = 0;
    udp->len = bpf_htons( sizeof(struct udphdr) + payload_bytes );

    __u32 b = ip->saddr;
    ip->saddr = ip->daddr;
    ip->daddr = b;
    ip->tot_len = bpf_htons( sizeof(struct iphdr) + sizeof(struct udphdr) + payload_bytes );
    ip->check = 0;

    char c[ETH_ALEN];
    memcpy( c, eth->h_source, ETH_ALEN );
    memcpy( eth->h_source, eth->h_dest, ETH_ALEN );
    memcpy( eth->h_dest, c, ETH_ALEN );

    __u16 * p = (__u16*) ip;
    __u32 checksum = p[0];
    checksum += p[1];
    checksum += p[2];
    checksum += p[3];
    checksum += p[4];
    checksum += p[5];
    checksum += p[6];
    checksum += p[7];
    checksum += p[8];
    checksum += p[9];
    checksum = ~ ( ( checksum & 0xFFFF ) + ( checksum >> 16 ) );
    ip->check = checksum;
}

SEC("server_xdp") int server_xdp_filter( struct xdp_md *ctx ) 
{ 
    void * data = (void*) (long) ctx->data; 

    void * data_end = (void*) (long) ctx->data_end; 

    struct ethhdr * eth = data;

    if ( (void*)eth + sizeof(struct ethhdr) < data_end )
    {
        if ( eth->h_proto == __constant_htons(ETH_P_IP) ) // IPV4
        {
            struct iphdr * ip = data + sizeof(struct ethhdr);

            if ( (void*)ip + sizeof(struct iphdr) < data_end )
            {
                if ( ip->protocol == IPPROTO_UDP ) // UDP
                {
                    struct udphdr * udp = (void*) ip + sizeof(struct iphdr);

                    if ( (void*)udp + sizeof(struct udphdr) <= data_end )
                    {
                        if ( udp->dest == __constant_htons(40000) )
                        {
                            __u8 * payload = (void*) udp + sizeof(struct udphdr);
                            int payload_bytes = data_end - (void*)payload;
                            if ( payload_bytes == 100 && (void*)payload + 100 <= data_end )    // IMPORTANT: for the verifier
                            {
                                reflect_packet( data, 8 );
                                __u64 hash = 0xCBF29CE484222325;
                                hash ^= payload[0]; hash *= 0x00000100000001B3;
                                hash ^= payload[1]; hash *= 0x00000100000001B3;
                                hash ^= payload[2]; hash *= 0x00000100000001B3;
                                hash ^= payload[3]; hash *= 0x00000100000001B3;
                                hash ^= payload[4]; hash *= 0x00000100000001B3;
                                hash ^= payload[5]; hash *= 0x00000100000001B3;
                                hash ^= payload[6]; hash *= 0x00000100000001B3;
                                hash ^= payload[7]; hash *= 0x00000100000001B3;
                                hash ^= payload[8]; hash *= 0x00000100000001B3;
                                hash ^= payload[9]; hash *= 0x00000100000001B3;
                                bpf_xdp_adjust_tail( ctx, -( payload_bytes - 8 ) );
                                payload[0] = ( hash       ) & 0xFF;
                                payload[1] = ( hash >> 8  ) & 0xFF;
                                payload[2] = ( hash >> 16 ) & 0xFF;
                                payload[3] = ( hash >> 24 ) & 0xFF;
                                payload[4] = ( hash >> 32 ) & 0xFF;
                                payload[5] = ( hash >> 40 ) & 0xFF;
                                payload[6] = ( hash >> 48 ) & 0xFF;
                                payload[7] = ( hash >> 56 );
                                return XDP_TX;
                            }
                            else
                            {
                                return XDP_DROP;
                            }
                        }
                    }
                }
            }
        }
    }

    return XDP_PASS;
}

char _license[] SEC("license") = "GPL";
