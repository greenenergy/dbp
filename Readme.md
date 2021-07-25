# dbp - database patch

This program allows you to define and maintain an SQL database through the use of patch files. These patch files can refer to each other (by ID) in order to define precedence.  This allows you to easily make sure that the patches are executed in the required order.  It will also make sure that there are no loops or breaks in the branch list (list of prerequisites).

Filenames and directories are ignored by the patcher, or more specifically they don't factor into the processing of the scripts. Only the IDs and the prereq relationships between them. So you are free to name the scripts any way you wish, and have any folder hierarchy you wish.

dbp uses ids to establish precedence, and these IDs can be any string without spaces or commas. UUIDs are a good choice, and the `./dbp new` command will generate a new UUID for you. However, you can use any string you wish, as long as it is unique and has no spaces or commas.

This patching system is designed for "forward" patching only - as in, there is no rollback functionality. Rolling back a patch for code is trivial, but there are database operations you can perform in your patch files that cannot be "rolled back" - such as deleting rows from tables. In order to be able to perform a true "rollback", you will need to snapshot your database ahead of the patch, and then you can roll back to the snapshot if necessary.

You can launch with "--dry" which will set dbp into `dry run` mode, where it will tell you what work it will do without actually doing anything. This also allows you to check for potential problems with IDs and prereqs.

Each patch file is executed in its own transaction, and the id is appended to a dbp private table in the same transaction. If there is a problem, then the transaction is rolled back and neither the patch nor the update remains. 

