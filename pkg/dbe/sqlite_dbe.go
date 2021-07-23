package dbe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

func NewSQLiteDBE(credsName string, verbose bool) (DBEngine, error) {
	var sqliteargs SQLITEArgs
	data, err := ioutil.ReadFile(credsName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &sqliteargs)
	if err != nil {
		return nil, err
	}

	if verbose {
		fmt.Println("about to open:", sqliteargs.Filename)
	}
	conn, err := sqlx.Open("sqlite3", sqliteargs.Filename)
	if err != nil {
		return nil, fmt.Errorf("problem opening sqlite db: %s", err.Error())
	}

	dbe := &SQLiteDBE{
		conn:    conn,
		verbose: verbose,
	}

	err = dbe.CheckInstall()
	if err != nil {
		return nil, err
	}
	return dbe, nil
}

func (p *SQLiteDBE) IsConfigured() bool {
	return false
}

func (p *SQLiteDBE) Configure() error {
	return errors.New("sqlite engine: Configure() unimplemented")
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

func (p *SQLiteDBE) CheckInstall() error {
	_, err := p.conn.Query("select count(*) from dbp_patch_table")
	if err != nil {
		fmt.Println("Error querying, now going to try to create dbp_patch_table")
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
