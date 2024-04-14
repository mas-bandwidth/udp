
KERNEL = $(shell uname -r)

.PHONY: build
build: server.c server_xdp.o
	gcc -O2 -g server.c -o server -lxdp /usr/src/linux-headers-$(KERNEL)/tools/bpf/resolve_btfids/libbpf/libbpf.a -lz -lelf

server_xdp.o: server_xdp.c
	clang -O2 -g -Ilibbpf/src -target bpf -c server_xdp.c -o server_xdp.o

.PHONY: clean
clean:
	rm -f server
	rm -f *.o