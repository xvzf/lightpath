#pragma once

#include <linux/bpf.h>

static void (*bpf_trace_printk)(const char *fmt, int fmt_size,
                                ...) = (void *)BPF_FUNC_trace_printk;

#ifndef errork
#define errork(fmt, ...)                                 \
  ({                                                     \
    char _fmt[] = "[lightpath][error]" fmt;              \
    bpf_trace_printk(_fmt, sizeof(_fmt), ##__VA_ARGS__); \
  })
#endif

#ifndef debugk
#ifndef DEBUG
#define debugk(fmt, ...)                                 \
  ({                                                     \
    char _fmt[] = "[lightpath][debug]" fmt;              \
    bpf_trace_printk(_fmt, sizeof(_fmt), ##__VA_ARGS__); \
  })
#else
#define debugk(fmt, ...) ({})
#endif
#endif