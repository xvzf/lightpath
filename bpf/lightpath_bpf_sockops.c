#include "headers/common.h"
#include "headers/logger.h"
#include "headers/maps.h"
#include <linux/in.h>
#include <linux/socket.h>

__attribute__((section("sockops"), used)) int
lightpath_sock_connect(struct bpf_sock_ops *ctx) {
  __u64 cookie;
  struct original_dst_info *orig_dst;

  // We're just interested in those two events, drop the rest
  if (ctx->op != BPF_SOCK_OPS_PASSIVE_ESTABLISHED_CB &&
      ctx->op != BPF_SOCK_OPS_ACTIVE_ESTABLISHED_CB) {
    return 0;
  }

  // Filter for IPv4&IPv6
  if (ctx->family != AF_INET && ctx->family != AF_INET6) {
    return 0;
  }

  // retrieve cookie
  cookie = bpf_get_socket_cookie(ctx);

  // Lookup original destination
  orig_dst = bpf_map_lookup_elem(&cookie_to_original_dst_map, &cookie);
  if (!orig_dst) {
    // No cookie exists -> either we're not interested in this connection or the
    // socket connection failed
    return 0;
  }

  // Build conn pari
  struct conn_pair conn = {
      .family = ctx->family,
      // Src (convert to network byte order)
      .src_port = bpf_htons(ctx->local_port),
      // ctx->remote_port is 16 bit, variable is 32bit in network endian
      .dst_port = ctx->remote_port >> 16,
  };

  // Extract IPs
  if (ctx->family == AF_INET) { // IPv4
    // dst ip
    conn.dst_ip[0] = 0;
    conn.dst_ip[1] = 0;
    conn.dst_ip[2] = 0;
    conn.dst_ip[3] = ctx->local_ip4;

    // src ip
    conn.src_ip[0] = 0;
    conn.src_ip[1] = 0;
    conn.src_ip[2] = 0;
    conn.src_ip[3] = ctx->remote_ip4;
  } else { // IPv6
    // dst ip
    conn.dst_ip[0] = ctx->remote_ip6[0];
    conn.dst_ip[1] = ctx->remote_ip6[1];
    conn.dst_ip[2] = ctx->remote_ip6[2];
    conn.dst_ip[3] = ctx->remote_ip6[3];
    // src ip
    conn.src_ip[0] = ctx->local_ip6[0];
    conn.src_ip[1] = ctx->local_ip6[1];
    conn.src_ip[2] = ctx->local_ip6[2];
    conn.src_ip[3] = ctx->local_ip6[3];
  }

  // update lookup table for original dst
  if (!bpf_map_update_elem(&conn_pair_to_original_dst_map, &conn, orig_dst,
                           BPF_ANY)) {
    errork("Failed updating original dst for connection pair");
  }

  // Update socket pair map
  if (!(bpf_sock_hash_update(ctx, &conn_pair_to_sock_map, &conn,
                             BPF_NOEXIST))) {
    errork("Failed updating sockethash map");
  }

  return 0;
}