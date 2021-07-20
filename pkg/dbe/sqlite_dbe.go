package dbe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
	"time"

	"github.com/greenenergy/dbp/pkg/patch"
	"github.com/greenenergy/dbp/pkg/set"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDBE struct {
	conn    *sqlx.DB
	already set.Set
}

type SQLITEArgs struct {
	Filename string `json:"filename"`
}

func NewSQLiteDBE(credsName string) (DBEngine, error) {
	var sqliteargs SQLITEArgs
	data, err := ioutil.ReadFile(credsName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &sqliteargs)
	if err != nil {
		return nil, err
	}

	conn, err := sqlx.Open("sqlite3", sqliteargs.Filename)
	if err != nil {
		return nil, fmt.Errorf("problem opening sqlite db:", err.Error())
	}

	return &SQLiteDBE{
		conn: conn,
	}, nil
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
	success := false
	retries := 10

	for x := 0; x < retries; x++ {
		_, err := p.conn.Queryx("select count(*) from dbp_patch_table")
		if err != nil {
			if _, ok := err.(*net.OpError); ok {
				fmt.Println("this is a network error, gonna retry:", err.Error())
				time.Sleep(time.Second)
			} else {
				//fmt.Printf("Got an error trying to see if I am alread installed")
				//log.Fatal(err.Error())
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

func (p *SQLiteDBE) Patch(ptch *patch.Patch) error {

	tx, err := p.conn.Begin()
	if err != nil {
		return fmt.Errorf("problem while trying to apply patch: %s", err.Error())
	}

	_, err = tx.Query(string(ptch.Body))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("problem applying patch %s (%s): %s", ptch.Id, ptch.Filename, err.Error())
	}

	prereqs := strings.Join(ptch.Prereqs, ",")
	_, err = tx.Query("insert into dbp_patch_table(id, prereqs, description) values ($1, $2, $3)",
		ptch.Id, prereqs, ptch.Description)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("problem updating patch record %s: %s", ptch.Id, err.Error())
	}

	tx.Commit()
	return nil
}
