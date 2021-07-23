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
	_ "github.com/lib/pq" // This  is the postgres driver for sqlx
)

type PGDBE struct {
	conn    *sqlx.DB
	already map[string]bool
	verbose bool
}

type PGArgs struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewPGDBE(credsName string, verbose bool) (DBEngine, error) {

	var pgargs PGArgs
	data, err := ioutil.ReadFile(credsName)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &pgargs)
	if err != nil {
		return nil, err
	}

	mode := "disable"

	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s password=%s",
		pgargs.Host, pgargs.Port, pgargs.Username, pgargs.Name, mode, pgargs.Password)

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
	}

	err = pgdbe.CheckInstall()
	if err != nil {
		return nil, err
	}

	return &pgdbe, nil
}

func (p *PGDBE) CheckInstall() error {
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

func (p *PGDBE) IsConfigured() bool {
	return false
}

func (p *PGDBE) Configure() error {
	return errors.New("postgres engine: Configure() unimplemented")
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

func (p *PGDBE) Patch(ptch *patch.Patch) error {

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
