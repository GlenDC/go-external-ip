package externalip

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

// NewHTTPSource creates a HTTP Source object,
// which can be used to request the (external) IP from.
// The Default HTTP Client will be used if no client is given.
func NewHTTPSource(url string) *HTTPSource {
	return &HTTPSource{
		url: url,
	}
}

// HTTPSource is the default source, to get the external IP from.
// It does so by requesting the IP from a URL, via an HTTP GET Request.
type HTTPSource struct {
	url    string
	parser ContentParser
}

// ContentParser can be used to add a parser to an HTTPSource
// to parse the raw content returned from a website, and return the IP.
// Spacing before and after the IP will be trimmed by the Consensus.
type ContentParser func(string) (string, error)

// WithParser sets the parser value as the value to be used by this HTTPSource,
// and returns the pointer to this source, to allow for chaining.
func (s *HTTPSource) WithParser(parser ContentParser) *HTTPSource {
	s.parser = parser
	return s
}

// IP implements Source.IP
func (s *HTTPSource) IP(timeout time.Duration) (net.IP, error) {
	// Define the GET method with the correct url,
	// setting the User-Agent to our library
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "go-external-ip (github.com/glendc/go-external-ip)")

	client := &http.Client{Timeout: timeout}
	// Do the request and read the body for non-error results.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// optionally parse the content
	raw := string(bytes)
	if s.parser != nil {
		raw, err = s.parser(raw)
		if err != nil {
			return nil, err
		}
	}

	// validate the IP
	externalIP := net.ParseIP(strings.TrimSpace(raw))
	if externalIP == nil {
		return nil, InvalidIPError(raw)
	}

	// returned the parsed IP
	return externalIP, nil
}
