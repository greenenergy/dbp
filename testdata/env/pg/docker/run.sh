#!/bin/sh

#sleep 10
until psql -h postgres -U "postgres" -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

# These will fail on subsequent runs, but these errors can be ignored
createuser -h postgres -U postgres -s pgpatcher
createdb -h postgres -U postgres -O pgpatcher patchtest

