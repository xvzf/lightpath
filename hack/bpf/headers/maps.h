#pragma once

#include "common.h"
#include "config.h"

// BPF functions (for the compiler to not complain)
static void *(*bpf_map_lookup_elem)(void *map, const void *key) = (void *)BPF_FUNC_map_lookup_elem;
static __u64 (*bpf_map_update_elem)(void *map, const void *key, const void *value, __u64 flags) = (void *)BPF_FUNC_map_update_elem;
static __u64 (*bpf_map_delete_elem)(void *map, const void *key) = (void *)BPF_FUNC_map_delete_elem;
static __u64 (*bpf_get_socket_cookie)(void *ctx) = (void *)BPF_FUNC_get_socket_cookie;
static __u32 (*bpf_sock_hash_update)(void *skt, void *map, void *key, __u64 flags) = (void *)BPF_FUNC_sock_hash_update;

// IPv4 cluster ip inclusion list; this map is dual stack!
struct bpf_map_def __attribute__((section("maps"), used)) cluster_ip_inclusion_map = {
    .type = BPF_MAP_TYPE_HASH,
    .key_size = sizeof(__u32[4]), // IPv6 compliant
    .value_size = sizeof(__u64),  // Dummy value, used as O(1) lookup
    .max_entries = CLUSTER_IP_COUNT_MAX,
};

struct bpf_map_def __attribute__((section("maps"), used)) cookie_to_original_dst_map = {
    .type = BPF_MAP_TYPE_LRU_HASH, // LRU replace strategy
    .key_size = sizeof(__u64),
    .value_size = sizeof(struct original_dst_info),
    .max_entries = CONNTRACK_MAX_ENTRIES,
};

struct bpf_map_def __attribute__((section("maps"), used)) conn_pair_to_original_dst_map = {
    .type = BPF_MAP_TYPE_LRU_HASH, // LRU replace strategy
    .key_size = sizeof(struct conn_pair),
    .value_size = sizeof(struct original_dst_info),
    .max_entries = CONNTRACK_MAX_ENTRIES,
};

struct bpf_map_def __attribute__((section("maps"), used)) conn_pair_to_sock_map = {
    .type = BPF_MAP_TYPE_SOCKHASH,
    .key_size = sizeof(struct conn_pair),
    .value_size = sizeof(__u32),
    .max_entries = CONNTRACK_MAX_ENTRIES,
};
