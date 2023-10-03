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

/*
Note:
This file was originally written mainly with Queryx() calls, but for some
reason the create & insert functions didn't actually perform any permanent
change to the database. After running, it was as though nothing had been
done, but no errors were reported. After changing from Query() to Exec(),
everything worked. I don't know why this is the case.
*/

import (
	"fmt"
	"strings"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDBE struct {
	conn    *sqlx.DB
	verbose bool
}

type SQLITEArgs struct {
	Filename string `json:"filename"`
}

func NewSQLiteDBE(args *EngineArgs) (DBEngine, error) {
	/*
		var sqliteargs SQLITEArgs
		data, err := ioutil.ReadFile(credsName)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &sqliteargs)
		if err != nil {
			return nil, err
		}
	*/

	filename := "sqlite_dummy.db"

	conn, err := sqlx.Open("sqlite3", filename)
	if err != nil {
		return nil, fmt.Errorf("problem opening sqlite db: %s", err.Error())
	}

	dbe := &SQLiteDBE{
		conn:    conn,
		verbose: true,
	}

	err = dbe.checkInstall()
	if err != nil {
		return nil, err
	}
	return dbe, nil
}

func (p *SQLiteDBE) GetInstalledIDs() (*set.Set, error) {
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

func (p *SQLiteDBE) checkInstall() error {
	_, err := p.conn.Query("select count(*) from dbp_patch_table")
	if err != nil {
		_, err := p.conn.Exec(`
create table dbp_patch_table (
	id text primary key,
	created timestamp with time zone not null default (datetime('now','localtime')),
	prereqs text,
	description text
);
`)
		if err != nil {
			fmt.Println("Error creating patch table:", err.Error())
			return fmt.Errorf("problem creating patch table:%q", err)
		}

	}
	return nil
}

func (p *SQLiteDBE) Patch(ptch *patch.Patch) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return fmt.Errorf("problem while trying to apply patch: %s", err.Error())
	}

	_, err = tx.Exec(string(ptch.Body))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("problem applying patch %s (%s): %s", ptch.Id, ptch.Filename, err.Error())
	}

	prereqs := strings.Join(ptch.Prereqs, ",")
	_, err = tx.Exec("insert into dbp_patch_table(id, prereqs, description) values ($1, $2, $3)",
		ptch.Id, prereqs, ptch.Description)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("problem updating patch record %s: %s", ptch.Id, err.Error())
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println("error committing:", err.Error())
	}
	return nil
}
