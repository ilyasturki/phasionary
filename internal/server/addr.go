package server

import "net"

// 100.64/10 (RFC 6598, Tailscale): reachable like a LAN, but IsPrivate misses it.
var cgnat = net.IPNet{IP: net.IPv4(100, 64, 0, 0).To4(), Mask: net.CIDRMask(10, 32)}

// LAN first: a VPN address needs that VPN running on the phone too.
func reachableAddresses() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var lan, vpn []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			n, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ip := n.IP.To4()
			switch {
			case ip == nil:
			case ip.IsPrivate():
				lan = append(lan, ip.String())
			case cgnat.Contains(ip):
				vpn = append(vpn, ip.String())
			}
		}
	}
	return append(lan, vpn...)
}

func advertisedHosts(host string) []string {
	if host != "" && host != "0.0.0.0" && host != "::" {
		return []string{host}
	}
	if lan := reachableAddresses(); len(lan) > 0 {
		return lan
	}
	return []string{"127.0.0.1"}
}
