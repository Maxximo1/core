package npp

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"github.com/libp2p/go-reuseport"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/netutil"
)

const protocol = "tcp"

type Port uint16

func DialContext(ctx context.Context, network, laddr, raddr string) (net.Conn, error) {
	if !reuseport.Available() {
		return nil, syscall.ENOPROTOOPT
	}

	var dialer reuseport.Dialer
	if laddr != "" {
		localAddr, err := reuseport.ResolveAddr(network, laddr)
		if err != nil {
			return nil, err
		}
		dialer.D.LocalAddr = localAddr
	}

	return dialer.DialContext(ctx, network, raddr)
}

// PrivateAddrs collects and returns private addresses of a network interfaces
// the socket bind on.
func privateAddrs(addr net.Addr) ([]net.Addr, error) {
	ip, port, err := netutil.SplitHostPort(addr.String())
	if err != nil {
		return nil, err
	}

	ips, err := util.GetAvailableIPs()
	if err != nil {
		return nil, err
	}

	if !ip.IsUnspecified() {
		ips = filteredIPs(ips, ip)
	}

	var addrs []net.Addr
	for _, ip := range ips {
		addr, err := net.ResolveTCPAddr(protocol, fmt.Sprintf("%s:%d", ip, uint16(port)))
		if err != nil {
			return nil, err
		}

		addrs = append(addrs, addr)
	}

	return addrs, nil
}

func filteredIPs(ips []net.IP, target net.IP) []net.IP {
	var filtered []net.IP
	for _, ip := range ips {
		if ip.Equal(target) {
			filtered = append(filtered, ip)
		}
	}

	return filtered
}
