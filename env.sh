#!/bin/bash
ROOT_PATH=$(dirname $(dirname $(cd $(dirname ${BASH_SOURCE:-$0});pwd)))
###Compile###
export GOPATH=$ROOT_PATH
echo $GOPATH
###Runtime###
