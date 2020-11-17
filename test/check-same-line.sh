#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "missing file for comparsion"
    exit 1
fi

if [ $(wc -l < output-test.txt) -eq $(wc -l < test/input-test.txt) ]
then
    exit 0
else
    exit 1
fi

