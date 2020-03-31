package main

import (
	"flag"
	"log"
	"os"

	"github.com/bitsbeats/dronetrigger/config"
	"github.com/bitsbeats/dronetrigger/drone"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	ref := flag.String("ref", "", "Git ref to trigger build.")
	repo := flag.String("repo", "", "Repository to build (i.e. octocat/awesome).")
	configFile := flag.String("config", "/etc/dronetrigger.yml", "Configuration file.")
	verbose := flag.Bool("v", false, "Verbose output.")
	flag.Parse()

	if *repo == "" {
		log.Print("dronetrigger\n\n")
		flag.PrintDefaults()
		log.Fatal("\nplease specify a repository.")
	}

	c, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	d := drone.New(c.Url, c.Token)
	build, err := d.RebuildLastBuild(*repo, *ref)
	if err != nil {
		log.Fatal(err)
	}
	if *verbose {
		log.Printf("started build sha %s for %s", build.After, *repo)
	}
}
