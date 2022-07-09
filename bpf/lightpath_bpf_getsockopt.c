#include <linux/in.h>

#include "headers/common.h"
#include "headers/maps.h"
#include "headers/logger.h"

#define SO_ORIGINAL_DST 80

__attribute__((section("getsockopt"), used)) int lightpath_getsocketopt(struct bpf_sockopt *ctx)
{

  struct original_dst_info *original_dst;
  struct conn_pair conn;
  struct sockaddr_in socket_addr4;
  struct sockaddr_in6 socket_addr6;

  // Check if this is a request to get the original DST
  if (ctx->optname != SO_ORIGINAL_DST)
  {
    return 0;
  }

  // Build conn pair
  conn.family = ctx->sk->family;
  // Swap SRC and DST around
  conn.src_port = ctx->sk->dst_port;
  conn.dst_port = bpf_htons(ctx->sk->src_port);

  // Extract IPs
  if (ctx->sk->family == AF_INET)
  { // IPv4
    // dst ip
    conn.dst_ip[0] = 0;
    conn.dst_ip[1] = 0;
    conn.dst_ip[2] = 0;
    conn.dst_ip[3] = ctx->sk->src_ip4;

    // src ip
    conn.src_ip[0] = 0;
    conn.src_ip[1] = 0;
    conn.src_ip[2] = 0;
    conn.src_ip[3] = ctx->sk->dst_ip4;
  }
  else
  { // IPv6
    // dst ip
    memcpy(&conn.dst_ip, ctx->sk->src_ip6, 4 * sizeof(__u32));
    // src ip
    memcpy(&conn.src_ip, ctx->sk->dst_ip6, 4 * sizeof(__u32));
  }

  // Retrieve original destination infos
  original_dst = bpf_map_lookup_elem(&conn_pair_to_original_dst_map, &conn);
  if (!original_dst)
  {
    return 0;
  }

  // Check requirements for rewrite
  if (conn.family == AF_INET)
  { // IPv4
    ctx->optlen = (__s32)sizeof(struct sockaddr_in);

    if ((void *)((struct sockaddr_in *)ctx->optval + 1) <= ctx->optval_end)
    {
      errork("Optname SO_ORIGINAL_DST invalid getsockopt optval size (AF_INET)");
      return 1;
    }

    // Fill ipv4 response buffer
    socket_addr4.sin_family = ctx->sk->family;
    socket_addr4.sin_addr.s_addr = original_dst->ip[3];
    socket_addr4.sin_port = original_dst->port;

    // Set value
    *(struct sockaddr_in *)ctx->optval = socket_addr4;
  }
  else
  { // IPv6

    ctx->optlen = (__s32)sizeof(struct sockaddr_in6);

    if ((void *)((struct sockaddr_in6 *)ctx->optval + 1) <= ctx->optval_end)
    {
      errork("Optname SO_ORIGINAL_DST invalid getsockopt optval size (AF_INET6)");
      return 1;
    }

    // Fill ipv6 response buffer
    socket_addr6.sin6_family = ctx->sk->family;
    memcpy(&socket_addr6.sin6_addr.in6_u, original_dst->ip, 4 * sizeof(__u32));
    socket_addr6.sin6_port = original_dst->port;

    // Set value
    *(struct sockaddr_in6 *)ctx->optval = socket_addr6;
  }

  return 0;
}