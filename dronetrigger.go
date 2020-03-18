package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/bitsbeats/dronetrigger/drone"
	"gopkg.in/yaml.v2"
)

type (
	config struct {
		Url   string
		Token string
	}
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	ref := flag.String("ref", "", "Git rev (i.e. branch) to trigger build.")
	repo := flag.String("repo", "", "Repository to build (i.e. octocat/awesome).")
	configFile := flag.String("config", "/etc/dronetrigger.yml", "Configuration file.")
	verbose := flag.Bool("v", false, "Verbose output.")
	flag.Parse()

	if *repo == "" {
		log.Print("dronetrigger\n\n")
		flag.PrintDefaults()
		log.Fatal("\nplease specify a repository.")
	}

	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	c := config{}
	err = yaml.Unmarshal(configData, &c)
	if err != nil {
		log.Fatal(err)
	}

	d := drone.New(c.Url, c.Token)
	lastBuild, err := d.LastBuild(*repo, *ref)
	if err != nil {
		log.Fatalf("unable to get last build: %v", err)
	}
	build, err := d.Trigger(*repo, lastBuild.After, lastBuild.Source)
	if err != nil {
		log.Fatal(err)
	}
	if *verbose {
		log.Printf("started build sha %s for %s", build.After, *repo)
	}
}
