#!/usr/bin/env bash

# Git root path
ROOT_PATH=$(git rev-parse --show-toplevel)

export PATH=$ROOT_PATH/.tools/go/bin:$PATH
export GOPATH=$(go env GOPATH)
export PATH=$GOPATH/bin:$PATH