#!/bin/sh

docker run --rm -v $PWD/api:/spec redocly/cli bundle index.yaml > ./static/openapi.yaml
docker run --rm -v $PWD/api:/spec redocly/cli bundle index.yaml --ext=json > ./static/openapi.json
