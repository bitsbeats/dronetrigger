# dronetrigger

Simple tool to trigger drone builds. Can be used for more advanced crons or in scripts.

**Note**: dronetrigger injects the variable `DRONETRIGGER=true` into the steps.
Use this to determine if a build was triggered.

## Configuration

The configfile is either `/etc/dronetrigger.yml` or supplied by `-config`.

Sample:

```yaml
---

url: https://drone.example.com
token: thisisnotavaliddronetoken1234567

web:  # only required if you are using dronetrigger-web
  bearer_token: secret_bearer_token
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
$ dronetrigger -h
Usage of ./dronetrigger:
  -config string
    	Configuration file. (default "/etc/dronetrigger.yml")
  -ref string
    	Git rev (i.e. branch) to trigger build.
  -repo string
    	Repository to build (i.e. octocat/awesome).
```
