# StackState Agent

[![CircleCI](https://circleci.com/gh/StackVista/stackstate-agent/tree/master.svg?style=svg)](https://circleci.com/gh/StackVista/stackstate-agent/tree/master)
[![Build status](https://ci.appveyor.com/api/projects/status/kcwhmlsc0oq3m49p/branch/master?svg=true)](https://ci.appveyor.com/project/StackVista/stackstate-agent/branch/master)
[![GoDoc](https://godoc.org/github.com/StackVista/stackstate-agent?status.svg)](https://godoc.org/github.com/StackVista/stackstate-agent)
[![Go Report Card](https://goreportcard.com/badge/github.com/StackVista/stackstate-agent)](https://goreportcard.com/report/github.com/StackVista/stackstate-agent)

The present repository contains the source code of the StackState Agent version 2.

## Documentation

The general documentation of the project is located under [the docs directory](docs) of the present repo.

## Getting started

To build the Agent you need:
 * [Go](https://golang.org/doc/install) 1.11.5 or later.
 * Python 2.7 along with development libraries.
 * Python dependencies. You may install these with `pip install -r requirements.txt`
   This will also pull in [Invoke](http://www.pyinvoke.org) if not yet installed.

**Note:** you may want to use a python virtual environment to avoid polluting your
      system-wide python environment with the agent build/dev dependencies. By default, this environment is only used for dev dependencies listed in `requirements.txt`, if you want the agent to use the virtual environment's interpreter and libraries instead of the system python's ones,
      add `--use-venv` to the build command.

**Note:** You may have previously installed `invoke` via brew on MacOS, or `pip` in
      any other platform. We recommend you use the version pinned in the requirements
      file for a smooth development/build experience.

Builds and tests are orchestrated with `invoke`, type `invoke --list` on a shell
to see the available tasks.

To start working on the Agent, you can build the `master` branch:

1. checkout the repo: `git clone https://github.com/StackVista/stackstate-agent.git $GOPATH/src/github.com/StackVista/stackstate-agent`.
2. cd into the project folder: `cd $GOPATH/src/github.com/StackVista/stackstate-agent`.
3. install project's dependencies: `invoke deps`.
   Make sure that `$GOPATH/bin` is in your `$PATH` otherwise this step might fail.
4. build the whole project with `invoke agent.build --build-exclude=snmp,systemd` (with `--use-venv` to use a python virtualenv)


## Run

To start the agent type `agent run` from the `bin/agent` folder, it will take
care of adjusting paths and run the binary in foreground.

You need to provide a valid API key. You can either use the config file or
overwrite it with the environment variable like:
```
STS_API_KEY=12345678990 ./bin/agent/agent -c bin/agent/dist/stackstate.yaml
```

## Install

Installation instructions are available on the [StackState docs site](https://docs.stackstate.com/stackpacks/integrations/agent).

##### Omnibus notes for windows build process

We ended up checking in a patched gem file under omnibus/vendor/cache/libyajl2-1.2.1.gem, to make windows builds work with newer msys toolchain.
The source of this can be found here https://github.com/StackVista/libyajl2-gem/tree/1.2.0-fixed-lssp. Ideally we'd be able to drop this hack once we
bump the ruby version > 2.6.5 because libyajl2 compiles proper on those ruby versions.

## GitLab cluster agent pipeline

If you want to speed up the GitLab pipeline and run only the steps related to the cluster agent, include the string `[cluster-agent]` in your commit message.

