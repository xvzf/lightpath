#pragma once

#define DEBUG

#ifndef LIGHTPATH_IP4_IP
#define LIGHTPATH_IP4_IP 0xa9fefafa // 169.254.250.250 // Do not collide with e.g. GCP internal metadata serv ices
#endif

#ifndef LIGHTPATH_IP6_IP
#define LIGHTPATH_IP6_IP // FIXME find a solution
#define LIGHTPATH_IP6_IP_0 0x00
#define LIGHTPATH_IP6_IP_1 0x00
#define LIGHTPATH_IP6_IP_2 0x00
#define LIGHTPATH_IP6_IP_3 0x00
#endif

#ifndef LIGHTPATH_REDIRECT_PORT
#define LIGHTPATH_REDIRECT_PORT 1666 // Envoy listening port
#endif

#ifndef CLUSTER_IP_COUNT_MAX
#define CLUSTER_IP_COUNT_MAX 4096 // FIXME check what a sane limit is
#endif

#ifndef CONNTRACK_MAX_ENTRIES
#define CONNTRACK_MAX_ENTRIES 65535
#endif