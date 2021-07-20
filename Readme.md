# dbp - database patch

This program allows you to define and maintain an SQL database through the use of patch files. These patch files can refer to each other (by ID) in order to define precedence.  This allows you to easily make sure that the patches are executed in the required order.  It will also make sure that there are no loops or breaks in the branch list (list of prerequisites).

Filenames and directories are ignored by the patcher, or more specifically they don't factor into the processing of the scripts. Only the IDs and the prereq relationships between them. So you are free to name the scripts any way you wish, and have any folder hierarchy you wish.

You can run `dbp new` in order to create a new patch file. dbp will generate a new UUID, and you can just fill out the rest for your patch.


## Operation

Each patch file will have a few specific and important fields in the top portion of the sql patch file. The very first thing that should be in the file, right at the top, is:

  -- PATCH vX.Y.Z
 
 If the first few bytes of the file are not "-- PATCH v" then this is not a patch file and to not even try to parse it.

 The next keys to look for are:
 * id  (unique string for the file -- uuids work great, but you can use any string)
 * prereq (comma separated list of IDs of patches that must have already been applied)
 * description (plain text description, which may be returned by list and info commands)

 This is enough for now. The magic is in the prereq line. This allows patches to reference each other by ID. This way the filenames don't matter, so you don't have to worry about someone else already using your filename.



