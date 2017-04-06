package externalip

import "net"

type Source interface {
	// IP returns IPv4/IPv6 address in a non-error case
	IP() (net.IP, error)
}

type Voter struct {
	source Source // provides the IP (see: vote)
	weight uint   // provides the weight of its vote (acts as a multiplier)
}

type Vote struct {
	IP    net.IP
	Count uint
	Error error
}
