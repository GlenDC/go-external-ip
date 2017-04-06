package externalip

import (
	"fmt"
	"testing"
)

func TestDefaultConsensus(t *testing.T) {
	consensus := DefaultConsensus(nil)
	if consensus == nil {
		t.Fatal("default consensus should never be nil")
	}

	ip, err := consensus.ExternalIP()
	if err != nil {
		t.Fatal("couldn't get external IP", err)
	}

	fmt.Println(ip)

	for i := 0; i < 2; i++ {
		ipAgain, err := consensus.ExternalIP()
		if err != nil {
			t.Fatal("couldn't get external IP", err)
		}
		if !ip.Equal(ipAgain) {
			t.Fatalf("expected %q, while received %q", ip, ipAgain)
		}
	}
}
