#!/bin/bash

set -e

find . -path ./vendor -prune -o -name "*_test.go" | xargs -n 1 dirname | sort -n | uniq | grep -v -E "^\.$" | grep -v -E '^\.\/\.' | xargs go test -cover "$@"
