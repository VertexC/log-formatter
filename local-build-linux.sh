#!/bin/bash

if [ $# -eq 0 ]
  then
    echo "No config yaml file sepecified"
    exit 1
fi


env GOOS=linux go build -v main.go

docker build . --tag log-formatter/latest --build-arg CONFIG=$1
