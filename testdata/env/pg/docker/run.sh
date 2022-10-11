#!/bin/sh
set -x
set -e

until PGPASSWORD=postgres psql -h db -U postgres -c '\q'; do
  >&2 echo "Postgres is unavailable - sleeping"
  sleep 1
done

#sleep 10
#until su -c "psql -h postgres -U 'postgres' -c '\q'" postgres ;  do
#  >&2 echo "Postgres is unavailable - sleeping"
#  sleep 1
#done

# These will fail on subsequent runs, but these errors can be ignored
PGPASSWORD=postgres createuser -h db -U postgres -s pgpatcher
PGPASSWORD=postgres createdb -h db -U postgres -O pgpatcher patchtest

#sudo -u postgres createuser -h postgres -U postgres -s pgpatcher
#sudo -u postgres createdb -h postgres -U postgres -O pgpatcher patchtest

