NS_A = ip netns exec nsa
NS_A = ip netns exec nsa

all: run

run: setup main
	ip netns exec ns_a ./main t1a
	# ip netns exec ns_b ./main t1b

bpf/bpf_bpfeb.go: $(wildcard bpf/kern/*)
	go generate ./...

main: bpf/bpf_bpfeb.go $(wildcard *.go bpf/*.go)
	go build -o $@ ./

setup:
	ip netns del ns_a || true
	ip netns del ns_b || true

	ip netns add ns_a
	ip netns add ns_b

	ip link add t1a netns ns_a type veth peer t1b netns ns_b

	ip -n ns_a link set dev t1a up
	ip -n ns_b link set dev t1b up

	ip -n ns_a addr add 10.0.0.1/24 dev t1a
	ip -n ns_b addr add 10.0.0.2/24 dev t1b

show:
	tc -n ns_b filter show dev t1b ingress
	tc -n ns_b filter show dev t1b egress

send:
	echo "hello1234" | ip netns exec ns_a nc -p 3333 -u 10.0.0.2 1234

recv:
	ip netns exec ns_b nc -lu 1234

dump-a:
	ip netns exec ns_a tshark \
		-o udp.check_checksum:TRUE \
		-o ip.check_checksum:TRUE \
		-i t1a -Vx

dump-b:
	ip netns exec ns_b tshark \
		-o udp.check_checksum:TRUE \
		-o ip.check_checksum:TRUE \
		-i t1b -Vx