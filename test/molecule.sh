#!/usr/bin/env bash

set -ex

VENV_PATH=./p-env

if [[ -z $CI_COMMIT_REF_NAME ]]; then
  export AGENT_GITLAB_BRANCH=`git rev-parse --abbrev-ref HEAD`
else
  export AGENT_GITLAB_BRANCH=$CI_COMMIT_REF_NAME
fi

if [[ ! -d $VENV_PATH ]]; then
  virtualenv  $VENV_PATH
  source $VENV_PATH/bin/activate
  pip install -r molecule-role/requirements.txt
else
  source $VENV_PATH/bin/activate
fi

cd molecule-role

molecule "$@"
