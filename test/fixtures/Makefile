# Test fixtures Makefile
# Build C test programs for procscope integration tests.

CC      ?= cc
CFLAGS  := -Wall -Wextra -O1

TARGETS := fork_exec file_ops net_connect

.PHONY: all clean

all: $(TARGETS)

fork_exec: fork_exec.c
	$(CC) $(CFLAGS) -o $@ $<

file_ops: file_ops.c
	$(CC) $(CFLAGS) -o $@ $<

net_connect: net_connect.c
	$(CC) $(CFLAGS) -o $@ $<

clean:
	rm -f $(TARGETS)
