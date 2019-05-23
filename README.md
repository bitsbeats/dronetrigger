# dronetrigger

Simple tool to trigger drone builds. Can be used for more advanced crons or in scripts.

## Configuration

The configfile is either in the working directory as `dronetrigger.yml` or supplied by `-config`.

Sample:

```yaml
---

url: https://drone.example.com
token: thisisnotavaliddronetoken1234567
```

## Usage

Examples:

```sh
# build default branch of a repo
dronetrigger -repo octocat/test

# build specific branch of a repo
dronetrigger -repo octocat/test -ref refs/heads/master
```

Help:

```sh
$ ./dronetrigger -h
Usage of ./dronetrigger:
  -config string
    	Configuration file, defaults to dronetrigger.yml in workdir. (default "dronetrigger.yml")
  -ref string
    	Git rev (i.e. branch) to trigger build.
  -repo string
    	Repository to build (i.e. octocat/awesome).
```
