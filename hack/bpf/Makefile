CC=clang
CFLAGS=-O2 -g  -Wall -target bpf -I/usr/include/$(shell uname -m)-linux-gnu

SRCS=$(wildcard *.c)
OBJS=$(SRCS:.c=.o)

compile: $(OBJS)

clean:
	rm *.o

# Compile C code
%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@