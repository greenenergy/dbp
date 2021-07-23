#!/bin/sh

mkdir -p data/pgdata
CURRENT_UID=$(id -u):$(id -g) docker-compose up $1

