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

web:
  bearer_token:
    octocat/test: s3cret_t0ken
```

* `url` represents the URL to a drone server
* `token` is used to authentificate against drone
* `web.bearer_token.*`: sets up a per repo secret to trigger builds


## Usage

Note: After adding new repos you need to sync the repositories, otherwise
you may get 404 errors. Also make sure that the access rights are configured
for the token.

```sh
DRONE_TOKEN="$token_from_etc_drontrigger_yml" drone repo sync
```

CLI examples:

```sh
# build default branch of a repo
dronetrigger -repo octocat/test

# build specific branch of a repo
dronetrigger -repo octocat/test -branch master

# rebuild a release
dronetigger -repo octocat/test -release
```

Web examples:

```sh
# rebuild last commit on a branch
curl -H 'Authorization: Bearer s3cret_token' -d '{"repo": "octocat/test", "branch": "master"}' $url

# rebuild last tag
curl -H 'Authorization: Bearer s3cret_token' -d '{"repo": "octocat/test", "release": true}' $url

# promote last commit on a branch
curl -H 'Authorization: Bearer s3cret_token' -d '{"repo": "octocat/test", "branch": "master", "target": "promote-name"}' $url

# promote last tag
curl -H 'Authorization: Bearer s3cret_token' -d '{"repo": "octocat/test", "release": true, "target": "promote-name"}' $url
```

Help:

```sh
$ dronetrigger -h
Usage of ./dronetrigger:
  -config string
    	Configuration file. (default "/etc/dronetrigger.yml")
  -branch string
    	Git rev (i.e. branch) to trigger build.
  -repo string
    	Repository to build (i.e. octocat/awesome).
```
