package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/odysseia-greek/agora/plato/logging"
	"github.com/odysseia-greek/mykenai/lysandros/strategos"
)

const standardPort = ":8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = standardPort
	}

	//https://patorjk.com/software/taag/#p=display&f=Crawford2&t=LYSANDROS&x=none&v=4&h=4&w=80&we=false
	logging.System(`
 _      __ __  _____  ____  ____   ___    ____   ___   _____
| |    |  |  |/ ___/ /    ||    \ |   \  |    \ /   \ / ___/
| |    |  |  (   \_ |  o  ||  _  ||    \ |  D  )     (   \_ 
| |___ |  ~  |\__  ||     ||  |  ||  D  ||    /|  O  |\__  |
|     ||___, |/  \ ||  _  ||  |  ||     ||    \|     |/  \ |
|     ||     |\    ||  |  ||  |  ||     ||  .  \     |\    |
|_____||____/  \___||__|__||__|__||_____||__|\_|\___/  \___|
`)
	logging.System("\"λέγεται δὲ ὁ Λυσάνδρου πατὴρ Ἀριστόκλειτος οἰκίας μὲν οὐ γενέσθαι βασιλικῆς, ἄλλως δὲ γένους εἶναι τοῦ τῶν Ἡρακλειδῶν\"")
	logging.System("\"The father of Lysander, Aristocleitus, is said to have been of the lineage of the Heracleidae, though not of the royal family.\"")
	logging.System("starting html viewer.....")

	logging.System("getting env variables and creating config")

	lysandrosConfig, err := strategos.CreateNewConfig()
	if err != nil {
		log.Fatal("death has found me")
	}

	srv := strategos.InitRoutes(lysandrosConfig)

	logging.System(fmt.Sprintf("%s : %s", "running on port", port))
	err = http.ListenAndServe(port, srv)
	if err != nil {
		panic(err)
	}
}
