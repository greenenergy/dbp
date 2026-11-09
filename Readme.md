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
        -- options: chop

The id can be any string without spaces.
The prereqs line is a comma separated list of patches that need to be applied before this one. The prereqs refer to the id field of the other patches, so it makes sense to use filenames for the ids. I usually use the filename without the .sql extension, ie:
The options line is for adding options. Current option is "chop", which chops the sql file into individual statements for tracking down errors within the file when applying. A ';' inside a comment, quoted string, quoted identifier or dollar-quoted string doesn't split, and the last statement doesn't need a ';'.

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

## Docker / Container Usage

The recommended deployment pattern for dbp is as a **purpose-built container** that carries both the `dbp` binary and the patch files for a specific project. The container runs as an init step before your application server starts, applies any outstanding patches, and exits cleanly.

### Dockerfile (in your project repo)

```dockerfile
FROM golang:1.25 AS builder
RUN go install github.com/greenenergy/dbp@latest

FROM alpine:latest
COPY --from=builder /go/bin/dbp /usr/local/bin/dbp

# Bake your project's patch files into the image
COPY patches /patches

ENTRYPOINT ["dbp", "apply", "-e", "postgres", "-f", "/patches"]
```

### docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: myapp
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U myapp"]
      interval: 5s
      timeout: 5s
      retries: 10

  dbp:
    build:
      context: .
      dockerfile: Dockerfile.dbp
    environment:
      DB_HOST: postgres
      DB_NAME: myapp
      DB_USER: myapp
      DB_PASSWORD: ${DB_PASSWORD}
      DB_SSLMODE: disable
    depends_on:
      postgres:
        condition: service_healthy

  server:
    image: myapp:latest
    depends_on:
      dbp:
        condition: service_completed_successfully

volumes:
  postgres_data:
```

The `service_completed_successfully` condition tells Compose to wait until the dbp container exits with code 0 before starting the application server. The retries flag (`-r`, default 10) handles the window between postgres becoming healthy and accepting connections.

### Environment variables

Connection parameters can be set via environment variables instead of CLI flags:

| Environment variable | CLI flag        | Default   |
|----------------------|-----------------|-----------|
| `DB_HOST`            | `--db.host`     | (none)    |
| `DB_USER`            | `--db.username` | (none)    |
| `DB_PASSWORD`        | `--db.password` | (none)    |
| `DB_NAME`            | `--db.name`     | (none)    |
| `DB_SSLMODE`         | `--db.sslmode`  | `require` |
| `DB_SSLCERT`         | `--db.sslcert`  | (none)    |
| `DB_SSLKEY`          | `--db.sslkey`   | (none)    |
| `DB_SSLROOTCERT`     | `--db.sslrootcert` | (none) |

Explicit CLI flags always override environment variables. Using environment variables is preferred in container environments — it keeps secrets out of process arguments and `docker inspect` output.

---

## Kubernetes / Helm Usage

The same purpose-built dbp+patches container from the Docker section above works as a Helm hook Job. Wire it up as a `pre-install,pre-upgrade` hook with a low (very negative) `hook-weight` so it's the *first* thing Helm runs — before any other resource in the chart is created. If the job fails, Helm aborts the release right there: no Deployment, no Service, nothing else gets created against a database that isn't in the state your app expects.

```yaml
# templates/job-dbp-migrate.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-dbp-migrate
  annotations:
    helm.sh/hook: pre-install,pre-upgrade
    helm.sh/hook-weight: "-10"
    helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded
spec:
  backoffLimit: 3
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: dbp-migrate
          image: "{{ .Values.dbpImage.registry }}/{{ .Values.dbpImage.repository }}:{{ .Values.dbpImage.tag }}"
          env:
            - name: DB_HOST
              value: {{ .Values.db.host | quote }}
            - name: DB_PORT
              value: {{ .Values.db.port | quote }}
            - name: DB_NAME
              value: {{ .Values.db.name | quote }}
            - name: DB_USER
              value: {{ .Values.db.migrateUser | quote }}
            - name: DB_SSLMODE
              value: {{ .Values.db.sslmode | quote }}
            - name: DB_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.db.secretName }}
                  key: {{ .Values.db.secretKey }}
