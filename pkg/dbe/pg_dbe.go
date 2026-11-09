/*
Copyright © 2021 Colin Fox <greenenergy@gmail.com>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package dbe

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
)

// Postgres error codes (SQLSTATE) that checkInstall handles specially.
const (
	pgUndefinedTable   = "42P01"
	pgCannotConnectNow = "57P03"
)

type PGDBE struct {
	conn    *sqlx.DB
	verbose bool
	debug   bool
	retries int
}

type PGArgs struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// func NewPGDBE(host string, port int, user, password, dbname string, sslmode bool, verbose, debug bool, retries int) (DBEngine, error) {
func NewPGDBE(args *EngineArgs) (DBEngine, error) {

	connStr := args.ToConnStr(false)
	safeConnStr := args.ToConnStr(true)

	/*
		connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
			args.Host, args.Port, args.User, args.Name, args.SSLMode, args.Password)

		safeConnStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=<redacted>",
			args.Host, args.Port, args.User, args.Name, args.SSLMode)
	*/

	/*
		if args.SSLCert != "" {
			connStr +=
		}
	*/

	if args.Verbose {
		fmt.Println("connstr:", safeConnStr)
	}

	conn, err := sqlx.Open("postgres", connStr)
	if err != nil {
		fmt.Println("error opening database:", err.Error())
		return nil, fmt.Errorf("couldn't open postgres: %v", err.Error())
	}

	// Once we have a connection, we should check for our table & data.
	// If it doesn't exist, then this is the first time we're running
	// here. If it does exist, then we should read those IDs in as well
	// and seed the memory map of IDs. The question here is -- do we expect
	// all the patch files to be present with a given database? If we do
	// expect that, then we wouldn't need to load in the existing patches.

	pgdbe := PGDBE{
		conn:    conn,
		verbose: args.Verbose,
		debug:   args.Debug,
		retries: args.Retries,
	}

	err = pgdbe.checkInstall()
	if err != nil {
		return nil, err
	}

	return &pgdbe, nil
}

func (p *PGDBE) checkInstall() error {
	p.DPrint("checkInstall()...")

	for attempt := 1; ; attempt++ {
		p.DPrint("querying dbp_patch_table...")

		_, err := p.conn.Exec("select count(*) from dbp_patch_table")
		switch {
		case err == nil:
			return nil
		case isUndefinedTable(err):
			p.DPrint("creating dbp_patch_table...")

			_, err = p.conn.Exec(`
create table dbp_patch_table (
	id text primary key,
	created timestamp with time zone not null default CURRENT_TIMESTAMP,
	prereqs text,
	description text
)
`)
			if err != nil {
				return fmt.Errorf("problem creating patch table: %w", err)
			}
			return nil
		case !isRetryableConnError(err):
			return fmt.Errorf("problem checking for patch table: %w", err)
		case attempt >= p.retries:
			return fmt.Errorf("couldn't reach the database after %d attempts: %w", attempt, err)
		}

		fmt.Printf("database not ready (%q), retrying %d more times\n", err.Error(), p.retries-attempt)
		time.Sleep(time.Second)
	}
}

func isUndefinedTable(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pgUndefinedTable
}

// isRetryableConnError reports whether err means the server can't be reached
// yet: a network failure, or a server that is still starting up.
func isRetryableConnError(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == pgCannotConnectNow
}

func (p *PGDBE) GetInstalledIDs() (*set.Set, error) {
	q, err := p.conn.Query("select id from dbp_patch_table")
	if err != nil {
		return nil, fmt.Errorf("problem getting installed ids: %s", err.Error())
	}
	defer q.Close()

	output := set.Set{}

	for q.Next() {
		var id string
		err = q.Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("problem scanning id: %s", err.Error())
		}
		output.Add(id)
	}
	if err := q.Err(); err != nil {
		return nil, fmt.Errorf("problem reading installed ids: %s", err.Error())
	}
	return &output, nil
}

func (p *PGDBE) DPrint(msg string) {
	if p.debug {
		fmt.Println(msg)
	}
}

func (p *PGDBE) Patch(ptch *patch.Patch) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return fmt.Errorf("problem while trying to apply patch: %s", err.Error())
	}

	if ptch.HasOption("chop") {
		fmt.Println("chop option detected")
		for _, stmtText := range splitStatements(ptch.Body) {
			fmt.Println("Executing:", string(stmtText))
			_, err = tx.Exec(string(stmtText))
			if err != nil {
				if re := tx.Rollback(); re != nil {
					fmt.Println("*** Error rolling back:", err)
				}
				switch e := err.(type) {
				case *pq.Error:
					if p.debug {
						fmt.Println("Problem patch:", string(ptch.Body))
						fmt.Println("*** Problem query:", string(stmtText))
					}
					return fmt.Errorf("problem applying patch %s (%s) [detail: %q]: %s", ptch.Id, ptch.Filename, e.Detail, err.Error())
				default:
					return fmt.Errorf("problem applying patch %s (%s): %s", ptch.Id, ptch.Filename, err.Error())
				}
			}
		}
	} else {
		fmt.Println("chop option NOT detected")
		_, err = tx.Exec(string(ptch.Body))
		if err != nil {
			if rerr := tx.Rollback(); rerr != nil {
				fmt.Println("*** Error rolling back:", err)
			}
			switch e := err.(type) {
			case *pq.Error:
				if p.debug {
					fmt.Println("Problem patch:", string(ptch.Body))
				}
				return fmt.Errorf("problem applying patch %s (%s) [detail: %q]: %s", ptch.Id, ptch.Filename, e.Detail, err.Error())
			default:
				return fmt.Errorf("problem applying patch %s (%s): %s", ptch.Id, ptch.Filename, err.Error())
			}
		}
	}

	prereqs := strings.Join(ptch.Prereqs, ",")
	_, err = tx.Exec("insert into dbp_patch_table(id, prereqs, description) values ($1, $2, $3)",
		ptch.Id, prereqs, ptch.Description)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("problem applying patch file %q: %s", ptch.Filename, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("problem committing patch %s (%s): %s", ptch.Id, ptch.Filename, err.Error())
	}
	return nil
}
