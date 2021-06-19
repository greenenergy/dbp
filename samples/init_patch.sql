-- PATCH: v0.0.1
-- id: fece2b8e-cf43-11eb-b7f3-07af1b70a47a
-- author: cfox@infoblox.com
-- description: Initial schema file. This is the one file that
--   must have this name in this folder. All other names are
--   irrelevant. Descriptions may be multi line if you make them
--   back to back, and two dashes followed by three spaces. The
--   description will be added to until the pattern breaks, and
--   a line is encountered that does not start with "--   ".
-- The description is the last thing read as metadata, so don't
-- put anything after it that should come before it.


-- The rest of the file is standard sql.

-- I could add a "reversable" feature, where a given sql file
-- could be reversed, but not all of them will be. For example,
-- if a table is getting dropped, that is an irreversable change.

-- The migration should also have a dry-run operation where it 
-- would print out which patch files it was going to execute,
-- and in what order. It could print them out by ID and description.
-- It could also print out hints that indicate why the order is
-- the way it is (based on prereq precedence)

-- I'm not sure if reversability is a desirable feature here, except
-- in the case of a developer working on his own database.

-- It would also be nice to remember the last used parameters
-- for database connection. This way if you were always working
-- with the same database, you would only need to specify the
-- connection parameters once, and it would remember them for
-- later.
-- Might have to store the db password in plaintext, since I don't
-- think postgres supports jwt login.

