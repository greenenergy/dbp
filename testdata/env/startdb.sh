#!/bin/sh

mkdir -p testdata/pgdata
CURRENT_UID=$(id -u):$(id -g) docker-compose up

