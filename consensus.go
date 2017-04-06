package externalip

import (
	"net"
	"net/http"
	"time"
)

// DefaultConsensusConfig returns the ConsensusConfig,
// with the default values:
//    + Timeout: 5 seconds;
func DefaultConsensusConfig() *ConsensusConfig {
	return &ConsensusConfig{
		Timeout: time.Second * 5,
	}
}

// DefaultConsensus returns a consensus filled
// with default and recommended HTTPSources.
// TLS-Protected providers get more power,
// compared to plain-text providers.
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

// NewConsensus creates a new Consensus, with no sources.
// When the given cfg is <nil>, the `DefaultConsensusConfig` will be used.
func NewConsensus(cfg *ConsensusConfig) *Consensus {
	if cfg == nil {
		cfg = DefaultConsensusConfig()
	}
	return &Consensus{
		client: &http.Client{Timeout: cfg.Timeout},
	}
}

// ConsensusConfig is used to configure the Consensus, while creating it.
type ConsensusConfig struct {
	Timeout time.Duration
}

// WithTimeout sets the timeout of this config,
// returning the config itself at the end, to allow for chaining
func (cfg *ConsensusConfig) WithTimeout(timeout time.Duration) *ConsensusConfig {
	cfg.Timeout = timeout
	return cfg
}

// Consensus the type at the center of this library,
// and is the main entry point for users.
// Its `ExternalIP` method allows you to ask for your ExternalIP,
// influenced by all its added voters.
type Consensus struct {
	voters []voter
	client *http.Client
}

// AddVoter adds a voter to this consensus.
// The source cannot be <nil> and
// the weight has to be of a value of 1 or above.
func (c *Consensus) AddVoter(source Source, weight uint) error {
	if source == nil {
		return ErrNoSource
	}
	if weight == 0 {
		return ErrInsufficientWeight
	}

	c.voters = append(c.voters, voter{
		source: source,
		weight: weight,
	})
	return nil
}

// AddHTTPVoter creates and adds an HTTP Voter to this consensus,
// using the HTTP Client of this Consensus, configured by the ConsensusConfig.
func (c *Consensus) AddHTTPVoter(url string, weight uint) error {
	return c.AddVoter(NewHTTPSource(c.client, url), weight)
}

// AddComplexHTTPVoter creates an adds an HTTP Voter to this consensus,
// using a given parser, and the HTTP Client of this Consensus,
// configured by the ConsensusConfig
func (c *Consensus) AddComplexHTTPVoter(url string, parser ContentParser, weight uint) error {
	return c.AddVoter(
		NewHTTPSource(c.client, url).WithParser(parser),
		weight,
	)
}

// ExternalIP requests asynchronously the externalIP from all added voters,
// returning the IP which received the most votes.
// The returned IP will always be valid, in case the returned error is <nil>.
func (c *Consensus) ExternalIP() (net.IP, error) {
	voteCollection := make(map[string]uint)
	ch := make(chan vote, len(c.voters))

	// start all source Requests on a seperate goroutine
	for _, v := range c.voters {
		go func(v voter) {
			ip, err := v.source.IP()
			if err == nil && ip == nil {
				err = InvalidIPError("")
			}
			ch <- vote{
				IP:    ip,
				Count: v.weight,
				Error: err,
			}
		}(v)
	}

	// Wait for all votes to come in
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

	// if no votes were casted succesfully,
	// return early with an error
	if len(voteCollection) == 0 {
		return nil, ErrNoIP
	}

	var max uint
	var externalIP string

	// find the IP which has received the most votes,
	// influinced by the voter's weight.
	for ip, votes := range voteCollection {
		if votes > max {
			max, externalIP = votes, ip
		}
	}

	// as the found IP was parsed previously,
	// we know it cannot be nil and is valid
	return net.ParseIP(externalIP), nil
}
