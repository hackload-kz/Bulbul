#!/bin/bash

set -e
set -x

docker run --network host --rm -i -u $(id -u) \
    -e API_URL=http://localhost:8081 \
    -e BASIC_AUTH=YXlzdWx0YW5fdGFsZ2F0XzFAZmVzdC50aXg6LzhlQyRBRD4= \
    -v $PWD:/app \
    -w /app \
    grafana/k6 run --http-debug - <auth-check.js