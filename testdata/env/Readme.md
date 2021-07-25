# Readme

This folder holds test environments for both sqlite and postgres.

---
## Postgres:

The startdb.sh script is used to bring up a docker-compose session running postgresql. It creates a folder using your current user and then launches postgres using that same user. This makes cleanup easier, since the default is to run postgres as root and you end up with root-owned files & directories for later cleanup.

There is a Makefile in the env/pg folder that can fully set up the test evaluation environment for postgres. Run `make` by itself to see the options.

--
## SQLite

To run the sqlite test session, simply go to env/sqlite and run "./go.sh". This will launch dbp in sqlite mode, which will create the database and apply the files. You can then examine the result with:

	sqlite3 test.db

