package main

import (
	"flag"
	"io/ioutil"
	"log"

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
	ref := flag.String("ref", "", "Git rev (i.e. branch) to trigger build.")
	repo := flag.String("repo", "", "Repository to build (i.e. octocat/awesome).")
	configFile := flag.String("config", "dronetrigger.yml", "Configuration file, defaults to dronetrigger.yml in workdir.")
	flag.Parse()

	if *repo == "" {
		log.Fatal("please specify a repository (see -h)")
	}

	log.SetFlags(0)
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
	build, err := d.Trigger(*repo, lastBuild.After)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("started build sha %s for %s", build.After, *repo)
}