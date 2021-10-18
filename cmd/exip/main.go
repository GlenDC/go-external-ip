package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	externalip "github.com/glendc/go-external-ip"
)

// CLI Flags
var (
	timeout  = flag.Duration("t", time.Second*5, "consensus's voting timeout")
	verbose  = flag.Bool("v", false, "log errors to STDERR, when defined")
	protocol = flag.Uint("p", 0, "IP Protocol to be used (0, 4, or 6)")
)

func main() {
	// configure the consensus
	cfg := externalip.DefaultConsensusConfig()
	if timeout != nil {
		cfg.WithTimeout(*timeout)
	}

	// optionally create the logger,
	// if no logger is defined, all logs will be discarded.
	var logger *log.Logger
	if verbose != nil && *verbose {
		logger = externalip.NewLogger(os.Stderr)
	}

	// create the consensus
	consensus := externalip.DefaultConsensus(cfg, logger)
	err := consensus.UseIPProtocol(*protocol)
	errCheck(err)

	// retrieve the external ip
	ip, err := consensus.ExternalIP()
	errCheck(err)

	// success, simply output the IP in string format
	fmt.Println(ip.String())
}

func errCheck(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Define customized usage output
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, "Retrieve your external IP.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n    %s [flags]\n\n", os.Args[0])
		fmt.Fprintln(os.Stderr, "Flags:")
		fmt.Fprintf(os.Stderr, "  -h help\n    \tshow this usage message\n")
		flag.PrintDefaults()
	}

	// Parse CLI Flags
	flag.Parse()
}