```

A few things that matter here:

- **`helm.sh/hook-delete-policy: before-hook-creation,hook-succeeded`** cleans up the completed Job before the next `helm upgrade` creates a new one — Jobs are immutable, so without this a second install/upgrade fails trying to recreate one with the same name.
- **`backoffLimit`** should stay low (2-3). If `dbp` fails, that's a real problem to surface immediately, not something to silently retry many times.
- **The database and role must already exist** before the first install — see "Least-Privilege Setup" below. dbp does not create databases or roles itself; it only applies patches to whatever it's pointed at, by design.
- **`DB_USER` here should be the DDL-privileged role**, not whatever role your application itself connects as — see below for why that split matters and how to set it up.

---

## Least-Privilege Setup (separate DDL and app roles)

dbp deliberately does not create databases or roles for you — you create the database and grant access yourself, so dbp can never accidentally touch a system it wasn't meant to. That also means it's up to you to decide how much power the connecting role has.

The recommended pattern is **two roles**: one with DDL rights that only `dbp` ever connects as, and one with plain CRUD rights (`SELECT`/`INSERT`/`UPDATE`/`DELETE`, no schema changes) that your application connects as. That way, even if your application has a bug or gets compromised, it can't alter or drop anything.

The gotcha people hit here: `GRANT ... ON ALL TABLES IN SCHEMA public` only affects tables that *already exist* at the moment you run it. Every table a *future* patch creates is left ungranted, and you end up hand-granting again after every migration. The fix is `ALTER DEFAULT PRIVILEGES`, which tells Postgres "whenever this specific role creates a table here, automatically grant these rights to that other role" — so every future patch that creates a table needs no follow-up.

Example, using a project called `timetracker` with a migration role and an app role:

```sql
-- Run once, as the postgres superuser

-- 1. Create the two roles
CREATE ROLE timetracker_migrate LOGIN PASSWORD '<ddl-password>';
CREATE ROLE timetracker_app LOGIN PASSWORD '<app-password>';

-- 2. Create the database, owned by the DDL role
CREATE DATABASE timetracker OWNER timetracker_migrate;

-- 3. Switch into it for the rest
\c timetracker

-- 4. Let the app role connect and see the schema
GRANT CONNECT ON DATABASE timetracker TO timetracker_app;
GRANT USAGE ON SCHEMA public TO timetracker_app;

-- 5. Grant CRUD on whatever exists right now (harmless on a fresh
--    empty database; re-run this once if you ever bulk-load
--    something outside of dbp)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO timetracker_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO timetracker_app;

-- 6. The actual fix: auto-grant CRUD on every table/sequence
--    timetracker_migrate creates from now on, i.e. every dbp patch
ALTER DEFAULT PRIVILEGES FOR ROLE timetracker_migrate IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO timetracker_app;

ALTER DEFAULT PRIVILEGES FOR ROLE timetracker_migrate IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO timetracker_app;
```

Then point `dbp` at `timetracker_migrate` (`DB_USER=timetracker_migrate`) and your application at `timetracker_app`.

Two things worth knowing:
- **Sequences matter.** Any `SERIAL`/`BIGSERIAL`/identity column backs onto a sequence, and `INSERT` needs `USAGE` on it to call `nextval()`. Easy to forget, and the resulting error ("permission denied for sequence") doesn't make the cause obvious.
- **The default-privilege grant is tied to the creator role.** It only fires for objects `timetracker_migrate` itself creates. If you ever run ad-hoc DDL as the superuser instead of through dbp, those objects won't get the automatic grant — so `dbp` should always be the one applying schema changes, consistently as the same role.

This also covers views, not just tables — Postgres's `ON TABLES` default-privilege rule applies to views as well, since they live in the same catalog.

---

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


