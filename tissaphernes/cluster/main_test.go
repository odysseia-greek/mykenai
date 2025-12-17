package cluster

import (
	"log"
	"os"
	"testing"

	"github.com/odysseia-greek/agora/plato/logging"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	//https://patorjk.com/software/taag/#p=display&f=Crawford2&t=MELETOS
	logging.System(`
 ______  ____ _____ _____  ____  ____  __ __    ___  ____   ____     ___  _____
|      ||    / ___// ___/ /    ||    \|  |  |  /  _]|    \ |    \   /  _]/ ___/
|      | |  (   \_(   \_ |  o  ||  o  )  |  | /  [_ |  D  )|  _  | /  [_(   \_ 
|_|  |_| |  |\__  |\__  ||     ||   _/|  _  ||    _]|    / |  |  ||    _]\__  |
  |  |   |  |/  \ |/  \ ||  _  ||  |  |  |  ||   [_ |    \ |  |  ||   [_ /  \ |
  |  |   |  |\    |\    ||  |  ||  |  |  |  ||     ||  .  \|  |  ||     |\    |
  |__|  |____|\___| \___||__|__||__|  |__|__||_____||__|\_||__|__||_____| \___|
`)
	logging.System("\"καὶ διενοεῖτο τὸ πλέον οὕτως ὁ Τισσαφέρνης, ὅσα γε ἀπὸ τῶν ποιουμένων ἦν εἰκάσαι.\"")
	logging.System("\"In the main Tissaphernes approved of this policy, so far at least as could be conjectured from his behaviour.\"")
	// Thucydides, History of the Peloponnesian War 8.46.5
	logging.System("starting test suite setup.....")

	logging.System("getting env variables and creating config")

	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("failed to load env config: %v", err)
	}

	testenv = env.NewWithConfig(cfg)

	// No Setup() for now: we are *not* creating clusters/namespaces yet.
	// Just run against the target cluster.
	os.Exit(testenv.Run(m))
}
