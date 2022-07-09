#include "headers/common.h"
#include "headers/maps.h"
#include "headers/logger.h"
#include <linux/in.h>
#include <linux/socket.h>

// lightpath_sock_connect handles generic socket connections
__attribute__((section("connect"), used)) int lightpath_sock_connect(struct bpf_sock_addr *ctx)
{

  __u32 dst_ip[4];
  __u64 cookie;

  // Drop non TCP sockets immediately
  if (ctx->protocol != IPPROTO_TCP)
  {
    debugk("Non supported protocol: %d", ctx->protocol);
    return 0;
  }

  // Extract IP
  {
    switch (ctx->family)
    {
    case AF_INET:
      // IPv4 -> let's fill the tuple
      dst_ip[0] = 0;
      dst_ip[1] = 0;
      dst_ip[2] = 0;
      dst_ip[2] = ctx->user_ip4;
      break;
    case AF_INET6:
      // IPv6
      dst_ip[0] = ctx->user_ip6[0];
      dst_ip[1] = ctx->user_ip6[1];
      dst_ip[2] = ctx->user_ip6[2];
      dst_ip[3] = ctx->user_ip6[3];
      break;
    default:
      // Drop unknown internet protocols
      debugk("Unknown family: %d", ctx->family);
      return 0;
    }
  }

  // Check if IP is a cluster IP
  {
    void *res = bpf_map_lookup_elem(&cluster_ip_inclusion_map, &dst_ip);
    if (res)
    {
      // Not in the hashmap -> no action required
      debugk("Not a cluster IP: %d.%d.%d.%d", dst_ip[0], dst_ip[1], dst_ip[2], dst_ip[3]);
      return 0;
    }
  }

  // Get socket cookie
  cookie = bpf_get_socket_cookie(ctx);
  debugk(
      "cgroup/connect{4,6}: cookie: %d, ip: %x:%x:%x:%x, port: %d",
      cookie,
      dst_ip[0],
      dst_ip[1],
      dst_ip[2],
      dst_ip[4],
      ctx->user_port);

  // Extract destination information
  {
    // Construct value
    struct original_dst_info original_dst = {
        .ip = {
            dst_ip[0],
            dst_ip[1],
            dst_ip[2],
            dst_ip[3],
        }, // FIXME
        .port = ctx->user_port,
        .family = ctx->family,
    };
    // Save in LRU hashmap
    if (bpf_map_update_elem(&cookie_to_original_dst_map, &cookie, &original_dst, BPF_ANY))
    {
      errork("Failed updating original dst for socket with cookie %d", cookie);
    }
  }

  // Redirect
  {
    // Update redirect port
    ctx->user_port = bpf_htons(LIGHTPATH_REDIRECT_PORT);
    // Update dst IP
    switch (ctx->user_family)
    {
    case AF_INET:
      // IPv4 redirect
      ctx->user_ip4 = LIGHTPATH_IP4_IP;
      break;
    default:
      // IPv6
      ctx->user_ip6[0] = LIGHTPATH_IP6_IP_0;
      ctx->user_ip6[1] = LIGHTPATH_IP6_IP_1;
      ctx->user_ip6[2] = LIGHTPATH_IP6_IP_2;
      ctx->user_ip6[3] = LIGHTPATH_IP6_IP_3;
      break;
    }
  }

  return 0;
}