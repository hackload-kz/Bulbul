#!/bin/bash

set -e
set -x

docker run --rm -i -u $(id -u) \
    -e API_URL=https://bulbul.hub.hackload.kz \
    -e BASIC_AUTH=YXlzdWx0YW5fdGFsZ2F0XzFAZmVzdC50aXg6LzhlQyRBRD4= \
    -v $PWD:/app \
    -w /app \
    grafana/k6 run - <load.js