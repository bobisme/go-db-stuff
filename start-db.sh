#!/bin/sh

docker run --rm -d --name pg -p 5432:5432 postgres:alpine
