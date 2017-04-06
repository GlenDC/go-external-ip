package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/glendc/go-external-ip"
)

// CLI Flags
var (
	timeout = flag.Duration("t", time.Second*2, "consensus's voting timeout")
	verbose = flag.Bool("v", false, "verbose logging")
)

func main() {
	// configure the consensus
	cfg := externalip.DefaultConsensusConfig()
	if timeout != nil {
		cfg.WithTimeout(*timeout)
	}

	// TODO: Add Logging (and use the verbose flag)

	// create the consensus
	consensus := externalip.DefaultConsensus(
		externalip.DefaultConsensusConfig().WithTimeout(*timeout))

	// retrieve the external ip
	ip, err := consensus.ExternalIP()

	// simple error handling
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// success, simply output the IP in string format
	fmt.Println(ip.String())
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
