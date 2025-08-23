#!/bin/bash

set -e

# api_url="http://localhost:8081"
api_url="https://bulbul.hub.hackload.kz"

docker run --network host --rm -i -u $(id -u) \
    -e "API_URL=${api_url}" \
    -e BASIC_AUTH=YXlzdWx0YW5fdGFsZ2F0XzFAZmVzdC50aXg6LzhlQyRBRD4= \
    -v $PWD:/app \
    -w /app \
    grafana/k6 run --http-debug - <booking.js