#include "headers/common.h"
#include "headers/logger.h"
#include "headers/maps.h"
#include <linux/in.h>

__attribute__((section("sk_msg"), used)) int
lightpath_redir(struct sk_msg_md *ctx) {
  struct conn_pair conn;

  // Build conn pair
  conn.family = ctx->sk->family;
  // Swap SRC and DST around
  conn.src_port = ctx->sk->dst_port;
  conn.dst_port = bpf_htons(ctx->sk->src_port);

  // Extract IPs
  if (ctx->sk->family == AF_INET) { // IPv4
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
  } else { // IPv6
    // dst ip
    memcpy(&conn.dst_ip, ctx->sk->src_ip6, 4 * sizeof(__u32));
    // src ip
    memcpy(&conn.src_ip, ctx->sk->dst_ip6, 4 * sizeof(__u32));
  }

  // Lookup socket for pair
  if (bpf_msg_redirect_hash(ctx, &conn_pair_to_sock_map, &conn,
                            BPF_F_INGRESS)) {
    debugk("redirect %d bytes", ctx->size);
  }

  return 0;
}