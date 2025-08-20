#!/usr/bin/env bash

set -e

# Decrypt and use the .env file
env_file="./secrets/environment"
ansible-vault decrypt $env_file
export $(cat $env_file | xargs)
ansible-vault encrypt $env_file