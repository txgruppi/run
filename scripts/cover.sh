#!/bin/bash

if ! which gocoverage &>/dev/null; then
	go get -v github.com/txgruppi/gocoverage
fi

gocoverage -covermode=atomic
