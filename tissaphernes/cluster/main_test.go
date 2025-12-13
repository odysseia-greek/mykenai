package cluster

import (
	"log"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("failed to load env config: %v", err)
	}

	testenv = env.NewWithConfig(cfg)

	// No Setup() for now: we are *not* creating clusters/namespaces yet.
	// Just run against the target cluster.
	os.Exit(testenv.Run(m))
}
