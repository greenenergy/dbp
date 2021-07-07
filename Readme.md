# Readme

The intention here is to create a system where you can specify update migration scripts and the system will make sure they get applied properly.

One thing I want for sure is the ability to specify prerequisites. You must be able to say that other files need to be applied before you, which will allow the migration system to apply the patches in the order they need to be applied.

---


## Signing
I like the idea of having each update be actually a gzipped package of two files - the sql script to be executed, and a digital fingerprint of the file signed by the author. 

To make it easy (create signature & zip together), I could build that into my migration tool. You would use the tool to create a migration package, and then you could also use the tool to apply the migration packages found in a directory.

	./migrate new package <sqlfile>

This will create a <sqlfile>.dbp (database package), which is a gzip of two files: The original sql file, and then a digital signature from this user.

	./migrate new package -k <which key> <sqlfile>

If the user has multiple keys, this allows them to specify which one should be used to sign the package.

## ON THE OTHER HAND

I'm not sure what actual value this adds. If you need someone's key to validate that they signed a particular patch, and that person is no longer with the company (or their key is no longer valid), then the signature would be undecypherable anyway.

And it seems like a bit much to require people to install gpg and create a private keyring so they could apply database patches. The two aren't really associated.

So I'd need to support un-signed patches. And if I am to accept unsigned patches, a bad guy who wishes to tamper could simply unzip the file, pull out the sql and create a new non-signed version with whatever alterations they wanted to make.

I could protect against that by forcing you to encrypt packages intended for a given server, and that server could decode those encrypted patches, but this seems like an inordinate amount of security at a level where you would basically have to be the sysadmin anyway, so what added security is it actually offering? And by being encrypted, it's now a mystery and cannot be reviewed by anyone (sercurity pros OR other coders).

So in the end, I think the best approach is plain text files. I don't want to use the filename for anything important though - too much potential collision between what I would like to use it as/for/with, and what the OS may want to do. Or even how a given sysadmin might want to layout their folders.

I don't need to use the filenames for anything other than a handle to load the data. I can put all the metadata into the first bit of the file, as specifically formatted comments. This way you can use any tool to load and edit the sql, and as long as you don't touch the header comments, it will be compatible with the migration system.


## Operation

Each patch file will have a few specific and important fields in the top portion of the sql patch file. The very first thing that should be in the file, right at the top, is:

  -- PATCH vX.Y.Z
 
 If the first few bytes of the file are not "-- PATCH v" then this is not a patch file and to not even try to parse it.

 The next keys to look for are:
 * id  (unique string for the file -- uuids work great, but you can use any string)
 * author (email of author for the file - with the ID is a unique key)
 * prereq (comma separated list of IDs of patches that must have already been applied)
 * description (plain text description, which may be returned by list and info commands)
 * tags (list of words that 'tag' the file. Can be used in search functions)
 * created (creation timestamp)

 This is enough for now. The magic is in the prereq line. This allows patches to reference each other by ID. This way the filenames don't matter, so you don't have to worry about someone else already using your filename.



