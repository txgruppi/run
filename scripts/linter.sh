#!/bin/bash

set -e

find . -path ./vendor -prune -o -name "*.go" | xargs -n 1 dirname | sort -n | uniq | xargs gometalinter --deadline=30s
