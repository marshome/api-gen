#!/usr/bin/env bash

export OUT=${GOPATH}/bin/marsapi
go build -o ${OUT} ./
echo ${OUT} installed