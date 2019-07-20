#!/bin/bash

echo "===== RUNNING =====" && \
  docker run --rm -it \
    --name golang-dev \
    --env-file=.env \
    -v "$PWD":/code \
    golang \
    bash
