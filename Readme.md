# dbp - database patch

This program allows you to define and maintain an SQL database through the use of patch files. These patch files can refer to each other (by ID) in order to define precedence.  This allows you to easily make sure that the patches are executed in the required order.  It will also make sure that there are no loops or breaks in the branch list (list of prerequisites).

Filenames and directories are ignored by the patcher, or more specifically they don't factor into the processing of the scripts. Only the IDs and the prereq relationships between them. So you are free to name the scripts any way you wish, and have any folder hierarchy you wish.

dbp uses ids to establish precedence, and these IDs can be any string without spaces or commas. UUIDs are a good choice, and the `./dbp new` command will generate a new UUID for you. However, you can use any string you wish, as long as it is unique and has no spaces or commas.

This patching system is designed for "forward" patching only - as in, there is no rollback functionality. Rolling back a patch for code is trivial, but there are database operations you can perform in your patch files that cannot be "rolled back" - such as deleting rows from tables. In order to be able to perform a true "rollback", you will need to snapshot your database ahead of the patch, and then you can roll back to the snapshot if necessary.

You can launch with "--dry" which will set dbp into `dry run` mode, where it will tell you what work it will do without actually doing anything. This also allows you to check for potential problems with IDs and prereqs.

Each patch file is executed in its own transaction, and the id is appended to a dbp private table in the same transaction. If there is a problem, then the transaction is rolled back and neither the patch nor the update remains. 
## Usage
At the top of each patch file, you need to add two comment lines:
    
        -- id: <id>
        -- prereqs: <id>,<id>,<id>...

The id can be any string without spaces.
The prereqs line is a comma separated list of patches that need to be applied before this one. The prereqs refer to the id field of the other patches, so it makes sense to use filenames for the ids. I usually use the filename without the .sql extension, ie:

File1, bugfix_1234.sql:
        
            -- id: bugfix_1234
            -- prereqs:
            ...sql here...

File2, bugfix_2468.sql:
        
            -- id: bugfix_2468
            -- prereqs: bugfix_1234 
            ...sql here...

dbp will make sure that File2 is applied after File1.

## Testing

You can demonstrate dbp's detection of various problems via patch hierarchies in the testdata folder. Each folder is named for the kind of error it illustrates, as in :

    ./dbp apply -f testdata/long_loop
    ./dbp apply -f testdata/short_loop
    ./dbp apply -f testdata/shortest_loop
    ./dbp apply -f testdata/dupe_id
    ./dbp apply -f testdata/missing_id_1
    ./dbp apply -f testdata/missing_id_2

## Why This Patcher?

The current tool we're using ("Migrate") has inadequate design features, IMHO:
* It does not use a transaction to wrap both the patch being applied and the update to the migration management table that records the patch was applied. In fact it apparently doesn't use a built in transaction at all for the patch, so it's up to the patch author to wrap their patch in BEGIN/END. If they don't, presumably you can end up with partial changes to the database if something goes wrong.
* It demands that authors write "up" and "down" (apply/rollback) versions of their patches - even though you can't always roll back. If a patch deleted records, or dropped a table, how would you "down" that? You can't - you'd have to roll the database back to a prior snapshot. The only safe way to undo a patch is to use snapshotting. Creating a "down" file provides an artificial sense of security.
* It uses filenames to order operations - which can be complicated if you have two (or more) developers working on patches in separate project branches. The standard method appears to be to prefix your filename with a number, indicating what order it will get applied in. (Is the number even parsed AS a number by the ordering system? If they just use alphabetic sorting, then 1_ and 11_ will sort back to back, and they shouldn't). What happens if two people come up with the same filename? Or at least the same prefix number - which is supposed to come first?
* It forces all patches to be in one directory, so you can't, for example, have a folder of bug fixes and a folder of features.
* Since a patch may have its UP and/or DOWN scripts applied, how do you know what state a given database is supposed to be in? If you applied EVERY patch in the system, you would have an empty database.

DBP, on the other hand, does things differently:
* Each patch, and the update to the management table happens within a transaction. Either the whole patch plus the mgment update happens, or neither happens.
* Only "forward" patches are required - you build the patches that are supposed to bring the database to the "current" state. If you need to undo a previous patch, you write another patch that has the previous one as a prereq, and you write the reversal. You only do this if you actually need to undo the previous patch to have the db in a correct state.
* Filenames don't matter. A comment header in each patch file contains an ID (which can be any string without commas in it, though the recommended ID would be the current filename minus the extension) and a prereq line listing all the prerequisite patches that should be applied before this patch is applied. This allows complex and sophisticated ordering to be managed, if need be. The prereq field can be ignored if order doesn't matter. One file with the filename "init_patch.sql" will be considered the starting point, everything else is ID based.
* A directory is passed as the starting point, and the whole tree is walked and the map of files and relationships is built from there.
* All patches are applied when the patcher is run, and the system is idempotent, so re-patching an existing system will have no effect unless there are new patches involved.
* You can exclude folders from the current patch run, meaning you can have demo data that gets installed during a development setup and ignored during a production update.


