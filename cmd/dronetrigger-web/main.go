package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/bitsbeats/dronetrigger/config"
	"github.com/bitsbeats/dronetrigger/drone"
	"github.com/bitsbeats/dronetrigger/web"
)

func main() {
	// parse flags
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	configFile := flag.String("config", "/etc/dronetrigger.yml", "Configuration file.")
	flag.Parse()

	// load and validate config
	c, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("unable to load config: %s", err)
	}
	if c.Web == nil {
		log.Fatalf("no configuration for web found")
	}
	for repo, token := range c.Web.BearerToken {
		if len(token) < 8 {
			log.Fatalf("configured bearer token for %q is to short", repo)
		}
	}

	// setup drone
	d := drone.New(c.Url, c.Token)

	// configure webserver
	w := web.NewWeb(c.Web, d)
	mux := http.NewServeMux()
	mux.HandleFunc("/", w.Handle)
	middlewared := w.Middleware(mux)

	// listen
	log.Printf("listening on %s", c.Web.Listen)
	err = http.ListenAndServe(c.Web.Listen, middlewared)
	if err != nil {
		log.Fatalf("webserver stopped: %s", err)
	}
}
