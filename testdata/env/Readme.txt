The startdb.sh script creates a folder using your current user
and then launches postgres using that same user. This makes cleanup
easier, since the default is to run postgres as root and you end up
with root-owned files & directories for later cleanup.

Once you start the database engine (postgres), you need to create the
database. dbp doesn't currently create databases, it expects them to
be present and with the user & password already configured.

To prepare, once you have the database running (such as the docker-compose
test db we have), log into the running database instance and execute these queries:

createuser -U postgres -s pgpatcher
createdb -U postgres -O pgpatcher patchtest


Because the patcher needs to be able to execute anything that your user would
do in the sql files, it should be made a sysadmin - this is why the -s flag.

The above database and user are only necessary if you wish to tuse the test "dbcreds.js"
provided in the testdata here.


