package externalip

import "net"

// Source defines the part of a voter which gives the actual voting value (IP).
type Source interface {
	// IP returns IPv4/IPv6 address in a non-error case
	// net.IP should never be <nil> when error is <nil>
	// NOTE: it is important that IP doesn't block indefinitely,
	//   as the entire Consensus Logic will be blocked indefinitely as well
	//   if this happens.
	IP() (net.IP, error)
}

// voter adds weight to the IP given by a source.
// The weight has to be at least 1, and the more it is, the more power the voter has.
type voter struct {
	source Source // provides the IP (see: vote)
	weight uint   // provides the weight of its vote (acts as a multiplier)
}

// vote is given by each voter,
// if the Error is not <nil>, the IP and Count values are ignored,
// and the vote has no effect.
// The IP value should never be <nil>, when Error is <nil> as well.
type vote struct {
	IP    net.IP // the IP proposed by the Voter in question
	Count uint   // equal to the Voter's weight
	Error error  // defines if the Vote was cast succesfully or not
}
