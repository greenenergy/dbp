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
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // This  is the postgres driver for sqlx
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

func NewPGDBE(host string, port int, user, password, dbname string, sslmode bool, verbose, debug bool, retries int) (DBEngine, error) {

	/*
		var pgargs PGArgs
		data, err := ioutil.ReadFile(credsName)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &pgargs)
		if err != nil {
			return nil, err
		}
	*/

	mode := "disable"
	if sslmode {
		mode = "enable"
	}

	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
		host, port, user, dbname, mode, password)
	//	pgargs.Host, pgargs.Port, pgargs.Username, pgargs.Name, mode, pgargs.Password)

	if verbose {
		fmt.Println("connstr:", connStr)
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
		verbose: verbose,
		debug:   debug,
		retries: retries,
	}

	err = pgdbe.checkInstall()
	if err != nil {
		return nil, err
	}

	return &pgdbe, nil
}

func (p *PGDBE) checkInstall() error {
	success := false

	for x := 0; x < p.retries; x++ {
		_, err := p.conn.Queryx("select count(*) from dbp_patch_table")
		if err != nil {
			if _, ok := err.(*net.OpError); ok {
				fmt.Printf("this is a network error(%q), retrying %d more times\n", err.Error(), p.retries-x)
				time.Sleep(time.Second)
			} else {
				_, err = p.conn.Queryx(`
create table dbp_patch_table (
	id text primary key,
	created timestamp with time zone not null default CURRENT_TIMESTAMP,
	prereqs text,
	description text
)
`)
				if err != nil {
					return fmt.Errorf("problem creating patch table:%q", err)
				}

			}
		} else {
			success = true
		}
	}
	if !success {
		fmt.Println("couldn't find the db")
		log.Fatal("no database")
	}
	return nil
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
	return &output, nil
}

func (p *PGDBE) DPrint(msg string) {
	if p.debug {
		fmt.Println(msg)
	}
}

func (p *PGDBE) Patch(ptch *patch.Patch) error {
	p.DPrint("Patch -1-")
	defer p.DPrint("Patch -exiting-")

	tx, err := p.conn.Begin()
	if err != nil {
		return fmt.Errorf("problem while trying to apply patch: %s", err.Error())
	}
	p.DPrint("Patch -2-")

	_, err = tx.Query(string(ptch.Body))
	if err != nil {
		tx.Rollback()
		switch e := err.(type) {
		case *pq.Error:
			return fmt.Errorf("problem applying patch %s (%s) [detail: %q]: %s", ptch.Id, ptch.Filename, e.Detail, err.Error())
		default:
			return fmt.Errorf("problem applying patch %s (%s): %s", ptch.Id, ptch.Filename, err.Error())
		}

	}
	p.DPrint("Patch -3-")

	prereqs := strings.Join(ptch.Prereqs, ",")
	_, err = tx.Query("insert into dbp_patch_table(id, prereqs, description) values ($1, $2, $3)",
		ptch.Id, prereqs, ptch.Description)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("problem updating patch record %s: %s", ptch.Id, err.Error())
	}
	p.DPrint("Patch -4-")

	err = tx.Commit()
	if err != nil {
		fmt.Println("*** Problem committing:", err.Error())
	}
	return nil
}
