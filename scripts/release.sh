#!/bin/bash

set -e

cd $(git rev-parse --show-toplevel)

if [ "$1" != "major" ] && [ "$1" != "minor" ] && [ "$1" != "patch" ] && [ "$1" != "keep" ]; then
  echo "You must specify which part of the version should be updated or specify keep to not change the version" >&2
  exit 1
fi

if [ "$1" != "keep" ] && [ ! -z "$(git status --porcelain)" ]; then
  echo "Working directory not clean. Can't update version" >&2
  exit 1
fi

if [ "$1" != "keep" ] && git rev-parse --abbrev-ref HEAD | grep -v -E '^master$' &>/dev/null; then
  echo "Not in master branch. Can't update version" >&2
  exit 1
fi

VERSION=$(cat VERSION | sed -E 's/^\s+|\s+$//g')

MAJOR=$(echo $VERSION | cut -d . -f 1)
MINOR=$(echo $VERSION | cut -d . -f 2)
PATCH=$(echo $VERSION | cut -d . -f 3)

if [ "$1" == "major" ]; then
  MAJOR=$(($MAJOR + 1))
  MINOR=0
  PATCH=0
fi

if [ "$1" == "minor" ]; then
  MINOR=$(($MINOR + 1))
  PATCH=0
fi

if [ "$1" == "patch" ]; then
  PATCH=$(($PATCH + 1))
fi

VERSION="$MAJOR.$MINOR.$PATCH"

rm -rf release

glide install

HASH=$(git rev-parse HEAD)
COMPILED=$(date "+%Y-%m-%dT%H:%M:%SZ%z")

gox \
  -osarch="linux/amd64 linux/386 darwin/amd64 darwin/386 windows/amd64 windows/386" \
  -output="release/{{.Dir}}_{{.OS}}_{{.Arch}}" \
  -ldflags "-w -s -X github.com/txgruppi/run/build.Version=$VERSION -X github.com/txgruppi/run/build.Commit=$HASH -X github.com/txgruppi/run/build.Compiled=$COMPILED"

which upx &>/dev/null && upx release/*

if [ "$1" != "keep" ]; then
  echo "$VERSION" > VERSION
  git add VERSION
  git commit -S -m "Released version $VERSION"
  git tag -s "$VERSION" -m "Version $VERSION"
  git push
  git push --tags
fi
