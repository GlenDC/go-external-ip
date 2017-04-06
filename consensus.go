package externalip

import (
	"net"
	"net/http"
	"time"
)

func DefaultConsensusConfig() *ConsensusConfig {
	return &ConsensusConfig{
		Timeout: time.Second * 5,
	}
}

func DefaultConsensus(cfg *ConsensusConfig) *Consensus {
	consensus := NewConsensus(cfg)

	// TLS-protected providers
	consensus.AddHTTPVoter("https://icanhazip.com/", 3)
	consensus.AddHTTPVoter("https://myexternalip.com/raw", 3)

	// Plain-text providers
	consensus.AddHTTPVoter("http://ifconfig.io/ip", 1)
	consensus.AddHTTPVoter("http://checkip.amazonaws.com/", 1)
	consensus.AddHTTPVoter("http://ident.me/", 1)
	consensus.AddHTTPVoter("http://whatismyip.akamai.com/", 1)
	consensus.AddHTTPVoter("http://tnx.nl/ip", 1)
	consensus.AddHTTPVoter("http://myip.dnsomatic.com/", 1)
	consensus.AddHTTPVoter("http://ipecho.net/plain", 1)
	consensus.AddHTTPVoter("http://diagnostic.opendns.com/myip", 1)

	return consensus
}

func NewConsensus(cfg *ConsensusConfig) *Consensus {
	if cfg == nil {
		cfg = DefaultConsensusConfig()
	}
	return &Consensus{
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

type ConsensusConfig struct {
	Timeout time.Duration
}

func (cfg *ConsensusConfig) WithTimout(timeout time.Duration) *ConsensusConfig {
	cfg.Timeout = timeout
	return cfg
}

type Consensus struct {
	voters []Voter
	client *http.Client
}

func (c *Consensus) AddVoter(source Source, weight uint) error {
	if source == nil {
		return NoSourceError
	}
	if weight == 0 {
		return InsufficientWeightError
	}

	c.voters = append(c.voters, Voter{
		source: source,
		weight: weight,
	})
	return nil
}

func (c *Consensus) AddHTTPVoter(url string, weight uint) error {
	return c.AddVoter(NewHTTPSource(c.client, url), weight)
}

func (c *Consensus) AddComplexHTTPVoter(url string, parser ContentParser, weight uint) error {
	return c.AddVoter(
		NewHTTPSource(c.client, url).WithParser(parser),
		weight,
	)
}

func (c *Consensus) ExternalIP() (net.IP, error) {
	voteCollection := make(map[string]uint)
	ch := make(chan Vote, len(c.voters))

	for _, voter := range c.voters {
		go func(voter Voter) {
			ip, err := voter.source.IP()
			ch <- Vote{
				IP:    ip,
				Count: voter.weight,
				Error: err,
			}
		}(voter)
	}

	var count int
	for count < len(c.voters) {
		select {
		case vote := <-ch:
			count++
			if vote.Error == nil {
				voteCollection[vote.IP.String()] += vote.Count
				continue
			}
		}
	}

	if len(voteCollection) == 0 {
		return nil, NoIPError
	}

	var max uint
	var externalIP string

	for ip, votes := range voteCollection {
		if votes > max {
			max, externalIP = votes, ip
		}
	}

	return net.ParseIP(externalIP), nil
}
