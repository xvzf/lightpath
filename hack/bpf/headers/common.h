#pragma once

#include <linux/bpf.h>
#include <linux/bpf_common.h>
#include <linux/types.h>
#include <linux/ip.h>
#include <linux/ipv6.h>
#include <linux/swab.h>

#ifndef AF_INET
#define AF_INET 2
#endif
#ifndef AF_INET6
#define AF_INET6 10
#endif

#include <linux/swab.h>
#include <linux/types.h>

// Inspired by linux selftest code
#if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
#define bpf_htons(x) __builtin_bswap16(x)
#define bpf_htonl(x) __builtin_bswap32(x)
#elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define bpf_htons(x) (x)
#define bpf_htonl(x) (x)
#endif

// memcpy is available to eBPF but we cannot include string.h
void *memcpy(void *s1, const void *s2, unsigned long);

// Anonymous struct used by BPF_MAP_CREATE
struct bpf_map_def
{
  __u32 type;
  __u32 key_size;    // Key size in bytes
  __u32 value_size;  // Value size in bytes
  __u32 max_entries; // Max number of entries
  __u32 map_flags;   // Pre-allocate or not?
};

// origin information
struct original_dst_info
{
  __u32 ip[4];
  __u16 port;

  __u16 family;
};

// connection tracking pair
struct conn_pair
{
  __u32 src_ip[4];
  __u32 dst_ip[4];
  __u16 src_port;
  __u16 dst_port;

  __u16 family;
};