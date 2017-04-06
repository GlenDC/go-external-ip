package externalip

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type HTTPSource struct {
	client *http.Client
	url    string
	parser ContentParser
}

type ContentParser func(string) (string, error)

func NewHTTPSource(client *http.Client, url string) *HTTPSource {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPSource{
		client: client,
		url:    url,
	}
}

func (s *HTTPSource) WithParser(parser ContentParser) *HTTPSource {
	s.parser = parser
	return s
}

func (s *HTTPSource) IP() (net.IP, error) {
	req, err := http.NewRequest("GET", s.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "go-external-ip (github.com/glendc/go-external-ip)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	raw := string(bytes)
	if s.parser != nil {
		raw, err = s.parser(raw)
		if err != nil {
			return nil, err
		}
	}

	externalIP := net.ParseIP(strings.TrimSpace(raw))
	if externalIP == nil {
		return nil, InvalidIPError(raw)
	}

	return externalIP, nil
}
