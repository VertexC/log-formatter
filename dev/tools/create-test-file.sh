#!/bin/bash

while true; do sleep 1 && echo $(date '+%Y-%m-%d %H:%M:%S') hello world >> test/input-test.txt; done
