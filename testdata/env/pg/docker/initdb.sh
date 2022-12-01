#!/bin/bash -e

PGPASSWORD=$POSTGRES_PASSWORD createuser -U postgres -s pgpatcher
PGPASSWORD=$POSTGRES_PASSWORD createdb -U postgres -O pgpatcher patchtest

echo "Done setup"

