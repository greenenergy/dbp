The startdb.sh script creates a folder using your current user
and then launches postgres using that same user. This makes cleanup
easier, since the default is to run postgres as root and you end up
with root-owned files & directories for later cleanup.



