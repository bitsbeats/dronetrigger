package main

import (
	"flag"
	"log"
	"os"

	"github.com/bitsbeats/dronetrigger/config"
	"github.com/bitsbeats/dronetrigger/core"
	"github.com/bitsbeats/dronetrigger/drone"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	branch := flag.String("branch", "", "Git branch to trigger build.")
	release := flag.Bool("release", false, "Rebuild last release tag. Mutally exclusive with -branch")
	repo := flag.String("repo", "", "Repository to build (i.e. octocat/awesome).")
	configFile := flag.String("config", "/etc/dronetrigger.yml", "Configuration file.")
	verbose := flag.Bool("v", false, "Verbose output.")
	flag.Parse()

	if *repo == "" {
		log.Print("dronetrigger\n\n")
		flag.PrintDefaults()
		log.Fatal("\nplease specify a repository.")
	}
	if *release && *branch != "" {
		flag.PrintDefaults()
		log.Fatal("unable to use -release with -branch")
	}

	c, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	d := drone.New(c.Url, c.Token)

	build := (*core.Build)(nil)
	if *release {
		build, err = d.RebuildLastTag(*repo)
	} else {
		build, err = d.RebuildLastBuild(*repo, *branch)
	}
	if err != nil {
		log.Fatal(err)
	}
	if *verbose {
		log.Printf("started build sha %s for %s", build.After, *repo)
	}
}
